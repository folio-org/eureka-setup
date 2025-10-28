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

	"github.com/folio-org/eureka-cli/internal"
	"github.com/spf13/cobra"
)

const (
	undeployUiCommand string = "Undeploy UI"

	singleUiContainerPattern string = "eureka-platform-complete-ui-%s"
)

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
	client := internal.CreateDockerClient(undeployUiCommand)
	defer func() {
		_ = client.Close()
	}()

	for _, value := range internal.GetTenants(undeployUiCommand, withEnableDebug, false, internal.NoneConsortium, internal.AllTenantTypes) {
		internal.UndeployModuleByNamePattern(undeployModuleCommand, client, fmt.Sprintf(singleUiContainerPattern, value.(map[string]any)["name"].(string)))
	}
}

func init() {
	rootCmd.AddCommand(undeployUiCmd)
}
