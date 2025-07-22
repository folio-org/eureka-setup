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
	"github.com/spf13/viper"
)

const createConsortiumsCommand string = "Create Consortiums"

// createConsortiumCmd represents the createConsortiums command
var createConsortiumsCmd = &cobra.Command{
	Use:   "createConsortiums",
	Short: "Create consortiums",
	Long:  `Create consortiums with multiple tenants.`,
	Run: func(cmd *cobra.Command, args []string) {
		CreateConsortium()
	},
}

func CreateConsortium() {
	if !viper.IsSet(internal.ConsortiumsKey) {
		return
	}

	consortiums := viper.GetStringMap(internal.ConsortiumsKey)
	tenants := viper.GetStringMap(internal.TenantsKey)
	users := viper.GetStringMap(internal.UsersKey)

	vaultRootToken := GetVaultRootToken()

	for consortium, properties := range consortiums {
		mapEntry := properties.(map[string]any)

		if !internal.GetBoolKey(mapEntry, internal.ConsortiumCreateConsortiumEntryKey) {
			slog.Info(createConsortiumsCommand, internal.GetFuncName(), fmt.Sprintf("### IGNORING CREATION OF %s CONSORTIUM ###", consortium))
			continue
		}

		centralTenant := internal.GetConsortiumCentralTenant(consortium, tenants)
		if centralTenant == "" {
			internal.LogErrorPanic(createConsortiumsCommand, fmt.Sprintf("internal.GetConsortiumCentralTenant error - %s consortium does not contain a central tenant", consortium))
			return
		}

		consortiumTenants := internal.GetSortedConsortiumTenants(consortium, tenants)
		consortiumUsers := internal.GetConsortiumUsers(consortium, users)
		keycloakAccessToken := internal.GetKeycloakAccessToken(createRolesCommand, withEnableDebug, vaultRootToken, centralTenant)

		slog.Info(createConsortiumsCommand, internal.GetFuncName(), fmt.Sprintf("### CREATING %s CONSORTIUM ###", consortium))
		consortiumId := internal.CreateConsortium(createConsortiumsCommand, withEnableDebug, centralTenant, keycloakAccessToken, consortium)

		slog.Info(createConsortiumsCommand, internal.GetFuncName(), fmt.Sprintf("### ADDING %s (%d) TENANTS TO %s CONSORTIUM ###", consortiumTenants, len(consortiumTenants), consortium))
		adminUsername := internal.GetAdminUsername(centralTenant, consortiumUsers)
		internal.CreateConsortiumTenants(createConsortiumsCommand, withEnableDebug, centralTenant, keycloakAccessToken, consortiumId, consortiumTenants, adminUsername)

		if !internal.GetBoolKey(mapEntry, internal.ConsortiumEnableCentralOrderingEntryKey) {
			slog.Info(createConsortiumsCommand, internal.GetFuncName(), fmt.Sprintf("### IGNORING ENABLEMENT OF CENTRAL ORDERING FOR %s TENANT IN %s CONSORTIUM ###", centralTenant, consortium))
			continue
		}

		slog.Info(createConsortiumsCommand, internal.GetFuncName(), fmt.Sprintf("### ENABLING CENTRAL ORDERING FOR %s TENANT IN %s CONSORTIUM ###", centralTenant, consortium))
		internal.EnableCentralOrdering(createConsortiumsCommand, withEnableDebug, centralTenant, keycloakAccessToken)
	}
}

func init() {
	rootCmd.AddCommand(createConsortiumsCmd)
}
