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

const undeployManagementCommand = "Undeploy Management"

// undeployManagementCmd represents the undeployManagement command
var undeployManagementCmd = &cobra.Command{
	Use:   "undeployManagement",
	Short: "Undeploy modules",
	Long:  `Undeploy all management modules.`,
	Run: func(cmd *cobra.Command, args []string) {
		UndeployManagement()
	},
}

func UndeployManagement() {
	slog.Info(undeployManagementCommand, "### UNDEPLOYING MANAGEMENT MODULES ###", "")
	client := internal.CreateClient(undeployManagementCommand)
	defer client.Close()

	filters := filters.NewArgs(filters.KeyValuePair{Key: "name", Value: "eureka-mgr"})
	deployedModules := internal.GetDeployedModules(undeployManagementCommand, client, filters)
	for _, deployedModule := range deployedModules {
		internal.UndeployModule(undeployManagementCommand, client, deployedModule)
	}
}

func init() {
	rootCmd.AddCommand(undeployManagementCmd)
}
