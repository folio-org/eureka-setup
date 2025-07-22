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

const createUsersCommand string = "Create Users"

// createUsersCmd represents the createUsers command
var createUsersCmd = &cobra.Command{
	Use:   "createUsers",
	Short: "Create users",
	Long:  `Create all users.`,
	Run: func(cmd *cobra.Command, args []string) {
		RunByConsortiumAndTenantType(createUsersCommand, func(consortium string, tenantType internal.TenantType) {
			CreateUsers(consortium, tenantType)
		})
	},
}

func CreateUsers(consortium string, tenantType internal.TenantType) {
	vaultRootToken := GetVaultRootToken()

	for _, value := range internal.GetTenants(createUsersCommand, withEnableDebug, false, consortium, tenantType) {
		mapEntry := value.(map[string]any)

		existingTenant := mapEntry["name"].(string)
		if !internal.HasTenant(existingTenant) {
			continue
		}

		slog.Info(createUsersCommand, internal.GetFuncName(), fmt.Sprintf("### CREATING USERS FOR %s TENANT ###", existingTenant))
		keycloakAccessToken := internal.GetKeycloakAccessToken(createUsersCommand, withEnableDebug, vaultRootToken, existingTenant)
		internal.CreateUsers(createUsersCommand, withEnableDebug, false, existingTenant, keycloakAccessToken)
	}
}

func init() {
	rootCmd.AddCommand(createUsersCmd)
}
