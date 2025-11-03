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

// createUsersCmd represents the createUsers command
var createUsersCmd = &cobra.Command{
	Use:   "createUsers",
	Short: "Create users",
	Long:  `Create all users.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := New(action.CreateUsers)
		if err != nil {
			return err
		}

		return r.ConsortiumPartitionErr(func(consortiumName string, tenantType constant.TenantType) error {
			return r.CreateUsers(consortiumName, tenantType)
		})
	},
}

func (r *Run) CreateUsers(consortiumName string, tenantType constant.TenantType) error {
	// TODO Abstract
	err := r.GetVaultRootToken()
	if err != nil {
		return err
	}

	tenants, err := r.RunConfig.ManagementSvc.GetTenants(consortiumName, tenantType)
	if err != nil {
		return err
	}

	for _, value := range tenants {
		mapEntry := value.(map[string]any)
		configTenant := mapEntry["name"].(string)
		hasTenant := helpers.HasTenant(configTenant, r.RunConfig.Action.ConfigTenants)
		if !hasTenant {
			continue
		}

		slog.Info(r.RunConfig.Action.Name, "text", "CREATING USERS FOR TENANT", "tenant", configTenant)
		keycloakAccessToken, err := r.RunConfig.KeycloakSvc.GetKeycloakAccessToken(configTenant)
		if err != nil {
			return err
		}
		r.RunConfig.Action.KeycloakAccessToken = keycloakAccessToken

		err = r.RunConfig.KeycloakSvc.CreateUsers(configTenant)
		if err != nil {
			return err
		}
	}

	return nil
}

func init() {
	rootCmd.AddCommand(createUsersCmd)
}
