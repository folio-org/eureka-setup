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

const reindexElasticsearchCommand = "Reindex Elasticsearch"

// reindexElasticsearchCmd represents the reindexElasticsearch command
var reindexElasticsearchCmd = &cobra.Command{
	Use:   "reindexElasticsearch",
	Short: "Reindex elasticsearch",
	Long:  `Reindex elasticsearch indices.`,
	Run: func(cmd *cobra.Command, args []string) {
		RunByConsortiumAndTenantType(reindexElasticsearchCommand, func(consortium string, tenantType internal.TenantType) {
			ReindexElasticsearch(consortium, tenantType)
		})
	},
}

func ReindexElasticsearch(consortium string, tenantType internal.TenantType) {
	vaultRootToken := GetVaultRootToken()

	for _, value := range internal.GetTenants(reindexElasticsearchCommand, withEnableDebug, false, consortium, tenantType) {
		mapEntry := value.(map[string]any)

		existingTenant := mapEntry["name"].(string)
		if !internal.HasTenant(existingTenant) {
			continue
		}

		tenantType := mapEntry["description"].(string)
		if viper.IsSet(internal.ConsortiumsKey) && tenantType != fmt.Sprintf("%s-%s", consortium, internal.CentralTenantType) {
			continue
		}

		slog.Info(reindexElasticsearchCommand, internal.GetFuncName(), fmt.Sprintf("### REINDEXING ELASTICSEARCH FOR %s TENANT ###", existingTenant))
		keycloakAccessToken := internal.GetKeycloakAccessToken(reindexElasticsearchCommand, withEnableDebug, vaultRootToken, existingTenant)
		internal.ReindexInventoryRecords(reindexElasticsearchCommand, withEnableDebug, existingTenant, keycloakAccessToken)
		internal.ReindexInstanceRecords(reindexElasticsearchCommand, withEnableDebug, existingTenant, keycloakAccessToken)
	}
}

func init() {
	rootCmd.AddCommand(reindexElasticsearchCmd)
}
