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
	"github.com/folio-org/eureka-cli/field"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

		err = r.UndeployModules()
		if err != nil {
			return err
		}

		return nil
	},
}

func (r *Run) UndeployModules() error {
	slog.Info(r.Config.Action.Name, "text", "REMOVING APPLICATIONS")
	applicationId := fmt.Sprintf("%s-%s", viper.GetString(field.ApplicationName), viper.GetString(field.ApplicationVersion))
	r.Config.ManagementStep.RemoveApplication(applicationId)

	slog.Info(r.Config.Action.Name, "text", "UNDEPLOYING MODULES")
	client, err := r.Config.DockerClient.Create()
	if err != nil {
		return err
	}
	defer func() {
		_ = client.Close()
	}()

	r.Config.ModuleStep.UndeployModuleByNamePattern(client, fmt.Sprintf(constant.ProfileContainerPattern, viper.GetString(field.ProfileName)), true)

	return nil
}

func init() {
	rootCmd.AddCommand(undeployModulesCmd)
}
