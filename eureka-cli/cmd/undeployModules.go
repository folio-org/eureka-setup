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

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/spf13/cobra"
)

// undeployModulesCmd represents the undeployModules command
var undeployModulesCmd = &cobra.Command{
	Use:   "undeployModules",
	Short: "Undeploy modules",
	Long:  `Undeploy multiple modules.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		run, err := New(action.UndeployModules)
		if err != nil {
			return err
		}

		return run.UndeployModules(true)
	},
}

func (run *Run) UndeployModules(deleteApplication bool) error {
	if deleteApplication {
		slog.Info(run.Config.Action.Name, "text", "REMOVING APPLICATION")
		_ = run.Config.ManagementSvc.RemoveApplication(run.Config.Action.ConfigApplicationID)
	}

	slog.Info(run.Config.Action.Name, "text", "UNDEPLOYING MODULES")
	client, err := run.Config.DockerClient.Create()
	if err != nil {
		return err
	}
	defer run.Config.DockerClient.Close(client)

	pattern := fmt.Sprintf(constant.ProfileContainerPattern, run.Config.Action.ConfigProfile)
	return run.Config.ModuleSvc.UndeployModuleByNamePattern(client, pattern)
}

func init() {
	rootCmd.AddCommand(undeployModulesCmd)
}
