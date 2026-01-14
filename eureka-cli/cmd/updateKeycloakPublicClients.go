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

	"github.com/j011195/eureka-setup/eureka-cli/action"
	"github.com/j011195/eureka-setup/eureka-cli/constant"
	"github.com/j011195/eureka-setup/eureka-cli/helpers"
	"github.com/spf13/cobra"
)

// updateKeycloakPublicClientsCmd represents the updateKeycloakPublicClients command
var updateKeycloakPublicClientsCmd = &cobra.Command{
	Use:   "updateKeycloakPublicClients",
	Short: "Update Keycloak public client params",
	Long:  `Update Keycloak public client params for each UI container.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		run, err := New(action.UpdateKeycloakPublicClients)
		if err != nil {
			return err
		}

		return run.ConsortiumPartition(func(consortiumName string, tenantType constant.TenantType) error {
			return run.UpdateKeycloakPublicClients(consortiumName, tenantType)
		})
	},
}

func (run *Run) UpdateKeycloakPublicClients(consortiumName string, tenantType constant.TenantType) error {
	if err := run.setKeycloakMasterAccessTokenIntoContext(constant.Password); err != nil {
		return err
	}

	return run.TenantPartition(consortiumName, tenantType, func(configTenant, tenantType string) error {
		slog.Info(run.Config.Action.Name, "text", "UPDATING KEYCLOAK PUBLIC CLIENTS")
		if helpers.IsUIEnabled(configTenant, run.Config.Action.ConfigTenants) {
			slog.Info(run.Config.Action.Name, "text", "Setting config tenant params")
			if err := run.Config.TenantSvc.SetConfigTenantParams(configTenant); err != nil {
				return err
			}

			slog.Info(run.Config.Action.Name, "text", "Updating keycloak public client")
			if err := run.Config.KeycloakSvc.UpdatePublicClientSettings(configTenant, params.PlatformCompleteURL); err != nil {
				return err
			}
		}

		return nil
	})
}

func init() {
	rootCmd.AddCommand(updateKeycloakPublicClientsCmd)
}
