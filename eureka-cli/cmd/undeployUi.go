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

// undeployUiCmd represents the undeployUi command
var undeployUiCmd = &cobra.Command{
	Use:   "undeployUi",
	Short: "Undeploy UI",
	Long:  `Undeploy the UI containers.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := New(action.UndeployUi)
		if err != nil {
			return err
		}

		err = r.UndeployUi()
		if err != nil {
			return err
		}

		return nil
	},
}

func (r *Run) UndeployUi() error {
	slog.Info(r.Config.Action.Name, "text", "UNDEPLOYING UI CONTAINERS")
	client, err := r.Config.DockerClient.Create()
	if err != nil {
		return err
	}
	defer func() {
		_ = client.Close()
	}()

	foundTenants, _ := r.Config.ManagementStep.GetTenants(constant.NoneConsortium, constant.All)

	for _, value := range foundTenants {
		r.Config.ModuleStep.UndeployModuleByNamePattern(client, fmt.Sprintf(constant.SingleUiContainerPattern, value.(map[string]any)["name"].(string)), false)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(undeployUiCmd)
}
