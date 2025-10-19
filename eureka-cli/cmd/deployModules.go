/*
Copyright © 2025 Open Library Foundation

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

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/field"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/folio-org/eureka-cli/models"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// deployModulesCmd represents the deployModules command
var deployModulesCmd = &cobra.Command{
	Use:   "deployModules",
	Short: "Deploy modules",
	Long:  `Deploy multiple module versions.`,
	Run: func(cmd *cobra.Command, args []string) {
		startPort := viper.GetInt(field.ApplicationPortStart)
		endPort := viper.GetInt(field.ApplicationPortEnd)
		NewCustomRun(action.DeployModules, startPort, endPort).DeployModules()
	},
}

func (r *Run) DeployModules() {
	registryURL := viper.GetString(field.RegistryURL)
	environment := helpers.GetConfigEnvVars(field.Environment)
	sidecarEnvironment := helpers.GetConfigEnvVars(field.SidecarModuleEnvironment)

	slog.Info(r.Config.Action.Name, "text", "READING BACKEND MODULES FROM CONFIG")
	backendModulesMap := r.Config.ModuleParams.GetBackendModulesFromConfig(false, true, viper.GetStringMap(field.BackendModules))

	slog.Info(r.Config.Action.Name, "text", "READING FRONTEND MODULES FROM CONFIG")
	frontendModulesMap := r.Config.ModuleParams.GetFrontendModulesFromConfig(true, viper.GetStringMap(field.FrontendModules), viper.GetStringMap(field.CustomFrontendModules))

	slog.Info(r.Config.Action.Name, "text", "READING BACKEND MODULE REGISTRIES")
	instalJsonURLs := map[string]string{
		constant.FolioRegistry:  viper.GetString(field.InstallFolio),
		constant.EurekaRegistry: viper.GetString(field.InstallEureka),
	}
	registryModules := r.Config.RegistryStep.GetModules(instalJsonURLs, true)

	slog.Info(r.Config.Action.Name, "text", "EXTRACTING MODULE NAME AND VERSION")
	r.Config.RegistryStep.ExtractModuleNameAndVersion(registryModules, true)

	vaultRootToken, client := r.GetVaultRootTokenWithDockerClient()
	defer func() {
		_ = client.Close()
	}()

	slog.Info(r.Config.Action.Name, "text", "CREATING APPLICATIONS")
	registryURLs := map[string]string{constant.FolioRegistry: registryURL, constant.EurekaRegistry: registryURL}
	registerModuleExtract := models.NewRegistryModuleExtract(registryURLs, registryModules, backendModulesMap, frontendModulesMap)
	r.Config.ManagementStep.CreateApplications(registerModuleExtract)

	slog.Info(r.Config.Action.Name, "text", "PULLING SIDECAR IMAGE")
	registryHosts := map[string]string{constant.FolioRegistry: "", constant.EurekaRegistry: ""}
	containers := models.NewCoreAndBusinessContainers(vaultRootToken, registryHosts, registryModules, backendModulesMap, environment, sidecarEnvironment)
	sidecarImage, pullSidecarImage := r.Config.ModuleStep.GetSidecarImage(containers.RegistryModules[constant.EurekaRegistry])
	slog.Info(r.Config.Action.Name, "text", fmt.Sprintf("Using sidecar image %s", sidecarImage))
	sidecarResources := helpers.CreateResources(false, viper.GetStringMap(field.SidecarModuleResources))
	if pullSidecarImage {
		r.Config.ModuleStep.PullModule(client, sidecarImage)
	}

	slog.Info(r.Config.Action.Name, "text", "DEPLOYING MODULES")
	deployedModules := r.Config.ModuleStep.DeployModules(client, containers, sidecarImage, sidecarResources)
	time.Sleep(5 * time.Second)

	slog.Info(r.Config.Action.Name, "text", "WAITING FOR MODULES TO INITIALIZE")
	var waitMutex sync.WaitGroup
	waitMutex.Add(len(deployedModules))
	for deployedModule := range deployedModules {
		go r.Config.ModuleStep.PerformModuleHealthCheck(&waitMutex, deployedModule, deployedModules[deployedModule])
	}
	waitMutex.Wait()
	slog.Info(r.Config.Action.Name, "text", "All modules have initialized")
}

func init() {
	rootCmd.AddCommand(deployModulesCmd)
}
