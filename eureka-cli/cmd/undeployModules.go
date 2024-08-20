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
	"log/slog"

	"github.com/docker/docker/api/types/filters"
	"github.com/folio-org/eureka-cli/internal"
	"github.com/spf13/cobra"
)

const undeployModulesCommand = "Undeploy Modules"

// undeployModulesCmd represents the undeployModules command
var undeployModulesCmd = &cobra.Command{
	Use:   "undeployModules",
	Short: "Undeploy modules",
	Long:  `Undeploy multiple modules.`,
	Run: func(cmd *cobra.Command, args []string) {
		slog.Info(undeployModulesCommand, internal.MessageKey, "### UNDEPLOYING MODULES ###")

		client := internal.CreateClient(undeployModulesCommand)
		defer client.Close()

		filters := filters.NewArgs(filters.KeyValuePair{Key: "name", Value: "eureka-mod"})

		deployedModules := internal.GetDeployedModules(undeployModulesCommand, client, filters)

		for _, deployedModule := range deployedModules {
			internal.UndeployModule(undeployModulesCommand, client, deployedModule)
		}

		slog.Info(undeployModulesCommand, internal.MessageKey, "### REMOVING APPLICATIONS ###")

		internal.RemoveApplications(undeployModulesCommand, "", enableDebug)

		slog.Info(undeployModulesCommand, internal.MessageKey, "### REMOVING TENANT ENTITLEMENTS ###")

		internal.RemoveTenantEntitlements(undeployModuleCommand, enableDebug)

		slog.Info(undeployModulesCommand, internal.MessageKey, "### REMOVING TENANTS ###")

		internal.RemoveTenants(undeployModuleCommand, enableDebug)
	},
}

func init() {
	rootCmd.AddCommand(undeployModulesCmd)
}
