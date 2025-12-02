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
	"os"

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/errors"
	"github.com/folio-org/eureka-cli/field"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// updateModuleDiscoveryCmd represents the redirect command
var updateModuleDiscoveryCmd = &cobra.Command{
	Use:   "updateModuleDiscovery",
	Short: "Update module discovery",
	Long:  `Update module discovery to point to a different sidecar URL.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		run, err := New(action.UpdateModuleDiscovery)
		if err != nil {
			return err
		}

		return run.UpdateModuleDiscovery()
	},
}

func (run *Run) UpdateModuleDiscovery() error {
	if err := run.setKeycloakMasterAccessTokenIntoContext(constant.ClientCredentials); err != nil {
		return err
	}
	if err := run.setModuleDiscoveryDataIntoContext(); err != nil {
		return err
	}

	slog.Info(run.Config.Action.Name, "text", "UPDATING MODULE DISCOVERY", "module", params.ModuleName, "id", params.ID)
	return run.Config.ManagementSvc.UpdateModuleDiscovery(params.ID, params.Restore, params.PrivatePort, params.SidecarURL)
}

func (run *Run) setModuleDiscoveryDataIntoContext() error {
	moduleDiscovery, err := run.Config.ManagementSvc.GetModuleDiscovery(params.ModuleName)
	if err != nil {
		return err
	}
	if len(moduleDiscovery.Discovery) == 0 {
		return errors.ModuleDiscoveryNotFound(params.ModuleName)
	}
	params.ID = moduleDiscovery.Discovery[0].ID

	return nil
}

func init() {
	rootCmd.AddCommand(updateModuleDiscoveryCmd)
	updateModuleDiscoveryCmd.PersistentFlags().StringVarP(&params.ModuleName, action.ModuleName.Long, action.ModuleName.Short, "", action.ModuleName.Description)
	if err := updateModuleDiscoveryCmd.RegisterFlagCompletionFunc(action.ModuleName.Long, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return helpers.GetBackendModuleNames(viper.GetStringMap(field.BackendModules)), cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		slog.Error(errors.RegisterFlagCompletionFailed(err).Error())
		os.Exit(1)
	}
	updateModuleDiscoveryCmd.PersistentFlags().StringVarP(&params.SidecarURL, action.SidecarURL.Long, action.SidecarURL.Short, "", action.SidecarURL.Description)
	updateModuleDiscoveryCmd.PersistentFlags().IntVarP(&params.PrivatePort, action.PrivatePort.Long, action.PrivatePort.Short, 8081, action.PrivatePort.Description)
	updateModuleDiscoveryCmd.PersistentFlags().BoolVarP(&params.Restore, action.Restore.Long, action.Restore.Short, false, action.Restore.Description)
	if err := updateModuleDiscoveryCmd.MarkPersistentFlagRequired(action.ModuleName.Long); err != nil {
		slog.Error(errors.MarkFlagRequiredFailed(action.ModuleName, err).Error())
		os.Exit(1)
	}
	if err := updateModuleDiscoveryCmd.MarkPersistentFlagRequired(action.SidecarURL.Long); err != nil {
		slog.Error(errors.MarkFlagRequiredFailed(action.SidecarURL, err).Error())
		os.Exit(1)
	}
	if err := updateModuleDiscoveryCmd.MarkPersistentFlagRequired(action.PrivatePort.Long); err != nil {
		slog.Error(errors.MarkFlagRequiredFailed(action.PrivatePort, err).Error())
		os.Exit(1)
	}
}
