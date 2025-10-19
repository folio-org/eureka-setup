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
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/folio-org/eureka-cli/tenanttype"
	"github.com/spf13/cobra"
)

// createRolesCmd represents the createRoles command
var createRolesCmd = &cobra.Command{
	Use:   "createRoles",
	Short: "Create roles",
	Long:  `Create all roles.`,
	Run: func(cmd *cobra.Command, args []string) {
		run := NewRun(action.CreateRoles)
		run.PartitionByConsortiumAndTenantType(func(consortiumName string, tenantType tenanttype.TenantType) {
			run.CreateRoles(consortiumName, tenantType)
		})
	},
}

func (r *Run) CreateRoles(consortiumName string, tenantType tenanttype.TenantType) {
	vaultRootToken := r.GetVaultRootToken()

	for _, value := range r.Config.ManagementStep.GetTenants(false, consortiumName, tenantType) {
		mapEntry := value.(map[string]any)

		existingTenant := mapEntry["name"].(string)
		if !helpers.HasTenant(existingTenant) {
			continue
		}

		slog.Info(r.Config.Action.Name, "text", fmt.Sprintf("CREATING ROLES FOR %s TENANT", existingTenant))
		keycloakAccessToken := r.Config.KeycloakStep.GetKeycloakAccessToken(vaultRootToken, existingTenant)
		r.Config.ManagementStep.CreateRoles(false, existingTenant, keycloakAccessToken)
	}
}

func init() {
	rootCmd.AddCommand(createRolesCmd)
}
