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

	"github.com/j011195/eureka-setup/eureka-cli/action"
	"github.com/j011195/eureka-setup/eureka-cli/constant"
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

		return run.UndeployModules(params.RemoveApplication)
	},
}

func (run *Run) UndeployModules(removeApplication bool) error {
	if removeApplication {
		slog.Info(run.Config.Action.Name, "text", "REMOVING APPLICATION")
		if err := run.setKeycloakMasterAccessTokenIntoContext(constant.ClientCredentials); err != nil {
			return err
		}
		if err := run.Config.ManagementSvc.RemoveApplication(run.Config.Action.ConfigApplicationID); err != nil {
			slog.Warn(run.Config.Action.Name, "text", "Application removal was unsuccessful", "error", err)
		}
	}

	slog.Info(run.Config.Action.Name, "text", "UNDEPLOYING MODULES")
	client, err := run.Config.DockerClient.Create()
	if err != nil {
		return err
	}
	defer run.Config.DockerClient.Close(client)

	pattern := fmt.Sprintf(constant.ProfileContainerPattern, run.Config.Action.ConfigProfileName)
	return run.Config.ModuleSvc.UndeployModuleByNamePattern(client, pattern)
}

func init() {
	rootCmd.AddCommand(undeployModulesCmd)
	undeployModulesCmd.PersistentFlags().BoolVarP(&params.RemoveApplication, action.RemoveApplication.Long, action.RemoveApplication.Short, true, action.RemoveApplication.Description)
}
