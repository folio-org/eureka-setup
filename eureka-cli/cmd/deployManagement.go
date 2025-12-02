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
	"time"

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/errors"
	"github.com/folio-org/eureka-cli/models"
	"github.com/spf13/cobra"
)

// deployManagementCmd represents the deployManagement command
var deployManagementCmd = &cobra.Command{
	Use:   "deployManagement",
	Short: "Deploy management",
	Long:  `Deploy all management modules.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		run, err := New(action.DeployManagement)
		if err != nil {
			return err
		}

		return run.DeployManagement()
	},
}

func (run *Run) DeployManagement() error {
	slog.Info(run.Config.Action.Name, "text", "READING BACKEND MODULES")
	backendModules, err := run.Config.ModuleProps.ReadBackendModules(true, true)
	if err != nil {
		return err
	}

	slog.Info(run.Config.Action.Name, "text", "READING BACKEND MODULE REGISTRIES")
	instalJsonURLs := run.Config.Action.GetEurekaInstallJsonURLs()
	modules, err := run.Config.RegistrySvc.GetModules(instalJsonURLs, true, true)
	if err != nil {
		return err
	}
	run.Config.RegistrySvc.ExtractModuleMetadata(modules)

	client, err := run.Config.DockerClient.Create()
	if err != nil {
		return err
	}
	defer run.Config.DockerClient.Close(client)
	if err := run.setVaultRootTokenIntoContext(client); err != nil {
		return err
	}

	slog.Info(run.Config.Action.Name, "text", "DEPLOYING MANAGEMENT MODULES")
	deployedModules, err := run.Config.ModuleSvc.DeployModules(client, &models.Containers{
		Modules:        modules,
		BackendModules: backendModules,
		IsManagement:   true,
	}, "", nil)
	if err != nil {
		return err
	}
	if len(deployedModules) == 0 {
		return errors.ModulesNotDeployed(len(deployedModules))
	}
	time.Sleep(constant.DeployManagementWait)

	slog.Info(run.Config.Action.Name, "text", "WAITING FOR MANAGEMENT MODULES TO BECOME READY")
	if err := run.CheckDeployedModuleReadiness(constant.Management, deployedModules); err != nil {
		return err
	}

	slog.Info(run.Config.Action.Name, "text", "WAITING FOR KONG ROUTES TO BECOME READY")
	if err := run.Config.KongSvc.CheckRouteReadiness(); err != nil {
		return err
	}

	slog.Info(run.Config.Action.Name, "text", "UPDATING REALM SETTINGS")
	if err := run.setKeycloakMasterAccessTokenIntoContext(constant.Password); err != nil {
		return err
	}
	if err := run.Config.KeycloakSvc.UpdateRealmAccessTokenSettings(constant.KeycloakMasterRealm, constant.KeycloakMasterRealmAccessTokenLifespan); err != nil {
		return err
	}
	slog.Info(run.Config.Action.Name, "text", "Realm settings have been updated", "realm", constant.KeycloakMasterRealm)

	return nil
}

func init() {
	rootCmd.AddCommand(deployManagementCmd)
	deployManagementCmd.PersistentFlags().BoolVarP(&params.SkipRegistry, action.SkipRegistry.Long, action.SkipRegistry.Short, false, action.SkipRegistry.Description)
}
