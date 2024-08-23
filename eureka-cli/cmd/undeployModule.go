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

	"github.com/docker/docker/api/types/filters"
	"github.com/folio-org/eureka-cli/internal"
	"github.com/spf13/cobra"
)

const undeployModuleCommand = "Undeploy Module"

// undeployModuleCmd represents the undeployModule command
var undeployModuleCmd = &cobra.Command{
	Use:   "undeployModule",
	Short: "Undeploy module",
	Long:  `Undeploy a single module.`,
	Run: func(cmd *cobra.Command, args []string) {
		UndeployModule()
	},
}

func UndeployModule() {
	slog.Info(undeployModulesCommand, internal.MessageKey, "### DEREGISTERING MODULE ###")
	internal.RemoveApplications(undeployModulesCommand, moduleName, enableDebug)

	slog.Info(undeployModuleCommand, internal.MessageKey, "### UNDEPLOYING MODULE ###")
	client := internal.CreateClient(undeployModuleCommand)
	defer client.Close()

	filters := filters.NewArgs(filters.KeyValuePair{Key: "name", Value: fmt.Sprintf("^(eureka-)(%[1]s|%[1]s-sc)$", moduleName)})
	deployedModules := internal.GetDeployedModules(undeployModuleCommand, client, filters)
	for _, deployedModule := range deployedModules {
		internal.UndeployModule(undeployModuleCommand, client, deployedModule)
	}
}

func init() {
	rootCmd.AddCommand(undeployModuleCmd)

	undeployModuleCmd.PersistentFlags().StringVarP(&moduleName, "moduleName", "m", "", "Module name (required)")
	undeployModuleCmd.MarkPersistentFlagRequired("moduleName")
}
