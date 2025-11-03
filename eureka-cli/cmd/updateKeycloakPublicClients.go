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
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := New(action.UpdateKeycloakPublicClients)
		if err != nil {
			return err
		}

		return r.UpdateKeycloakPublicClients()
	},
}

func (r *Run) UpdateKeycloakPublicClients() error {
	// TODO Abstract
	slog.Info(r.RunConfig.Action.Name, "text", "UPDATING KEYCLOAK PUBLIC CLIENTS")
	keycloakMasterAccessToken, err := r.RunConfig.KeycloakSvc.GetKeycloakMasterAccessToken()
	if err != nil {
		return err
	}
	r.RunConfig.Action.KeycloakMasterAccessToken = keycloakMasterAccessToken

	tenants, err := r.RunConfig.ManagementSvc.GetTenants(constant.NoneConsortium, constant.All)
	if err != nil {
		return err
	}

	for _, value := range tenants {
		mapEntry := value.(map[string]any)
		configTenant := mapEntry["name"].(string)
		hasTenant := helpers.HasTenant(configTenant, r.RunConfig.Action.ConfigTenants)
		isUIEnabled := helpers.IsUIEnabled(configTenant, r.RunConfig.Action.ConfigTenants)
		if !hasTenant || !isUIEnabled {
			continue
		}

		err = r.RunConfig.TenantSvc.SetConfigTenantParams(configTenant)
		if err != nil {
			return err
		}

		slog.Info(r.RunConfig.Action.Name, "text", "Updating keycloak public client")
		err = r.RunConfig.KeycloakSvc.UpdateKeycloakPublicClientParams(configTenant, actionParams.PlatformCompleteURL)
		if err != nil {
			return err
		}
	}

	return nil
}

func init() {
	rootCmd.AddCommand(updateKeycloakPublicClientsCmd)
}
