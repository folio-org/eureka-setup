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
	"os/exec"

	"github.com/folio-org/eureka-cli/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	listModulesCommand string = "List Modules"

	module     string = "module"
	sidecar    string = "sidecar"
	management string = "management"
)

var availableModuleTypes []string = []string{module, sidecar, management}

// listModulesCmd represents the listModules command
var listModulesCmd = &cobra.Command{
	Use:   "listModules",
	Short: "List modules",
	Long:  `List all modules.`,
	Run: func(cmd *cobra.Command, args []string) {
		ListModules()
	},
}

func ListModules() {
	filter := fmt.Sprintf("name=%s", createFilter(withModuleName, withModuleType, withAll))
	internal.RunCommand(listModulesCommand, exec.Command("docker", "container", "ls", "--all", "--filter", filter))
}

func createFilter(moduleName string, moduleType string, all bool) string {
	if all {
		return internal.AllContainerPattern
	}

	currentProfile := viper.GetString(internal.ProfileNameKey)
	if moduleName != "" {
		return fmt.Sprintf(internal.SingleModuleOrSidecarContainerPattern, currentProfile, moduleName)
	}

	switch moduleType {
	case management:
		return fmt.Sprintf(internal.ManagementContainerPattern)
	case module:
		return fmt.Sprintf(internal.ModuleContainerPattern, currentProfile)
	case sidecar:
		return fmt.Sprintf(internal.SidecarContainerPattern, currentProfile)
	default:
		return fmt.Sprintf(internal.ProfileContainerPattern, currentProfile)
	}
}

func init() {
	rootCmd.AddCommand(listModulesCmd)
	listModulesCmd.Flags().BoolVarP(&withAll, "all", "a", false, "All modules for all profiles")
	listModulesCmd.Flags().StringVarP(&withModuleName, "moduleName", "m", "", "By module name, e.g. mod-orders")
	listModulesCmd.Flags().StringVarP(&withModuleType, "moduleType", "M", "", fmt.Sprintf("By module type, options: %s", availableModuleTypes))
	if err := listModulesCmd.RegisterFlagCompletionFunc("moduleType", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return availableModuleTypes, cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		slog.Error(listModulesCommand, internal.GetFuncName(), "listModulesCmd.RegisterFlagCompletionFunc error")
		panic(err)
	}
}
