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
	"log/slog"
	"sync"

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
	registryFolioInstallJsonUrl := viper.GetString(internal.RegistryFolioInstallJsonUrlKey)
	registryEurekaInstallJsonUrl := viper.GetString(internal.RegistryEurekaInstallJsonUrlKey)
	backendModulesAnyMap := viper.GetStringMap(internal.BackendModuleKey)
	frontendModulesAnyMap := viper.GetStringMap(internal.FrontendModuleKey)
	internal.PortIndex = viper.GetInt(internal.ApplicationPortStart)
	environment := internal.GetEnvironmentFromConfig(deployModulesCommand, internal.EnvironmentKey)
	sidecarEnvironment := internal.GetEnvironmentFromConfig(deployModulesCommand, internal.SidecarModuleEnvironmentKey)

	slog.Info(deployModulesCommand, internal.GetFuncName(), "### READING BACKEND MODULES FROM CONFIG ###")
	backendModulesMap := internal.GetBackendModulesFromConfig(deployModulesCommand, backendModulesAnyMap, false)

	slog.Info(deployModulesCommand, internal.GetFuncName(), "### READING FRONTEND MODULES FROM CONFIG ###")
	frontendModulesMap := internal.GetFrontendModulesFromConfig(deployModulesCommand, frontendModulesAnyMap)

	slog.Info(deployModulesCommand, internal.GetFuncName(), "### READING BACKEND MODULE REGISTRIES ###")
	instalJsonUrls := map[string]string{"folio": registryFolioInstallJsonUrl, "eureka": registryEurekaInstallJsonUrl}
	registryModules := internal.GetModulesFromRegistries(deployModulesCommand, instalJsonUrls)

	slog.Info(deployModulesCommand, internal.GetFuncName(), "### EXTRACTING MODULE NAME AND VERSION ###")
	internal.ExtractModuleNameAndVersion(deployModulesCommand, enableDebug, registryModules)

	slog.Info(deployModulesCommand, internal.GetFuncName(), "### ACQUIRING VAULT ROOT TOKEN ###")
	client := internal.CreateClient(deployModulesCommand)
	defer client.Close()
	vaultRootToken := internal.GetRootVaultToken(deployModulesCommand, client)

	slog.Info(deployModulesCommand, internal.GetFuncName(), "### CREATING APPLICATIONS ###")
	moduleDescriptorsMap := make(map[string]interface{})
	registryUrls := map[string]string{"folio": registryUrl, "eureka": registryUrl}
	registerModuleDto := internal.NewRegisterModuleDto(registryUrls, registryModules, backendModulesMap, frontendModulesMap, moduleDescriptorsMap, enableDebug)
	internal.CreateApplications(deployModulesCommand, enableDebug, registerModuleDto)

	slog.Info(deployModulesCommand, internal.GetFuncName(), "### DEPLOYING MODULES ###")
	registryHostnames := map[string]string{"folio": "", "eureka": ""}
	deployModulesDto := internal.NewDeployModulesDto(vaultRootToken, registryHostnames, registryModules, backendModulesMap, environment, sidecarEnvironment)
	deployedModules := internal.DeployModules(deployModulesCommand, client, deployModulesDto)

	slog.Info(deployModulesCommand, internal.GetFuncName(), "### WAITING FOR MODULES TO INITIALIZE ###")
	var waitMutex sync.WaitGroup
	waitMutex.Add(len(deployedModules))
	for deployedModule := range deployedModules {
		go internal.PerformModuleHealthcheck(deployModulesCommand, enableDebug, &waitMutex, deployedModule, deployedModules[deployedModule])
	}
	waitMutex.Wait()
	slog.Info(deployModulesCommand, internal.GetFuncName(), "All modules have initialized")
}

func init() {
	rootCmd.AddCommand(deployModulesCmd)
}
