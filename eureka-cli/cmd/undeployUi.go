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

		return r.UndeployUi()
	},
}

func (r *Run) UndeployUi() error {
	// TODO Abstract
	slog.Info(r.RunConfig.Action.Name, "text", "UNDEPLOYING UI CONTAINERS")
	client, err := r.RunConfig.DockerClient.Create()
	if err != nil {
		return err
	}
	defer r.RunConfig.DockerClient.Close(client)

	tenants, err := r.RunConfig.ManagementSvc.GetTenants(constant.NoneConsortium, constant.All)
	if err != nil {
		return err
	}

	for _, value := range tenants {
		pattern := fmt.Sprintf(constant.SingleUiContainerPattern, value.(map[string]any)["name"].(string))
		err = r.RunConfig.ModuleSvc.UndeployModuleByNamePattern(client, pattern)
		if err != nil {
			return err
		}
	}

	return nil
}

func init() {
	rootCmd.AddCommand(undeployUiCmd)
}
