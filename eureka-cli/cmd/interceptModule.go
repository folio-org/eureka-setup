/*
Copyright © 2024 EPAM_Systems/Thunderjet/Boburbek_Kadirkhodjaev

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/folio-org/eureka-cli/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const interceptModuleCommand = "Intercept Module"

type InterceptModuleDto struct {
	id         string
	moduleName string
	moduleUrl  *string
	sidecarUrl *string

	portStart int
	portEnd   int

	deployModulesDto *internal.DeployModulesDto
	networkConfig    *network.NetworkingConfig
	backendModule    *internal.BackendModule
	registryModule   *internal.RegistryModule
}

func NewInterceptModuleDto(id, moduleUrl, sidecarUrl string, portStart, portEnd int) *InterceptModuleDto {
	id = strings.ReplaceAll(id, ":", "-")
	return &InterceptModuleDto{
		id:         id,
		moduleName: internal.TrimModuleName(internal.ModuleIdRegexp.ReplaceAllString(id, `$1`)),
		moduleUrl:  &moduleUrl,
		sidecarUrl: &sidecarUrl,
		portStart:  portStart,
		portEnd:    portEnd,
	}
}

func (dto *InterceptModuleDto) ClearUrls() {
	dto.moduleUrl = nil
	dto.sidecarUrl = nil
}

// interceptModuleCmd represents the interceptModule command
var interceptModuleCmd = &cobra.Command{
	Use:   "interceptModule",
	Short: "Intercept module",
	Long:  `Intercept/redirect module traffic to IntelliJ.`,
	Run: func(cmd *cobra.Command, args []string) {
		InterceptModule()
	},
}

func InterceptModule() {
	dto := NewInterceptModuleDto(id, moduleUrl, sidecarUrl, viper.GetInt(internal.ApplicationPortStart), viper.GetInt(internal.ApplicationPortEnd))

	slog.Info(interceptModuleCommand, internal.GetFuncName(), fmt.Sprintf("### INTERCEPTING %s MODULE ###", dto.moduleName))
	internal.PortStartIndex = viper.GetInt(internal.ApplicationPortStart)
	internal.PortEndIndex = viper.GetInt(internal.ApplicationPortEnd)
	globalEnvironment := internal.GetEnvironmentFromConfig(interceptModuleCommand, internal.EnvironmentKey)
	globalSidecarEnvironment := internal.GetEnvironmentFromConfig(deployModulesCommand, internal.SidecarModuleEnvironmentKey)

	slog.Info(interceptModuleCommand, internal.GetFuncName(), "### READING BACKEND MODULES FROM CONFIG ###")
	backendModulesMap := internal.GetBackendModulesFromConfig(interceptModuleCommand, viper.GetStringMap(internal.BackendModuleKey), false)

	slog.Info(interceptModuleCommand, internal.GetFuncName(), "### READING BACKEND MODULE REGISTRIES ###")
	instalJsonUrls := map[string]string{internal.FolioRegistry: viper.GetString(internal.RegistryFolioInstallJsonUrlKey), internal.EurekaRegistry: viper.GetString(internal.RegistryEurekaInstallJsonUrlKey)}
	registryModules := internal.GetModulesFromRegistries(interceptModuleCommand, instalJsonUrls)

	slog.Info(interceptModuleCommand, internal.GetFuncName(), "### EXTRACTING MODULE NAME AND VERSION ###")
	internal.ExtractModuleNameAndVersion(interceptModuleCommand, enableDebug, registryModules)

	vaultRootToken, client := GetVaultRootTokenWithDockerClient()
	defer client.Close()

	slog.Info(interceptModuleCommand, internal.GetFuncName(), "### UNDEPLOYING DEFAULT MODULE AND SIDECAR PAIR ###")
	internal.UndeployModuleByNamePattern(interceptModuleCommand, client, fmt.Sprintf(internal.SingleModuleContainerPattern, viper.GetString(internal.ProfileNameKey), dto.moduleName), false)
	dto.deployModulesDto = internal.NewDeployModulesDto(vaultRootToken, map[string]string{internal.FolioRegistry: "", internal.EurekaRegistry: ""}, registryModules, backendModulesMap, globalEnvironment, globalSidecarEnvironment)

	UpdateModuleDiscovery()
	if restore {
		deployDefaultModuleAndSidecar(dto, client)
		return
	}
	deployCustomSidecarForInterception(dto, client)
}

func deployDefaultModuleAndSidecar(dto *InterceptModuleDto, client *client.Client) {
	slog.Info(interceptModuleCommand, internal.GetFuncName(), "### DEPLOYING DEFAULT MODULE AND SIDECAR PAIR ###")
	dto.ClearUrls()

	prepareContainerNetwork(dto, true)
	deployModule(dto, client)
	deploySidecar(dto, client)

	slog.Info(interceptModuleCommand, internal.GetFuncName(), "### WAITING FOR MODULE TO INITIALIZE ###")
	var waitMutex sync.WaitGroup
	waitMutex.Add(1)
	go internal.PerformModuleHealthcheck(interceptModuleCommand, enableDebug, &waitMutex, dto.moduleName, dto.backendModule.ModuleExposedServerPort)
	waitMutex.Wait()
}

func deployCustomSidecarForInterception(dto *InterceptModuleDto, client *client.Client) {
	slog.Info(interceptModuleCommand, internal.GetFuncName(), "### DEPLOYING CUSTOM SIDECAR FOR INTERCEPTION ###")
	prepareContainerNetwork(dto, false)
	deploySidecar(dto, client)
}

func prepareContainerNetwork(dto *InterceptModuleDto, moduleAndSidecar bool) {
	dto.networkConfig = internal.NewModuleNetworkConfig()

	if moduleAndSidecar {
		moduleServerPort := internal.GetFreePortFromRange(interceptModuleCommand, dto.portStart, dto.portEnd, []int{})
		moduleDebugPort := internal.GetFreePortFromRange(interceptModuleCommand, dto.portStart, dto.portEnd, []int{moduleServerPort})
		sidecarServerPort := internal.GetFreePortFromRange(interceptModuleCommand, dto.portStart, dto.portEnd, []int{moduleServerPort, moduleDebugPort})
		sidecarDebugPort := internal.GetFreePortFromRange(interceptModuleCommand, dto.portStart, dto.portEnd, []int{moduleServerPort, moduleDebugPort, sidecarServerPort})

		dto.backendModule, dto.registryModule = internal.GetBackendModule(interceptModuleCommand, dto.deployModulesDto, dto.moduleName)
		dto.backendModule.ModulePortBindings = internal.CreatePortBindings(moduleServerPort, moduleDebugPort, dto.backendModule.ModuleServerPort)
		dto.backendModule.SidecarPortBindings = internal.CreatePortBindings(sidecarServerPort, sidecarDebugPort, dto.backendModule.ModuleServerPort)
		dto.backendModule.ModuleExposedServerPort = moduleServerPort
		return
	}

	sidecarServer, err := strconv.Atoi(strings.TrimSpace(regexp.MustCompile(internal.ColonDelimitedPattern).ReplaceAllString(sidecarUrl, `$1`)))
	if err != nil {
		slog.Error(interceptModuleCommand, internal.GetFuncName(), "strconv.Atoi error")
		panic(err)
	}
	sidecarDebugPort := internal.GetFreePortFromRange(interceptModuleCommand, dto.portStart, dto.portEnd, []int{sidecarServer})

	dto.backendModule, dto.registryModule = internal.GetBackendModule(interceptModuleCommand, dto.deployModulesDto, dto.moduleName)
	dto.backendModule.SidecarPortBindings = internal.CreatePortBindings(sidecarServer, sidecarDebugPort, dto.backendModule.ModuleServerPort)
}

func deployModule(dto *InterceptModuleDto, client *client.Client) {
	moduleVersion := internal.GetModuleImageVersion(*dto.backendModule, dto.registryModule)
	moduleImage := internal.GetModuleImage(interceptModuleCommand, moduleVersion, dto.registryModule)
	moduleEnvironment := internal.GetModuleEnvironment(dto.deployModulesDto, *dto.backendModule)
	moduleDeployDto := internal.NewDeployModuleDto(dto.registryModule.Name, moduleImage, moduleEnvironment, *dto.backendModule, dto.networkConfig)
	internal.DeployModule(interceptModuleCommand, client, moduleDeployDto)
}

func deploySidecar(dto *InterceptModuleDto, client *client.Client) {
	sidecarImage := internal.GetSidecarImage(interceptModuleCommand, dto.deployModulesDto.RegistryModules[internal.EurekaRegistry])
	sidecarResources := internal.CreateResources(false, viper.GetStringMap(internal.SidecarModuleResourcesKey))
	sidecarEnvironment := internal.GetSidecarEnvironment(dto.deployModulesDto, dto.registryModule, *dto.backendModule, dto.moduleUrl, dto.sidecarUrl)
	sidecarDeployDto := internal.NewDeploySidecarDto(dto.registryModule.SidecarName, sidecarImage, sidecarEnvironment, *dto.backendModule, dto.networkConfig, sidecarResources)
	internal.DeployModule(interceptModuleCommand, client, sidecarDeployDto)
}

func init() {
	rootCmd.AddCommand(interceptModuleCmd)
	interceptModuleCmd.PersistentFlags().StringVarP(&id, "id", "i", "", "Module id, e.g. mod-orders:13.1.0-SNAPSHOT.1021 (required)")
	interceptModuleCmd.PersistentFlags().StringVarP(&moduleUrl, "moduleUrl", "m", "", "Module URL, e.g. http://host.docker.internal:36002")
	interceptModuleCmd.PersistentFlags().StringVarP(&sidecarUrl, "sidecarUrl", "s", "", "Sidecar URL e.g. http://host.docker.internal:37002")
	interceptModuleCmd.PersistentFlags().BoolVarP(&restore, "restore", "r", false, "Restore module & sidecar")
	interceptModuleCmd.MarkPersistentFlagRequired("id")
}
