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

const createRolesCommand string = "Create Roles"

// createRolesCmd represents the createRoles command
var createRolesCmd = &cobra.Command{
	Use:   "createRoles",
	Short: "Create roles",
	Long:  `Create all roles.`,
	Run: func(cmd *cobra.Command, args []string) {
		CreateRoles()
	},
}

func CreateRoles() {
	slog.Info(createRolesCommand, "### ACQUIRING VAULT ROOT TOKEN ###", "")
	client := internal.CreateClient(createRolesCommand)
	defer client.Close()
	vaultRootToken := internal.GetRootVaultToken(createRolesCommand, client)

	for _, value := range internal.GetTenants(createRolesCommand, enableDebug, false) {
		mapEntry := value.(map[string]interface{})
		tenant := mapEntry["name"].(string)

		if !slices.Contains(viper.GetStringSlice(internal.TenantsKey), tenant) {
			continue
		}

		slog.Info(createRolesCommand, "### ACQUIRING KEYCLOAK ACCESS TOKEN ###", "")
		accessToken := internal.GetKeycloakAccessToken(createRolesCommand, enableDebug, vaultRootToken, tenant)

		slog.Info(createRolesCommand, "### CREATING ROLES ###", "")
		internal.CreateRoles(createRolesCommand, enableDebug, accessToken)
	}
}

func init() {
	rootCmd.AddCommand(createRolesCmd)
}
