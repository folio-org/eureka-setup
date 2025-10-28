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
	env := action.GetConfigEnvVars(field.Env)

	slog.Info(r.Config.Action.Name, "text", "READING BACKEND MODULES FROM CONFIG")
	backendModules, err := r.Config.ModuleParams.ReadBackendModulesFromConfig(true)
	if err != nil {
		return err
	}

	slog.Info(r.Config.Action.Name, "text", "READING BACKEND MODULE REGISTRIES")
	instalJsonURLs := map[string]string{
		constant.EurekaRegistry: r.Config.Action.ConfigEurekaRegistry,
	}
	registryModules, err := r.Config.RegistrySvc.GetModules(instalJsonURLs, true)
	if err != nil {
		return err
	}
	r.Config.RegistrySvc.ExtractModuleNameAndVersion(registryModules)

	client, err := r.Config.DockerClient.Create()
	if err != nil {
		return err
	}
	defer r.Config.DockerClient.Close(client)

	slog.Info(r.Config.Action.Name, "text", "DEPLOYING MANAGEMENT MODULES")
	registryHosts := map[string]string{
		constant.EurekaRegistry: "",
	}
	containers := models.NewManagementContainers(r.Config.Action.VaultRootToken, registryHosts, registryModules, backendModules, env)
	deployedModules, err := r.Config.ModuleSvc.DeployModules(client, containers, "", nil)
	if err != nil {
		return err
	}
	time.Sleep(constant.DeployManagementWait)

	slog.Info(r.Config.Action.Name, "text", "WAITING FOR MANAGEMENT MODULES TO BECOME READY")
	var deployManagementWG sync.WaitGroup
	errCh := make(chan error, len(deployedModules))

	deployManagementWG.Add(len(deployedModules))
	for deployedModule := range deployedModules {
		go r.Config.ModuleSvc.CheckModuleReadiness(&deployManagementWG, errCh, deployedModule, deployedModules[deployedModule])
	}
	deployManagementWG.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			return err
		}
	}

	slog.Info(r.Config.Action.Name, "text", "WAITING FOR KONG ROUTES TO BECOME READY")
	expressions := []string{
		`(http.path == "/applications" && http.method == "POST")`,
		`(http.path ~ "^/modules/([^/]+)/discovery$" && http.method == "POST")`,
		`(http.path == "/entitlements" && http.method == "POST")`,
	}

	for retryCount := range constant.KongRouteReadinessMaxRetries {
		routes, err := r.Config.KongSvc.FindRouteByExpressions(expressions)
		if err != nil {
			if retryCount == constant.KongRouteReadinessMaxRetries {
				return fmt.Errorf("cannot continue deployment with error: %v", err)
			} else {
				slog.Info(r.Config.Action.Name, "text", "Kong routes are unready", "retryCount", retryCount, "maxRetries", constant.KongRouteReadinessMaxRetries)
				time.Sleep(5 * time.Second)
			}
		}

		if len(routes) == len(expressions) {
			for _, route := range routes {
				slog.Info(r.Config.Action.Name, "text", "Kong routes are ready", "id", route.ID, "expression", route.Expression)
			}
			break
		}
	}

	slog.Info(r.Config.Action.Name, "text", "All management modules are ready")

	return nil
}

func init() {
	rootCmd.AddCommand(deployManagementCmd)
}
