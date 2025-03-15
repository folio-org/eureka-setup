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
	internal.PortStartIndex = viper.GetInt(internal.ApplicationPortStart)
	environment := internal.GetEnvironmentFromConfig(deployModulesCommand, internal.EnvironmentKey)
	sidecarEnvironment := internal.GetEnvironmentFromConfig(deployModulesCommand, internal.SidecarModuleEnvironmentKey)

	slog.Info(deployModulesCommand, internal.GetFuncName(), "### READING BACKEND MODULES FROM CONFIG ###")
	backendModulesMap := internal.GetBackendModulesFromConfig(deployModulesCommand, viper.GetStringMap(internal.BackendModuleKey), false)

	slog.Info(deployModulesCommand, internal.GetFuncName(), "### READING FRONTEND MODULES FROM CONFIG ###")
	frontendModulesMap := internal.GetFrontendModulesFromConfig(deployModulesCommand, viper.GetStringMap(internal.FrontendModuleKey), viper.GetStringMap(internal.CustomFrontendModuleKey))

	slog.Info(deployModulesCommand, internal.GetFuncName(), "### READING BACKEND MODULE REGISTRIES ###")
	instalJsonUrls := map[string]string{internal.FolioRegistry: viper.GetString(internal.RegistryFolioInstallJsonUrlKey), internal.EurekaRegistry: viper.GetString(internal.RegistryEurekaInstallJsonUrlKey)}
	registryModules := internal.GetModulesFromRegistries(deployModulesCommand, instalJsonUrls)

	slog.Info(deployModulesCommand, internal.GetFuncName(), "### EXTRACTING MODULE NAME AND VERSION ###")
	internal.ExtractModuleNameAndVersion(deployModulesCommand, enableDebug, registryModules)

	slog.Info(deployModulesCommand, internal.GetFuncName(), "### ACQUIRING VAULT ROOT TOKEN ###")
	client := internal.CreateClient(deployModulesCommand)
	defer client.Close()
	vaultRootToken := internal.GetRootVaultToken(deployModulesCommand, client)

	slog.Info(deployModulesCommand, internal.GetFuncName(), "### CREATING APPLICATIONS ###")
	registryUrls := map[string]string{internal.FolioRegistry: registryUrl, "eureka": registryUrl}
	registerModuleDto := internal.NewRegisterModuleDto(registryUrls, registryModules, backendModulesMap, frontendModulesMap, enableDebug)
	internal.CreateApplications(deployModulesCommand, enableDebug, registerModuleDto)

	slog.Info(deployModulesCommand, internal.GetFuncName(), "### PULLING SIDECAR IMAGE ###")
	deployModulesDto := internal.NewDeployModulesDto(vaultRootToken, map[string]string{internal.FolioRegistry: "", "eureka": ""}, registryModules, backendModulesMap, environment, sidecarEnvironment)
	sidecarImage := internal.GetSidecarImage(deployManagementCommand, deployModulesDto.RegistryModules["eureka"])
	internal.PullModule(deployManagementCommand, client, sidecarImage)

	slog.Info(deployModulesCommand, internal.GetFuncName(), "### DEPLOYING MODULES ###")
	deployedModules := internal.DeployModules(deployModulesCommand, client, deployModulesDto, sidecarImage)
	time.Sleep(5 * time.Second)

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
