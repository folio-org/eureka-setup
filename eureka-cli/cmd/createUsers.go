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
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/folio-org/eureka-cli/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const createUsersCommand string = "Create Users"

// createUsersCmd represents the createUsers command
var createUsersCmd = &cobra.Command{
	Use:   "createUsers",
	Short: "Create users",
	Long:  `Create all users.`,
	Run: func(cmd *cobra.Command, args []string) {
		CreateUsers()
	},
}

func CreateUsers() {
	slog.Info(deployModulesCommand, "### ACQUIRING VAULT ROOT TOKEN ###", "")
	client := internal.CreateClient(createUsersCommand)
	defer client.Close()
	vaultRootToken := internal.GetRootVaultToken(createUsersCommand, client)

	for _, value := range internal.GetTenants(createUsersCommand, enableDebug, true) {
		mapEntry := value.(map[string]interface{})
		tenantName := mapEntry["name"].(string)

		if slices.Contains(viper.GetStringSlice(internal.TenantsKey), tenantName) {
			slog.Info(createUsersCommand, fmt.Sprintf("### ACQUIRING VAULT SECRETS FOR %s TENANT ###", strings.ToUpper(tenantName)), "")
			secretMap := internal.GetVaultKey(createUsersCommand, enableDebug, vaultRootToken, fmt.Sprintf("folio/%s", tenantName))
			clientId := internal.GetEnvironmentFromMapByKey(createUsersCommand, "KC_SERVICE_CLIENT_ID")
			clientSecret := secretMap[clientId].(string)
			systemUser := fmt.Sprintf("%s-system-user", tenantName)
			systemUserPassword := secretMap[systemUser].(string)

			slog.Info(createUsersCommand, "### ACQUIRING KEYCLOAK ACCESS TOKEN ###", "")
			dto := internal.NewDirectGrantTokenDto(tenantName, clientId, clientSecret, systemUser, systemUserPassword)
			accessToken := internal.GetKeycloakAccessToken(createUsersCommand, enableDebug, dto)

			slog.Info(createUsersCommand, "### CREATING USERS ###", "")
			internal.CreateUsers(createUsersCommand, enableDebug, accessToken)
		}
	}
}

func init() {
	rootCmd.AddCommand(createUsersCmd)
}
