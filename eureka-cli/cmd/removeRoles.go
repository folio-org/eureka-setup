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

const removeRolesCommand string = "Remove Roles"

// removeRolesCmd represents the removeRoles command
var removeRolesCmd = &cobra.Command{
	Use:   "removeRoles",
	Short: "Remove roles",
	Long:  `Remove all roles.`,
	Run: func(cmd *cobra.Command, args []string) {
		RemoveRoles()
	},
}

func RemoveRoles() {
	vaultRootToken := GetVaultRootToken()

	for _, value := range internal.GetTenants(removeRolesCommand, withEnableDebug, false) {
		mapEntry := value.(map[string]any)

		existingTenant := mapEntry["name"].(string)
		if !internal.HasTenant(existingTenant) {
			continue
		}

		slog.Info(removeRolesCommand, internal.GetFuncName(), fmt.Sprintf("### REMOVING ROLES FOR %s TENANT ###", existingTenant))
		keycloakAccessToken := internal.GetKeycloakAccessToken(removeRolesCommand, withEnableDebug, vaultRootToken, existingTenant)
		internal.RemoveRoles(removeRolesCommand, withEnableDebug, false, existingTenant, keycloakAccessToken)
	}
}

func init() {
	rootCmd.AddCommand(removeRolesCmd)
}
