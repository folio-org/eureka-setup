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

const deployManagementCommand string = "Deploy Management"

// deployManagementCmd represents the deployManagement command
var deployManagementCmd = &cobra.Command{
	Use:   "deployManagement",
	Short: "Deploy mananagement",
	Long:  `Deploy all management modules.`,
	Run: func(cmd *cobra.Command, args []string) {
		DeployManagement()
	},
}

func DeployManagement() {
	internal.PortStartIndex = viper.GetInt(internal.ApplicationPortStartKey)
	internal.PortEndIndex = viper.GetInt(internal.ApplicationPortEndKey)
	internal.ReservedPorts = []int{}
	environment := internal.GetEnvironmentFromConfig(deployManagementCommand, internal.EnvironmentKey)

	slog.Info(deployManagementCommand, internal.GetFuncName(), "### READING BACKEND MODULES FROM CONFIG ###")
	backendModulesMap := internal.GetBackendModulesFromConfig(deployManagementCommand, true, true, viper.GetStringMap(internal.BackendModulesKey))

	slog.Info(deployManagementCommand, internal.GetFuncName(), "### READING BACKEND MODULE REGISTRIES ###")
	registryModules := internal.GetModulesFromRegistries(deployManagementCommand, map[string]string{internal.EurekaRegistry: viper.GetString(internal.InstallEurekaKey)}, true)

	slog.Info(deployManagementCommand, internal.GetFuncName(), "### EXTRACTING MODULE NAME AND VERSION ###")
	internal.ExtractModuleNameAndVersion(deployManagementCommand, withEnableDebug, registryModules, true)

	vaultRootToken, client := GetVaultRootTokenWithDockerClient()
	defer client.Close()

	slog.Info(deployManagementCommand, internal.GetFuncName(), "### DEPLOYING MANAGEMENT MODULES ###")
	registryHostnames := map[string]string{internal.EurekaRegistry: ""}
	deployModulesDto := internal.NewDeployManagementModulesDto(vaultRootToken, registryHostnames, registryModules, backendModulesMap, environment)
	deployedModules := internal.DeployModules(deployManagementCommand, client, deployModulesDto, "", nil)
	time.Sleep(5 * time.Second)

	slog.Info(deployManagementCommand, internal.GetFuncName(), "### WAITING FOR MANAGEMENT MODULES TO INITIALIZE ###")
	var waitMutex sync.WaitGroup
	waitMutex.Add(len(deployedModules))
	for deployedModule := range deployedModules {
		go internal.PerformModuleHealthcheck(deployManagementCommand, withEnableDebug, &waitMutex, deployedModule, deployedModules[deployedModule])
	}
	waitMutex.Wait()
	time.Sleep(5 * time.Second)
	slog.Info(deployManagementCommand, internal.GetFuncName(), "All management modules have initialized")
}

func init() {
	rootCmd.AddCommand(deployManagementCmd)
}
