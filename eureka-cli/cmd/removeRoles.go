/*
Copyright © 2024 EPAM_Systems/Thunderjet/Boburbek_Kadirkhodjaev

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
	"slices"

	"github.com/folio-org/eureka-cli/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	if internal.RemoveRoleUnsupported {
		slog.Info(removeRolesCommand, "### REMOVAL OF ROLES IS UNSUPPORTED BY CURRENT GATEWAY SETUP ###", "")
		return
	}

	slog.Info(removeRolesCommand, "### ACQUIRING VAULT ROOT TOKEN ###", "")
	client := internal.CreateClient(removeRolesCommand)
	defer client.Close()
	vaultRootToken := internal.GetRootVaultToken(removeRolesCommand, client)

	for _, value := range internal.GetTenants(removeRolesCommand, enableDebug, false) {
		mapEntry := value.(map[string]interface{})
		tenant := mapEntry["name"].(string)

		if !slices.Contains(viper.GetStringSlice(internal.TenantsKey), tenant) {
			continue
		}

		slog.Info(removeRolesCommand, "### ACQUIRING KEYCLOAK ACCESS TOKEN ###", "")
		accessToken := internal.GetKeycloakAccessToken(removeRolesCommand, enableDebug, vaultRootToken, tenant)

		slog.Info(removeRolesCommand, "### REMOVING ROLES ###", "")
		internal.RemoveRoles(removeRolesCommand, enableDebug, false, tenant, accessToken)
	}
}

func init() {
	rootCmd.AddCommand(removeRolesCmd)
}
