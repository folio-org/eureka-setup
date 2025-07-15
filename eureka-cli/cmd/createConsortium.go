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
	"sort"

	"github.com/folio-org/eureka-cli/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const createConsortiumCommand string = "Create Consortium"

// createConsortiumCmd represents the createConsortium command
var createConsortiumCmd = &cobra.Command{
	Use:   "createConsortium",
	Short: "Create a consortium",
	Long:  `Create a consortium with multiple tenants.`,
	Run: func(cmd *cobra.Command, args []string) {
		CreateConsortium()
	},
}

func CreateConsortium() {
	if !viper.IsSet(internal.ConsortiumKey) || !viper.IsSet(internal.ConsortiumNameKey) {
		return
	}

	consortiumName := viper.GetString(internal.ConsortiumNameKey)
	tenants := viper.GetStringMap(internal.TenantsKey)
	users := viper.GetStringMap(internal.UsersKey)
	centralTenant := getCentralTenant(tenants)

	vaultRootToken := GetVaultRootToken()
	keycloakAccessToken := internal.GetKeycloakAccessToken(createRolesCommand, withEnableDebug, vaultRootToken, centralTenant)

	slog.Info(createConsortiumCommand, internal.GetFuncName(), fmt.Sprintf("### CREATING %s CONSORTIUM ###", consortiumName))
	consortiumId := internal.CreateConsortium(createConsortiumCommand, withEnableDebug, centralTenant, keycloakAccessToken, consortiumName)

	slog.Info(createConsortiumCommand, internal.GetFuncName(), fmt.Sprintf("### ADDING %d TENANTS TO %s CONSORTIUM ###", len(tenants), consortiumName))
	consortiumTenants := getSortedConsortiumTenants(tenants)
	adminUsername := getAdminUsername(centralTenant, users)
	internal.CreateConsortiumTenants(createConsortiumCommand, withEnableDebug, centralTenant, keycloakAccessToken, consortiumId, consortiumTenants, adminUsername)

	if viper.IsSet(internal.ConsortiumCentralOrderingKey) && viper.GetBool(internal.ConsortiumCentralOrderingKey) {
		slog.Info(createConsortiumCommand, internal.GetFuncName(), fmt.Sprintf("### ENABLING CENTRAL ORDERING FOR %s TENANT ###", centralTenant))
		internal.EnableCentralOrdering(createConsortiumCommand, withEnableDebug, centralTenant, keycloakAccessToken)
	}
}

func getCentralTenant(tenants map[string]any) string {
	for tenant, properties := range tenants {
		if properties == nil {
			continue
		}

		if getSortableIsCentral(properties) == 1 {
			return tenant
		}
	}

	return ""
}

func getSortedConsortiumTenants(tenants map[string]any) internal.ConsortiumTenants {
	var consortiumTenants internal.ConsortiumTenants
	for tenant, properties := range tenants {
		if properties == nil {
			consortiumTenants = append(consortiumTenants, &internal.ConsortiumTenant{Tenant: tenant, IsCentral: 0})
			continue
		}

		consortiumTenants = append(consortiumTenants, &internal.ConsortiumTenant{Tenant: tenant, IsCentral: getSortableIsCentral(properties)})
	}

	sort.Slice(consortiumTenants, func(i, j int) bool {
		return consortiumTenants[i].IsCentral > consortiumTenants[j].IsCentral
	})

	return consortiumTenants
}

func getSortableIsCentral(properties any) int {
	value := properties.(map[string]any)[internal.TenantsCentralTenantEntryKey]
	if value != nil && value.(bool) {
		return 1
	}

	return 0
}

func getAdminUsername(centralTenant string, users map[string]any) string {
	for username, properties := range users {
		tenant := properties.(map[string]any)[internal.UsersTenantEntryKey]
		if tenant != nil && tenant.(string) == centralTenant {
			return username
		}
	}

	return ""
}

func init() {
	rootCmd.AddCommand(createConsortiumCmd)
}
