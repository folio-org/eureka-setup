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

	"github.com/docker/docker/api/types/filters"
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
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Registries
		folioRegistryUrl := viper.GetString(internal.FolioRegistryUrlKey)
		folioInstallJsonUrl := viper.GetString(internal.FolioInstallJsonUrlKey)
		folioRegistryHostname := viper.GetString(internal.FolioRegistryHostnameKey)
		eurekaInstallJsonUrl := viper.GetString(internal.EurekaInstallJsonUrlKey)
		eurekaRegistryHostname := viper.GetString(internal.EurekaRegistryHostnameKey)
		eurekaRegistryUrl := viper.GetString(internal.EurekaRegistryUrlKey)

		// Shared ENV
		sharedEnvMap := viper.GetStringMapString(internal.SharedEnvKey)

		// Cache files
		cacheFileModuleEnv := path.Join(home, internal.WorkDir, viper.GetString(internal.CacheFileModuleEnvKey))
		cacheFileModuleDescriptors := path.Join(home, internal.WorkDir, viper.GetString(internal.CacheFileModuleDescriptorsKey))

		// Backend modules
		backendModulesAnyMap := viper.GetStringMap(internal.BackendModuleKey)

		// UI modules
		// TODO Add support for UI modules

		slog.Info(deployModulesCommand, internal.PrimaryMessageKey, "### READING SHARED ENV FROM CONFIG ###")

		sharedEnv := internal.GetSharedEnvFromConfig(deployModulesCommand, sharedEnvMap)

		slog.Info(deployModulesCommand, internal.PrimaryMessageKey, "### READING BACKEND MODULES FROM CONFIG ###")

		backendModulesMap := internal.GetBackendModulesFromConfig(deployModulesCommand, backendModulesAnyMap)

		slog.Info(deployModulesCommand, internal.PrimaryMessageKey, "### READING BACKEND MODULES REGISTRIES ###")

		instalJsonUrls := map[string]string{
			"folio":  folioInstallJsonUrl,
			"eureka": eurekaInstallJsonUrl,
		}

		registryModules := internal.GetModulesFromRegistries(deployModulesCommand, instalJsonUrls)

		slog.Info(deployModulesCommand, internal.PrimaryMessageKey, "### ACQUIRING VAULT TOKEN ###")

		containerdCli := internal.CreateContainerdCli(deployModulesCommand)
		defer containerdCli.Close()

		vaultToken := internal.GetVaultToken(deployModulesCommand, containerdCli)

		slog.Info(deployModulesCommand, internal.PrimaryMessageKey, "### ACQUIRING EUREKA REGISTRY AUTH TOKEN ###")

		eurekaRegistryAuthToken := internal.GetEurekaRegistryAuthToken(deployModulesCommand)

		slog.Info(deployModulesCommand, internal.PrimaryMessageKey, "### CREATING MODULE ENV CACHE FILE ###")

		cacheFileModuleEnvPointer := internal.CreateModuleEnvCacheFile(deployModulesCommand, cacheFileModuleEnv)
		defer cacheFileModuleEnvPointer.Close()

		slog.Info(deployModulesCommand, internal.PrimaryMessageKey, "### DEREGISTERING MODULES ###")

		internal.DeregisterModules(deployModulesCommand, "", enableDebug)

		slog.Info(deployModulesCommand, internal.PrimaryMessageKey, "### REGISTERING MODULES ###")

		moduleDescriptorsMap := make(map[string]interface{})

		registerModuleDto := &internal.RegisterModuleDto{
			RegistryUrls: map[string]string{
				"folio":  folioRegistryUrl,
				"eureka": eurekaRegistryUrl,
			},
			RegistryModules:           registryModules,
			BackendModulesMap:         backendModulesMap,
			ModuleDescriptorsMap:      moduleDescriptorsMap,
			CacheFileModuleEnvPointer: cacheFileModuleEnvPointer,
			EnableDebug:               enableDebug,
		}

		internal.RegisterModules(deployModulesCommand, enableDebug, registerModuleDto)

		slog.Info(deployModulesCommand, internal.SecondaryMessageKey, "Created module ENV cache file")

		cacheFileModuleDescriptorsPointer := internal.CreateModuleDescriptorsCacheFile(deployModulesCommand, cacheFileModuleDescriptors)
		defer cacheFileModuleDescriptorsPointer.Close()

		encoder := json.NewEncoder(cacheFileModuleDescriptorsPointer)
		encoder.Encode(moduleDescriptorsMap)

		slog.Info(deployModulesCommand, internal.SecondaryMessageKey, "Created module descriptors cache file")

		slog.Info(deployModulesCommand, internal.PrimaryMessageKey, "### UNDEPLOYING MODULES ###")

		filters := filters.NewArgs(filters.KeyValuePair{Key: "name", Value: "eureka"})

		deployedModules := internal.GetDeployedModules(deployModulesCommand, containerdCli, filters)

		for _, deployedModule := range deployedModules {
			internal.UndeployModule(deployModulesCommand, containerdCli, deployedModule)
		}

		slog.Info(deployModulesCommand, internal.PrimaryMessageKey, "### DEPLOYING MODULES ###")

		deployModulesDto := &internal.DeployModulesDto{
			VaultToken: vaultToken,
			RegistryHostname: map[string]string{
				"folio":  folioRegistryHostname,
				"eureka": eurekaRegistryHostname,
			},
			EurekaRegistryAuthToken: eurekaRegistryAuthToken,
			RegistryModules:         registryModules,
			BackendModulesMap:       backendModulesMap,
			SharedEnv:               sharedEnv,
		}

		internal.DeployModules(deployModulesCommand, containerdCli, deployModulesDto)

		slog.Info(deployModulesCommand, internal.PrimaryMessageKey, "### CREATING TENANTS ###")

		internal.CreateTenants(deployModulesCommand, enableDebug)
	},
}

func init() {
	rootCmd.AddCommand(deployModulesCmd)
}
