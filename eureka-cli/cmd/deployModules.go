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
	"encoding/json"
	"log/slog"
	"os"
	"path"
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
	home, err := os.UserHomeDir()
	cobra.CheckErr(err)

	registryUrl := viper.GetString(internal.RegistryUrlKey)
	registryFolioInstallJsonUrl := viper.GetString(internal.RegistryFolioInstallJsonUrlKey)
	registryEurekaInstallJsonUrl := viper.GetString(internal.RegistryEurekaInstallJsonUrlKey)
	environmentMap := viper.GetStringMapString(internal.EnvironmentKey)
	fileModuleEnv := path.Join(home, internal.WorkDir, viper.GetString(internal.FilesModuleEnvKey))
	fileModuleDescriptors := path.Join(home, internal.WorkDir, viper.GetString(internal.FilesModuleDescriptorsKey))
	backendModulesAnyMap := viper.GetStringMap(internal.BackendModuleKey)
	frontendModulesAnyMap := viper.GetStringMap(internal.FrontendModuleKey)

	slog.Info(deployModulesCommand, "### READING ENVIRONMENT FROM CONFIG ###", "")
	environment := internal.GetEnvironmentFromConfig(deployModulesCommand, environmentMap)

	slog.Info(deployModulesCommand, "### READING BACKEND MODULES FROM CONFIG ###", "")
	backendModulesMap := internal.GetBackendModulesFromConfig(deployModulesCommand, backendModulesAnyMap)

	slog.Info(deployModulesCommand, "### READING FRONTEND MODULES FROM CONFIG ###", "")
	frontendModulesMap := internal.GetFrontendModulesFromConfig(deployModulesCommand, frontendModulesAnyMap)

	slog.Info(deployModulesCommand, "### READING BACKEND MODULE REGISTRIES ###", "")
	instalJsonUrls := map[string]string{"folio": registryFolioInstallJsonUrl, "eureka": registryEurekaInstallJsonUrl}
	registryModules := internal.GetModulesFromRegistries(deployModulesCommand, instalJsonUrls)

	slog.Info(deployModulesCommand, "### EXTRACTING MODULE NAME AND VERSION ###", "")
	internal.ExtractModuleNameAndVersion(deployModulesCommand, enableDebug, registryModules)

	slog.Info(deployModulesCommand, "### ACQUIRING VAULT TOKEN ###", "")
	client := internal.CreateClient(deployModulesCommand)
	defer client.Close()
	vaultToken := internal.GetVaultToken(deployModulesCommand, client)

	slog.Info(deployModulesCommand, "### CREATING MODULE ENV FILE ###", "")
	fileModuleEnvPointer := internal.CreateModuleEnvFile(deployModulesCommand, fileModuleEnv)
	defer fileModuleEnvPointer.Close()

	slog.Info(deployModulesCommand, "### CREATING APPLICATIONS ###", "")
	moduleDescriptorsMap := make(map[string]interface{})
	registryUrls := map[string]string{"folio": registryUrl, "eureka": registryUrl}
	registerModuleDto := internal.NewRegisterModuleDto(registryUrls, registryModules, backendModulesMap, frontendModulesMap, moduleDescriptorsMap, fileModuleEnvPointer, enableDebug)
	internal.CreateApplications(deployModulesCommand, enableDebug, registerModuleDto)

	slog.Info(deployModulesCommand, "Created module environment file", "")
	fileModuleDescriptorsPointer := internal.CreateModuleDescriptorsFile(deployModulesCommand, fileModuleDescriptors)
	defer fileModuleDescriptorsPointer.Close()
	encoder := json.NewEncoder(fileModuleDescriptorsPointer)
	encoder.Encode(moduleDescriptorsMap)
	slog.Info(deployModulesCommand, "Created module descriptors file", "")

	slog.Info(deployModulesCommand, "### DEPLOYING MODULES ###", "")
	registryHostname := map[string]string{"folio": "", "eureka": ""}
	deployModulesDto := internal.NewDeployModulesDto(vaultToken, registryHostname, registryModules, backendModulesMap, environment)
	internal.DeployModules(deployModulesCommand, client, deployModulesDto)

	slog.Info(deployModulesCommand, "### WAITING FOR MODULES TO INITIALIZE ###", "")
	time.Sleep(180 * time.Second)
}

func init() {
	rootCmd.AddCommand(deployModulesCmd)
}
