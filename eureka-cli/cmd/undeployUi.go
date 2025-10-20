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
	Run: func(cmd *cobra.Command, args []string) {
		NewRun(action.UndeployUi).UndeployUi()
	},
}

func (r *Run) UndeployUi() {
	slog.Info(r.Config.Action.Name, "text", "UNDEPLOYING UI CONTAINERS")
	client := r.Config.DockerClient.Create()
	defer func() {
		_ = client.Close()
	}()

	for _, value := range r.Config.ManagementStep.GetTenants(false, constant.NoneConsortium, constant.All) {
		r.Config.ModuleStep.UndeployModuleByNamePattern(client, fmt.Sprintf(constant.SingleUiContainerPattern, value.(map[string]any)["name"].(string)), false)
	}
}

func init() {
	rootCmd.AddCommand(undeployUiCmd)
}
