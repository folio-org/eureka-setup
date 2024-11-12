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
	"fmt"
	"log/slog"

	"github.com/docker/docker/api/types/filters"
	"github.com/folio-org/eureka-cli/internal"
	"github.com/spf13/cobra"
)

const undeployUiCommand string = "Undeploy UI"

// undeployUiCmd represents the undeployUi command
var undeployUiCmd = &cobra.Command{
	Use:   "undeployUi",
	Short: "Undeploy UI",
	Long:  `Undeploy the UI containers.`,
	Run: func(cmd *cobra.Command, args []string) {
		UndeployUi()
	},
}

func UndeployUi() {
	slog.Info(undeployUiCommand, internal.GetFuncName(), "### UNDEPLOYING UI CONTAINERS ###")
	client := internal.CreateClient(undeployUiCommand)
	defer client.Close()

	for _, value := range internal.GetTenants(undeployUiCommand, enableDebug, false) {
		mapEntry := value.(map[string]interface{})
		tenant := mapEntry["name"].(string)

		filters := filters.NewArgs(filters.KeyValuePair{Key: "name", Value: fmt.Sprintf(internal.SingleUiContainerPattern, tenant)})
		deployedModules := internal.GetDeployedModules(undeployUiCommand, client, filters)
		for _, deployedModule := range deployedModules {
			internal.UndeployModule(undeployUiCommand, client, deployedModule)
		}
	}
}

func init() {
	rootCmd.AddCommand(undeployUiCmd)
}
