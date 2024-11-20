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
	"log/slog"
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

var showAll bool

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
	slog.Info(listModulesCommand, internal.GetFuncName(), "### LISTING MODULES ###")
	var filter string
	if showAll {
		filter = allProfilesModulesPattern
	} else {
		filter = fmt.Sprintf(currentProfileModulesPattern, viper.GetString(internal.ProfileNameKey))
	}
	if moduleName != "" {
		filter = fmt.Sprintf(internal.SingleModuleContainerPattern, viper.GetString(internal.ProfileNameKey), moduleName)
	}
	internal.RunCommand(listSystemCommand, exec.Command("docker", "container", "ls", "--all", "--filter", fmt.Sprintf("name=%s", filter)))
}

func init() {
	rootCmd.AddCommand(listModulesCmd)
	listModulesCmd.Flags().StringVarP(&moduleName, "moduleName", "m", "", "Module name only, e.g. mod-users")
	listModulesCmd.Flags().BoolVarP(&showAll, "showAll", "s", false, "Show all modules for all profiles")
	listModulesCmd.MarkPersistentFlagRequired("moduleName")
}
