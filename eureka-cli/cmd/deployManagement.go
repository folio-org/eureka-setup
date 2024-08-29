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
	registryFolioInstallJsonUrl := viper.GetString(internal.RegistryFolioInstallJsonUrlKey)
	registryEurekaInstallJsonUrl := viper.GetString(internal.RegistryEurekaInstallJsonUrlKey)
	backendModulesAnyMap := viper.GetStringMap(internal.BackendModuleKey)

	slog.Info(deployManagementCommand, "### READING ENVIRONMENT FROM CONFIG ###", "")
	environment := internal.GetEnvironmentFromConfig(deployManagementCommand)

	slog.Info(deployManagementCommand, "### READING BACKEND MODULES FROM CONFIG ###", "")
	backendModulesMap := internal.GetBackendModulesFromConfig(deployManagementCommand, backendModulesAnyMap)

	slog.Info(deployManagementCommand, "### READING BACKEND MODULE REGISTRIES ###", "")
	instalJsonUrls := map[string]string{"folio": registryFolioInstallJsonUrl, "eureka": registryEurekaInstallJsonUrl}
	registryModules := internal.GetModulesFromRegistries(deployManagementCommand, instalJsonUrls)

	slog.Info(deployManagementCommand, "### EXTRACTING MODULE NAME AND VERSION ###", "")
	internal.ExtractModuleNameAndVersion(deployManagementCommand, enableDebug, registryModules)

	slog.Info(deployManagementCommand, "### ACQUIRING VAULT TOKEN ###", "")
	client := internal.CreateClient(deployManagementCommand)
	defer client.Close()
	vaultToken := internal.GetVaultToken(deployManagementCommand, client)

	slog.Info(deployManagementCommand, "### DEPLOYING MANAGEMENT MODULES ###", "")
	registryHostname := map[string]string{"folio": "", "eureka": ""}
	deployModulesDto := internal.NewDeployManagementModulesDto(vaultToken, registryHostname, registryModules, backendModulesMap, environment)
	internal.DeployModules(deployManagementCommand, client, deployModulesDto)

	slog.Info(deployManagementCommand, "### WAITING FOR MANAGEMENT MODULES TO INITIALIZE ###", "")
	time.Sleep(150 * time.Second)
}

func init() {
	rootCmd.AddCommand(deployManagementCmd)
}
