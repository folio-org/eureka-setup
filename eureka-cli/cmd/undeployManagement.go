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

	"github.com/j011195/eureka-setup/eureka-cli/action"
	"github.com/j011195/eureka-setup/eureka-cli/constant"
	"github.com/spf13/cobra"
)

// undeployManagementCmd represents the undeployManagement command
var undeployManagementCmd = &cobra.Command{
	Use:   "undeployManagement",
	Short: "Undeploy management",
	Long:  `Undeploy all management modules.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		run, err := New(action.UndeployManagement)
		if err != nil {
			return err
		}

		return run.UndeployManagement()
	},
}

func (run *Run) UndeployManagement() error {
	slog.Info(run.Config.Action.Name, "text", "UNDEPLOYING MANAGEMENT MODULES")
	client, err := run.Config.DockerClient.Create()
	if err != nil {
		return err
	}
	defer run.Config.DockerClient.Close(client)

	return run.Config.ModuleSvc.UndeployModuleByNamePattern(client, constant.ManagementContainerPattern)
}

func init() {
	rootCmd.AddCommand(undeployManagementCmd)
}
