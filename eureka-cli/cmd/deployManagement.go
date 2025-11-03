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
	"github.com/folio-org/eureka-cli/models"
	"github.com/spf13/cobra"
)

// deployManagementCmd represents the deployManagement command
var deployManagementCmd = &cobra.Command{
	Use:   "deployManagement",
	Short: "Deploy management",
	Long:  `Deploy all management modules.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := New(action.DeployManagement)
		if err != nil {
			return err
		}

		return r.DeployManagement()
	},
}

func (r *Run) DeployManagement() error {
	slog.Info(r.RunConfig.Action.Name, "text", "READING BACKEND MODULES FROM CONFIG")
	backendModules, err := r.RunConfig.ModuleParams.ReadBackendModulesFromConfig(true)
	if err != nil {
		return err
	}

	slog.Info(r.RunConfig.Action.Name, "text", "READING BACKEND MODULE REGISTRIES")
	instalJsonURLs := map[string]string{
		constant.EurekaRegistry: r.RunConfig.Action.ConfigEurekaRegistry,
	}
	registryModules, err := r.RunConfig.RegistrySvc.GetModules(instalJsonURLs, true)
	if err != nil {
		return err
	}
	r.RunConfig.RegistrySvc.ExtractModuleNameAndVersion(registryModules)

	client, err := r.RunConfig.DockerClient.Create()
	if err != nil {
		return err
	}
	defer r.RunConfig.DockerClient.Close(client)

	err = r.setVaultRootTokenIntoContext(client)
	if err != nil {
		return err
	}

	slog.Info(r.RunConfig.Action.Name, "text", "DEPLOYING MANAGEMENT MODULES")
	registryHosts := map[string]string{
		constant.EurekaRegistry: "",
	}
	env := action.GetConfigEnvVars(field.Env)
	containers := models.NewManagementContainers(r.RunConfig.Action.VaultRootToken, registryHosts, registryModules, backendModules, env)
	deployedModules, err := r.RunConfig.ModuleSvc.DeployModules(client, containers, "", nil)
	if err != nil {
		return err
	}
	time.Sleep(constant.DeployManagementWait)

	slog.Info(r.RunConfig.Action.Name, "text", "WAITING FOR MANAGEMENT MODULES TO BECOME READY")
	var deployManagementWG sync.WaitGroup
	errCh := make(chan error, len(deployedModules))
	deployManagementWG.Add(len(deployedModules))
	for deployedModule := range deployedModules {
		go r.RunConfig.ModuleSvc.CheckModuleReadiness(&deployManagementWG, errCh, deployedModule, deployedModules[deployedModule])
	}
	deployManagementWG.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			return err
		}
	}

	slog.Info(r.RunConfig.Action.Name, "text", "WAITING FOR KONG ROUTES TO BECOME READY")
	if err := r.RunConfig.KongSvc.CheckRouteReadiness(); err != nil {
		return err
	}
	slog.Info(r.RunConfig.Action.Name, "text", "All management modules are ready")

	return nil
}

func init() {
	rootCmd.AddCommand(deployManagementCmd)
}
