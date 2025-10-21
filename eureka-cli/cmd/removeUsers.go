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

// removeUsersCmd represents the removeUsers command
var removeUsersCmd = &cobra.Command{
	Use:   "removeUsers",
	Short: "Create users",
	Long:  `Create all users.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := New(action.RemoveUsers)
		if err != nil {
			return err
		}

		err = r.PartitionErr(func(consortiumName string, tenantType constant.TenantType) error {
			return r.RemoveUsers(consortiumName, tenantType)
		})
		if err != nil {
			return err
		}

		return nil
	},
}

func (r *Run) RemoveUsers(consortiumName string, tenantType constant.TenantType) error {
	vaultRootToken, err := r.GetVaultRootToken()
	if err != nil {
		return err
	}

	foundTenants, _ := r.Config.ManagementSvc.GetTenants(consortiumName, tenantType)

	for _, value := range foundTenants {
		mapEntry := value.(map[string]any)

		existingTenant := mapEntry["name"].(string)
		if !helpers.HasTenant(existingTenant) {
			continue
		}

		slog.Info(r.Config.Action.Name, "text", "REMOVING USERS FOR TENANT", "tenant", existingTenant)
		keycloakAccessToken, err := r.Config.KeycloakSvc.GetKeycloakAccessToken(vaultRootToken, existingTenant)
		if err != nil {
			return err
		}

		_ = r.Config.KeycloakSvc.RemoveUsers(existingTenant, keycloakAccessToken)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(removeUsersCmd)
}
