/*
Copyright © 2024 EPAM_Systems/Thunderjet/Boburbek_Kadirkhodjaev

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
	"os/exec"

	"github.com/folio-org/eureka-cli/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	listModulesCommand string = "List Modules"

	allProfilesModulesPattern    string = "eureka-"
	currentProfileModulesPattern string = "eureka-%s-"
)

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
	internal.RunCommand(listModulesCommand, exec.Command("docker", "container", "ls", "--all", "--filter", fmt.Sprintf("name=%s", createFilter())))
}

func createFilter() string {
	if !showAll {
		return fmt.Sprintf(currentProfileModulesPattern, viper.GetString(internal.ProfileNameKey))
	}
	if moduleName != "" {
		return fmt.Sprintf(internal.SingleModuleContainerPattern, viper.GetString(internal.ProfileNameKey), moduleName)
	}
	return allProfilesModulesPattern
}

func init() {
	rootCmd.AddCommand(listModulesCmd)
	listModulesCmd.Flags().StringVarP(&moduleName, "moduleName", "m", "", "Module name, e.g. mod-users")
	listModulesCmd.Flags().BoolVarP(&showAll, "all", "a", false, "Show all modules for all profiles")
	listModulesCmd.MarkPersistentFlagRequired("moduleName")
}
