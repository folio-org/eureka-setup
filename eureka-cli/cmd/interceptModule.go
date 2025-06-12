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

	sidecarServerPort int

	portStart int
	portEnd   int

	deployModulesDto *internal.DeployModulesDto
	networkConfig    *network.NetworkingConfig
	backendModule    *internal.BackendModule
	registryModule   *internal.RegistryModule
}

func NewInterceptModuleDto(id string, defaultGateway bool, moduleUrl, sidecarUrl string, portStart, portEnd int) *InterceptModuleDto {
	id = strings.ReplaceAll(id, ":", "-")

	var moduleUrlTemp, sidecarUrlTemp string
	if defaultGateway {
		schemaAndUrl := internal.GetGatewaySchemaAndUrl(interceptModuleCommand)
		moduleUrlTemp = fmt.Sprintf("%s:%s", schemaAndUrl, moduleUrl)
		sidecarUrlTemp = fmt.Sprintf("%s:%s", schemaAndUrl, sidecarUrl)
	} else {
		moduleUrlTemp = moduleUrl
		sidecarUrlTemp = sidecarUrl
	}

	return &InterceptModuleDto{
		id:         id,
		moduleName: internal.TrimModuleName(internal.ModuleIdRegexp.ReplaceAllString(id, `$1`)),
		moduleUrl:  &moduleUrlTemp,
		sidecarUrl: &sidecarUrlTemp,
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
	dto := NewInterceptModuleDto(withId, withDefaultGateway, withModuleUrl, withSidecarUrl, viper.GetInt(internal.ApplicationPortStartKey), viper.GetInt(internal.ApplicationPortEndKey))

	slog.Info(interceptModuleCommand, internal.GetFuncName(), fmt.Sprintf("### INTERCEPTING %s MODULE ###", dto.moduleName))
	internal.PortStartIndex = viper.GetInt(internal.ApplicationPortStartKey)
	internal.PortEndIndex = viper.GetInt(internal.ApplicationPortEndKey)
	globalEnvironment := internal.GetEnvironmentFromConfig(interceptModuleCommand, internal.EnvironmentKey)
	globalSidecarEnvironment := internal.GetEnvironmentFromConfig(deployModulesCommand, internal.SidecarModuleEnvironmentKey)
	backendModulesMap := internal.GetBackendModulesFromConfig(interceptModuleCommand, false, false, viper.GetStringMap(internal.BackendModulesKey))
	instalJsonUrls := map[string]string{internal.FolioRegistry: viper.GetString(internal.InstallFolioKey), internal.EurekaRegistry: viper.GetString(internal.InstallEurekaKey)}
	registryModules := internal.GetModulesFromRegistries(interceptModuleCommand, instalJsonUrls, false)
	internal.ExtractModuleNameAndVersion(interceptModuleCommand, withEnableDebug, registryModules, false)

	vaultRootToken, client := GetVaultRootTokenWithDockerClient()
	defer client.Close()

	slog.Info(interceptModuleCommand, internal.GetFuncName(), "### UNDEPLOYING DEFAULT MODULE AND SIDECAR PAIR ###")
	internal.UndeployModuleByNamePattern(interceptModuleCommand, client, fmt.Sprintf(internal.SingleModuleOrSidecarContainerPattern, viper.GetString(internal.ProfileNameKey), dto.moduleName), false)

	registryHostnames := map[string]string{internal.FolioRegistry: "", internal.EurekaRegistry: ""}
	dto.deployModulesDto = internal.NewDeployModulesDto(vaultRootToken, registryHostnames, registryModules, backendModulesMap, globalEnvironment, globalSidecarEnvironment)

	UpdateModuleDiscovery(*dto.sidecarUrl)
	if withRestore {
		deployDefaultModuleAndSidecar(dto, client)
		return
	}
	deployCustomSidecarForInterception(!withRestore, dto, client)
}

func deployDefaultModuleAndSidecar(dto *InterceptModuleDto, client *client.Client) {
	slog.Info(interceptModuleCommand, internal.GetFuncName(), "### DEPLOYING DEFAULT MODULE AND SIDECAR PAIR ###")
	dto.ClearUrls()

	prepareContainerNetwork(dto, true)
	deployModule(dto, client)
	deploySidecar(false, dto, client)

	slog.Info(interceptModuleCommand, internal.GetFuncName(), "### WAITING FOR MODULE TO INITIALIZE ###")
	var waitMutex sync.WaitGroup
	waitMutex.Add(1)
	go internal.PerformModuleHealthcheck(interceptModuleCommand, withEnableDebug, &waitMutex, dto.moduleName, dto.backendModule.ModuleExposedServerPort)
	waitMutex.Wait()
}

func deployCustomSidecarForInterception(printModuleEnvironment bool, dto *InterceptModuleDto, client *client.Client) {
	slog.Info(interceptModuleCommand, internal.GetFuncName(), "### DEPLOYING CUSTOM SIDECAR FOR INTERCEPTION ###")
	prepareContainerNetwork(dto, false)
	deploySidecar(printModuleEnvironment, dto, client)
}

func prepareContainerNetwork(dto *InterceptModuleDto, moduleAndSidecar bool) {
	dto.networkConfig = internal.NewModuleNetworkConfig()

	if moduleAndSidecar {
		moduleServerPort := internal.GetAndSetFreePortFromRange(interceptModuleCommand, dto.portStart, dto.portEnd, &internal.ReservedPorts)
		moduleDebugPort := internal.GetAndSetFreePortFromRange(interceptModuleCommand, dto.portStart, dto.portEnd, &internal.ReservedPorts)
		sidecarServerPort := internal.GetAndSetFreePortFromRange(interceptModuleCommand, dto.portStart, dto.portEnd, &internal.ReservedPorts)
		sidecarDebugPort := internal.GetAndSetFreePortFromRange(interceptModuleCommand, dto.portStart, dto.portEnd, &internal.ReservedPorts)

		dto.backendModule, dto.registryModule = internal.GetBackendModule(interceptModuleCommand, dto.deployModulesDto, dto.moduleName)
		dto.backendModule.ModulePortBindings = internal.CreatePortBindings(moduleServerPort, moduleDebugPort, dto.backendModule.ModuleServerPort)
		dto.backendModule.SidecarPortBindings = internal.CreatePortBindings(sidecarServerPort, sidecarDebugPort, dto.backendModule.ModuleServerPort)
		dto.backendModule.ModuleExposedServerPort = moduleServerPort
		return
	}

	sidecarServerPort := internal.ExtractPortFromUrl(interceptModuleCommand, *dto.sidecarUrl)
	sidecarDebugPort := internal.GetAndSetFreePortFromRange(interceptModuleCommand, dto.portStart, dto.portEnd, &internal.ReservedPorts)

	dto.sidecarServerPort = sidecarServerPort

	dto.backendModule, dto.registryModule = internal.GetBackendModule(interceptModuleCommand, dto.deployModulesDto, dto.moduleName)
	dto.backendModule.SidecarPortBindings = internal.CreatePortBindings(sidecarServerPort, sidecarDebugPort, dto.backendModule.ModuleServerPort)
}

func deployModule(dto *InterceptModuleDto, client *client.Client) {
	moduleVersion := internal.GetModuleImageVersion(*dto.backendModule, dto.registryModule)
	moduleImage := internal.GetModuleImage(interceptModuleCommand, moduleVersion, dto.registryModule)
	moduleEnvironment := internal.GetModuleEnvironment(dto.deployModulesDto, dto.registryModule, *dto.backendModule)
	moduleDeployDto := internal.NewDeployModuleDto(dto.registryModule.Name, moduleImage, moduleEnvironment, *dto.backendModule, dto.networkConfig)
	internal.DeployModule(interceptModuleCommand, client, moduleDeployDto)
}

func deploySidecar(printModuleEnvironment bool, dto *InterceptModuleDto, client *client.Client) {
	sidecarImage, pullSidecarImage := internal.GetSidecarImage(interceptModuleCommand, dto.deployModulesDto.RegistryModules[internal.EurekaRegistry])
	sidecarResources := internal.CreateResources(false, viper.GetStringMap(internal.SidecarModuleResourcesKey))
	sidecarEnvironment := internal.GetSidecarEnvironment(dto.deployModulesDto, dto.registryModule, *dto.backendModule, dto.moduleUrl, dto.sidecarUrl)
	sidecarDeployDto := internal.NewDeploySidecarDto(dto.registryModule.SidecarName, sidecarImage, sidecarEnvironment, *dto.backendModule, dto.networkConfig, sidecarResources)

	sidecarDeployDto.PullImage = pullSidecarImage

	internal.DeployModule(interceptModuleCommand, client, sidecarDeployDto)

	if printModuleEnvironment {
		moduleEnvironment := internal.GetModuleEnvironment(dto.deployModulesDto, dto.registryModule, *dto.backendModule)

		if dto.backendModule.UseOkapiUrl {
			moduleOkapiEnvironment := []string{"OKAPI_HOST=localhost",
				fmt.Sprintf("OKAPI_PORT=%d", dto.sidecarServerPort),
				"OKAPI_SERVICE_HOST=localhost",
				fmt.Sprintf("OKAPI_SERVICE_PORT=%d", dto.sidecarServerPort),
				fmt.Sprintf("OKAPI_SERVICE_URL=http://localhost:%d", dto.sidecarServerPort),
				fmt.Sprintf("OKAPI_URL=http://localhost:%d", dto.sidecarServerPort),
			}

			moduleEnvironment = append(moduleEnvironment, moduleOkapiEnvironment...)
		}

		fmt.Println()
		fmt.Printf("### %s ###\n", "Can be embedded into IntelliJ Run/Debug Configuration")
		for _, value := range moduleEnvironment {
			fmt.Println(value)
		}
		fmt.Println()
	}
}

func init() {
	rootCmd.AddCommand(interceptModuleCmd)
	interceptModuleCmd.PersistentFlags().StringVarP(&withId, "id", "i", "", "Module id, e.g. mod-orders:13.1.0-SNAPSHOT.1021 (required)")
	interceptModuleCmd.PersistentFlags().StringVarP(&withModuleUrl, "moduleUrl", "m", "", "Module URL, e.g. http://host.docker.internal:36002 or 36002 (if -g is used)")
	interceptModuleCmd.PersistentFlags().StringVarP(&withSidecarUrl, "sidecarUrl", "s", "", "Sidecar URL e.g. http://host.docker.internal:37002 or 37002 (if -g is used)")
	interceptModuleCmd.PersistentFlags().BoolVarP(&withRestore, "restore", "r", false, "Restore module & sidecar")
	interceptModuleCmd.PersistentFlags().BoolVarP(&withDefaultGateway, "defaultGateway", "g", false, "Use default gateway in URLs, .e.g http://host.docker.internal:{{port}} will be set automatically")
	interceptModuleCmd.MarkPersistentFlagRequired("id")
}
