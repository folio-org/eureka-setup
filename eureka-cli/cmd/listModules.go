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
	"github.com/folio-org/eureka-cli/errors"
	"github.com/folio-org/eureka-cli/field"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	filter := fmt.Sprintf("name=%s", run.createFilter(params.ModuleName, params.ModuleType, params.All))
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
	case constant.Management:
		return fmt.Sprintf(constant.ManagementContainerPattern)
	case constant.Module:
		return fmt.Sprintf(constant.ModuleContainerPattern, currentProfile)
	case constant.Sidecar:
		return fmt.Sprintf(constant.SidecarContainerPattern, currentProfile)
	default:
		return fmt.Sprintf(constant.ProfileContainerPattern, currentProfile)
	}
}

func init() {
	rootCmd.AddCommand(listModulesCmd)
	listModulesCmd.Flags().BoolVarP(&params.All, action.All.Long, action.All.Short, false, action.All.Description)
	listModulesCmd.Flags().StringVarP(&params.ModuleName, action.ModuleName.Long, action.ModuleName.Short, "", action.ModuleName.Description)
	if err := listModulesCmd.RegisterFlagCompletionFunc(action.ModuleName.Long, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return helpers.GetBackendModuleNames(viper.GetStringMap(field.BackendModules)), cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		slog.Error(errors.RegisterFlagCompletionFailed(err).Error())
		os.Exit(1)
	}
	listModulesCmd.Flags().StringVarP(&params.ModuleType, action.ModuleType.Long, action.ModuleType.Short, "", action.ModuleType.Description)
	if err := listModulesCmd.RegisterFlagCompletionFunc(action.ModuleType.Long, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return constant.GetContainerTypes(), cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		slog.Error(errors.RegisterFlagCompletionFailed(err).Error())
		os.Exit(1)
	}
}
