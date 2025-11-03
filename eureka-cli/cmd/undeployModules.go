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
		r, err := New(action.UndeployModules)
		if err != nil {
			return err
		}

		return r.UndeployModules()
	},
}

func (r *Run) UndeployModules() error {
	slog.Info(r.RunConfig.Action.Name, "text", "REMOVING APPLICATIONS")
	_ = r.RunConfig.ManagementSvc.RemoveApplication(r.RunConfig.Action.ConfigApplicationID)

	slog.Info(r.RunConfig.Action.Name, "text", "UNDEPLOYING MODULES")
	client, err := r.RunConfig.DockerClient.Create()
	if err != nil {
		return err
	}
	defer r.RunConfig.DockerClient.Close(client)

	pattern := fmt.Sprintf(constant.ProfileContainerPattern, r.RunConfig.Action.ConfigProfile)
	err = r.RunConfig.ModuleSvc.UndeployModuleByNamePattern(client, pattern)
	if err != nil {
		return err
	}

	return nil
}

func init() {
	rootCmd.AddCommand(undeployModulesCmd)
}
