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
	"log/slog"

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/spf13/cobra"
)

// updateKeycloakPublicClientsCmd represents the updateKeycloakPublicClients command
var updateKeycloakPublicClientsCmd = &cobra.Command{
	Use:   "updateKeycloakPublicClients",
	Short: "Update Keycloak public client params",
	Long:  `Update Keycloak public client params for each UI container.`,
	Run: func(cmd *cobra.Command, args []string) {
		NewRun(action.UpdateKeycloakPublicClients).UpdateKeycloakPublicClients()
	},
}

func (r *Run) UpdateKeycloakPublicClients() {
	slog.Info(r.Config.Action.Name, "text", "UPDATING KEYCLOAK PUBLIC CLIENTS")
	keycloakMasterAccessToken := r.Config.KeycloakStep.GetKeycloakMasterAccessToken()

	for _, value := range r.Config.ManagementStep.GetTenants(false, constant.NoneConsortium, constant.All) {
		mapEntry := value.(map[string]any)

		existingTenant := mapEntry["name"].(string)
		if !helpers.HasTenant(existingTenant) || !helpers.IsUIEnabled(existingTenant) {
			continue
		}

		r.Config.TenantStep.SetDefaultConfigTenantParams(&rp, existingTenant)

		slog.Info(r.Config.Action.Name, "text", "Updating keycloak public client")
		r.Config.KeycloakStep.UpdateKeycloakPublicClientParams(existingTenant, keycloakMasterAccessToken, rp.PlatformCompleteURL)
	}
}

func init() {
	rootCmd.AddCommand(updateKeycloakPublicClientsCmd)
}
