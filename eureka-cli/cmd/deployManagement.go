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
	RunE: func(cmd *cobra.Command, args []string) error {
		startPort := viper.GetInt(field.ApplicationPortStart)
		endPort := viper.GetInt(field.ApplicationPortEnd)
		r, err := NewCustom(action.DeployManagement, startPort, endPort)
		if err != nil {
			return err
		}

		err = r.DeployManagement()
		if err != nil {
			return err
		}

		return nil
	},
}

func (r *Run) DeployManagement() error {
	env := helpers.GetConfigEnvVars(field.Env)

	slog.Info(r.Config.Action.Name, "text", "READING BACKEND MODULES FROM CONFIG")
	backendModulesMap, err := r.Config.ModuleParams.GetBackendModulesFromConfig(true, true, viper.GetStringMap(field.BackendModules))
	if err != nil {
		return err
	}

	slog.Info(r.Config.Action.Name, "text", "READING BACKEND MODULE REGISTRIES")
	instalJsonURLs := map[string]string{constant.EurekaRegistry: viper.GetString(field.InstallEureka)}
	registryModules, err := r.Config.RegistrySvc.GetModules(instalJsonURLs, true)
	if err != nil {
		return err
	}

	slog.Info(r.Config.Action.Name, "text", "EXTRACTING MODULE NAME AND VERSION")
	r.Config.RegistrySvc.ExtractModuleNameAndVersion(registryModules, true)

	vaultRootToken, client, err := r.GetVaultRootTokenWithDockerClient()
	if err != nil {
		return err
	}
	defer r.Config.DockerClient.Close(client)

	slog.Info(r.Config.Action.Name, "text", "DEPLOYING MANAGEMENT MODULES")
	registryHosts := map[string]string{constant.EurekaRegistry: ""}
	containers := models.NewManagementContainers(vaultRootToken, registryHosts, registryModules, backendModulesMap, env)
	deployedModules, err := r.Config.ModuleSvc.DeployModules(client, containers, "", nil)
	if err != nil {
		return err
	}
	time.Sleep(constant.DeployManagementWait)

	slog.Info(r.Config.Action.Name, "text", "WAITING FOR MANAGEMENT MODULES TO BECOME READY")
	var wg sync.WaitGroup
	errCh := make(chan error, len(deployedModules))

	wg.Add(len(deployedModules))
	for deployedModule := range deployedModules {
		go r.Config.ModuleSvc.PerformModuleReadinessCheck(&wg, errCh, deployedModule, deployedModules[deployedModule])
	}
	wg.Wait()
	close(errCh)

	select {
	case err := <-errCh:
		return err
	default:
	}

	slog.Info(r.Config.Action.Name, "text", "All management modules are ready")

	return nil
}

func init() {
	rootCmd.AddCommand(deployManagementCmd)
}
