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

// deployManagementCmd represents the deployManagement command
var deployManagementCmd = &cobra.Command{
	Use:   "deployManagement",
	Short: "Deploy management",
	Long:  `Deploy all management modules.`,
	Run: func(cmd *cobra.Command, args []string) {
		startPort := viper.GetInt(field.ApplicationPortStart)
		endPort := viper.GetInt(field.ApplicationPortEnd)
		NewCustomRun(action.DeployManagement, startPort, endPort).DeployManagement()
	},
}

func (r *Run) DeployManagement() {
	env := helpers.GetConfigEnvVars(field.Env)

	slog.Info(r.Config.Action.Name, "text", "READING BACKEND MODULES FROM CONFIG")
	backendModulesMap := r.Config.ModuleParams.GetBackendModulesFromConfig(true, true, viper.GetStringMap(field.BackendModules))

	slog.Info(r.Config.Action.Name, "text", "READING BACKEND MODULE REGISTRIES")
	instalJsonURLs := map[string]string{constant.EurekaRegistry: viper.GetString(field.InstallEureka)}
	registryModules := r.Config.RegistryStep.GetModules(instalJsonURLs, true)

	slog.Info(r.Config.Action.Name, "text", "EXTRACTING MODULE NAME AND VERSION")
	r.Config.RegistryStep.ExtractModuleNameAndVersion(registryModules, true)

	vaultRootToken, client := r.GetVaultRootTokenWithDockerClient()
	defer func() {
		_ = client.Close()
	}()

	slog.Info(r.Config.Action.Name, "text", "DEPLOYING MANAGEMENT MODULES")
	registryHosts := map[string]string{constant.EurekaRegistry: ""}
	containers := models.NewManagementContainers(vaultRootToken, registryHosts, registryModules, backendModulesMap, env)
	deployedModules := r.Config.ModuleStep.DeployModules(client, containers, "", nil)
	time.Sleep(5 * time.Second)

	slog.Info(r.Config.Action.Name, "text", "WAITING FOR MANAGEMENT MODULES TO INITIALIZE")
	var waitMutex sync.WaitGroup
	waitMutex.Add(len(deployedModules))
	for deployedModule := range deployedModules {
		go r.Config.ModuleStep.PerformModuleHealthCheck(&waitMutex, deployedModule, deployedModules[deployedModule])
	}
	waitMutex.Wait()
	time.Sleep(5 * time.Second)
	slog.Info(r.Config.Action.Name, "text", "All management modules have initialized")
}

func init() {
	rootCmd.AddCommand(deployManagementCmd)
}
