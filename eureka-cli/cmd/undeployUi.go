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

	"github.com/folio-org/eureka-setup/eureka-cli/action"
	"github.com/folio-org/eureka-setup/eureka-cli/constant"
	"github.com/folio-org/eureka-setup/eureka-cli/helpers"
	"github.com/spf13/cobra"
)

// undeployUiCmd represents the undeployUi command
var undeployUiCmd = &cobra.Command{
	Use:   "undeployUi",
	Short: "Undeploy UI",
	Long:  `Undeploy the UI containers.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		run, err := New(action.UndeployUi)
		if err != nil {
			return err
		}

		return run.UndeployUI()
	},
}

func (run *Run) UndeployUI() error {
	slog.Info(run.Config.Action.Name, "text", "UNDEPLOYING UI CONTAINERS")
	client, err := run.Config.DockerClient.Create()
	if err != nil {
		return err
	}
	defer run.Config.DockerClient.Close(client)
	if err := run.setKeycloakMasterAccessTokenIntoContext(constant.ClientCredentials); err != nil {
		return err
	}

	tenants, err := run.Config.ManagementSvc.GetTenants(constant.NoneConsortium, constant.All)
	if err != nil {
		return err
	}

	for _, value := range tenants {
		entry := value.(map[string]any)
		pattern := fmt.Sprintf(constant.SingleUiContainerPattern, helpers.GetString(entry, "name"))
		if err := run.Config.ModuleSvc.UndeployModuleByNamePattern(client, pattern); err != nil {
			return err
		}
	}

	return nil
}

func init() {
	rootCmd.AddCommand(undeployUiCmd)
}
