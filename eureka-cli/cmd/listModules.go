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
	"os"
	"os/exec"

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/spf13/cobra"
)

// listModulesCmd represents the listModules command
var listModulesCmd = &cobra.Command{
	Use:   "listModules",
	Short: "List modules",
	Long:  `List all modules.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		run, err := New(action.ListModules)
		if err != nil {
			return err
		}

		return run.ListModules()
	},
}

func (run *Run) ListModules() error {
	filter := fmt.Sprintf("name=%s", run.createFilter(actionParams.ModuleName, actionParams.ModuleType, actionParams.All))
	return run.Config.ExecSvc.Exec(exec.Command("docker", "container", "ls", "--all", "--filter", filter))
}

func (run *Run) createFilter(moduleName string, moduleType string, all bool) string {
	if all {
		return constant.AllContainerPattern
	}

	currentProfile := run.Config.Action.ConfigProfile
	if moduleName != "" {
		return fmt.Sprintf(constant.SingleModuleOrSidecarContainerPattern, currentProfile, moduleName)
	}

	switch moduleType {
	case constant.ManagementType:
		return fmt.Sprintf(constant.ManagementContainerPattern)
	case constant.ModuleType:
		return fmt.Sprintf(constant.ModuleContainerPattern, currentProfile)
	case constant.SidecarType:
		return fmt.Sprintf(constant.SidecarContainerPattern, currentProfile)
	default:
		return fmt.Sprintf(constant.ProfileContainerPattern, currentProfile)
	}
}

func init() {
	rootCmd.AddCommand(listModulesCmd)
	listModulesCmd.Flags().BoolVarP(&actionParams.All, "all", "a", false, "All modules for all profiles")
	listModulesCmd.Flags().StringVarP(&actionParams.ModuleName, "moduleName", "n", "", "Module name, e.g. mod-orders")
	listModulesCmd.Flags().StringVarP(&actionParams.ModuleType, "moduleType", "y", "", "Module type, e.g. management")
	if err := listModulesCmd.RegisterFlagCompletionFunc("moduleType", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return constant.GetContainerTypes(), cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		slog.Error("failed to register flag completion function", "error", err)
		os.Exit(1)
	}
}
