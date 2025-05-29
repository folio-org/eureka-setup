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
	"sync"
	"time"

	"github.com/folio-org/eureka-cli/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const deployModulesCommand string = "Deploy Modules"

// deployModulesCmd represents the deployModules command
var deployModulesCmd = &cobra.Command{
	Use:   "deployModules",
	Short: "Deploy modules",
	Long:  `Deploy multiple module versions.`,
	Run: func(cmd *cobra.Command, args []string) {
		DeployModules()
	},
}

func DeployModules() {
	registryUrl := viper.GetString(internal.RegistryUrlKey)
	internal.PortStartIndex = viper.GetInt(internal.ApplicationPortStartKey)
	internal.PortEndIndex = viper.GetInt(internal.ApplicationPortEndKey)
	internal.ReservedPorts = []int{}
	environment := internal.GetEnvironmentFromConfig(deployModulesCommand, internal.EnvironmentKey)
	sidecarEnvironment := internal.GetEnvironmentFromConfig(deployModulesCommand, internal.SidecarModuleEnvironmentKey)

	slog.Info(deployModulesCommand, internal.GetFuncName(), "### READING BACKEND MODULES FROM CONFIG ###")
	backendModulesMap := internal.GetBackendModulesFromConfig(deployModulesCommand, false, true, viper.GetStringMap(internal.BackendModulesKey))

	slog.Info(deployModulesCommand, internal.GetFuncName(), "### READING FRONTEND MODULES FROM CONFIG ###")
	frontendModulesMap := internal.GetFrontendModulesFromConfig(deployModulesCommand, true, viper.GetStringMap(internal.FrontendModulesKey), viper.GetStringMap(internal.CustomFrontendModulesKey))

	slog.Info(deployModulesCommand, internal.GetFuncName(), "### READING BACKEND MODULE REGISTRIES ###")
	instalJsonUrls := map[string]string{internal.FolioRegistry: viper.GetString(internal.InstallFolioKey), internal.EurekaRegistry: viper.GetString(internal.InstallEurekaKey)}
	registryModules := internal.GetModulesFromRegistries(deployModulesCommand, instalJsonUrls, true)

	slog.Info(deployModulesCommand, internal.GetFuncName(), "### EXTRACTING MODULE NAME AND VERSION ###")
	internal.ExtractModuleNameAndVersion(deployModulesCommand, withEnableDebug, registryModules, true)

	vaultRootToken, client := GetVaultRootTokenWithDockerClient()
	defer client.Close()

	slog.Info(deployModulesCommand, internal.GetFuncName(), "### CREATING APPLICATIONS ###")
	registryUrls := map[string]string{internal.FolioRegistry: registryUrl, internal.EurekaRegistry: registryUrl}
	registerModuleDto := internal.NewRegisterModuleDto(registryUrls, registryModules, backendModulesMap, frontendModulesMap, withEnableDebug)
	internal.CreateApplications(deployModulesCommand, withEnableDebug, registerModuleDto)

	slog.Info(deployModulesCommand, internal.GetFuncName(), "### PULLING SIDECAR IMAGE ###")
	registryHostnames := map[string]string{internal.FolioRegistry: "", internal.EurekaRegistry: ""}
	deployModulesDto := internal.NewDeployModulesDto(vaultRootToken, registryHostnames, registryModules, backendModulesMap, environment, sidecarEnvironment)
	sidecarImage, pullSidecarImage := internal.GetSidecarImage(deployModulesCommand, deployModulesDto.RegistryModules[internal.EurekaRegistry])
	slog.Info(deployModulesCommand, internal.GetFuncName(), fmt.Sprintf("Using sidecar image %s", sidecarImage))
	sidecarResources := internal.CreateResources(false, viper.GetStringMap(internal.SidecarModuleResourcesKey))
	if pullSidecarImage {
		internal.PullModule(deployModulesCommand, client, sidecarImage)
	}

	slog.Info(deployModulesCommand, internal.GetFuncName(), "### DEPLOYING MODULES ###")
	deployedModules := internal.DeployModules(deployModulesCommand, client, deployModulesDto, sidecarImage, sidecarResources)
	time.Sleep(5 * time.Second)

	slog.Info(deployModulesCommand, internal.GetFuncName(), "### WAITING FOR MODULES TO INITIALIZE ###")
	var waitMutex sync.WaitGroup
	waitMutex.Add(len(deployedModules))
	for deployedModule := range deployedModules {
		go internal.PerformModuleHealthcheck(deployModulesCommand, withEnableDebug, &waitMutex, deployedModule, deployedModules[deployedModule])
	}
	waitMutex.Wait()
	slog.Info(deployModulesCommand, internal.GetFuncName(), "All modules have initialized")
}

func init() {
	rootCmd.AddCommand(deployModulesCmd)
}
