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
	if err := run.setModuleDiscoveryDataIntoContext(); err != nil {
		return err
	}

	slog.Info(run.Config.Action.Name, "text", "UPDATING MODULE DISCOVERY", "module", actionParams.ModuleName, "id", actionParams.ID)
	return run.Config.ManagementSvc.UpdateModuleDiscovery(actionParams.ID, actionParams.Restore, actionParams.PrivatePort, actionParams.SidecarURL)
}

func (run *Run) setModuleDiscoveryDataIntoContext() error {
	moduleDiscovery, err := run.Config.ManagementSvc.GetModuleDiscovery(actionParams.ModuleName)
	if err != nil {
		return err
	}
	if len(moduleDiscovery.Discovery) == 0 {
		return errors.ModuleDiscoveryNotFound(actionParams.ModuleName)
	}
	actionParams.ID = moduleDiscovery.Discovery[0].ID

	return nil
}

func init() {
	rootCmd.AddCommand(updateModuleDiscoveryCmd)
	updateModuleDiscoveryCmd.PersistentFlags().StringVarP(&actionParams.ModuleName, "moduleName", "n", "", "Module name, e.g. mod-orders")
	if err := updateModuleDiscoveryCmd.RegisterFlagCompletionFunc("moduleName", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return helpers.GetBackendModuleNames(viper.GetStringMap(field.BackendModules)), cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		slog.Error("failed to register flag completion function", "error", err)
		os.Exit(1)
	}
	updateModuleDiscoveryCmd.PersistentFlags().StringVarP(&actionParams.SidecarURL, "sidecarUrl", "s", "", "Sidecar URL e.g. http://host.docker.internal:37002")
	updateModuleDiscoveryCmd.PersistentFlags().IntVarP(&actionParams.PrivatePort, "privatePort", "", 8081, "Private port e.g. 8081")
	updateModuleDiscoveryCmd.PersistentFlags().BoolVarP(&actionParams.Restore, "restore", "r", false, "Restore sidecar URL")
	if err := updateModuleDiscoveryCmd.MarkPersistentFlagRequired("moduleName"); err != nil {
		slog.Error("failed to mark moduleName flag as required", "error", err)
		os.Exit(1)
	}
	if err := updateModuleDiscoveryCmd.MarkPersistentFlagRequired("sidecarUrl"); err != nil {
		slog.Error("failed to mark sidecarUrl flag as required", "error", err)
		os.Exit(1)
	}
	if err := updateModuleDiscoveryCmd.MarkPersistentFlagRequired("privatePort"); err != nil {
		slog.Error("failed to mark privatePort flag as required", "error", err)
		os.Exit(1)
	}
}
