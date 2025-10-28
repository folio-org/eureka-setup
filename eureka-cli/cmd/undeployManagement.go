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

	"github.com/folio-org/eureka-cli/internal"
	"github.com/spf13/cobra"
)

const undeployManagementCommand = "Undeploy Management"

// undeployManagementCmd represents the undeployManagement command
var undeployManagementCmd = &cobra.Command{
	Use:   "undeployManagement",
	Short: "Undeploy management",
	Long:  `Undeploy all management modules.`,
	Run: func(cmd *cobra.Command, args []string) {
		UndeployModules()
		UndeployManagement()
	},
}

func UndeployManagement() {
	slog.Info(undeployManagementCommand, internal.GetFuncName(), "### UNDEPLOYING MANAGEMENT MODULES ###")
	client := internal.CreateDockerClient(undeployManagementCommand)
	defer func() {
		_ = client.Close()
	}()

	internal.UndeployModuleByNamePattern(undeployModuleCommand, client, internal.ManagementContainerPattern)
}

func init() {
	rootCmd.AddCommand(undeployManagementCmd)
}
