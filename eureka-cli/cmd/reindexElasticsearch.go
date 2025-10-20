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
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/field"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// reindexElasticsearchCmd represents the reindexElasticsearch command
var reindexElasticsearchCmd = &cobra.Command{
	Use:   "reindexElasticsearch",
	Short: "Reindex elasticsearch",
	Long:  `Reindex elasticsearch indices.`,
	Run: func(cmd *cobra.Command, args []string) {
		r := NewRun(action.ReindexElasticsearch)
		r.Partition(func(consortiumName string, tenantType constant.TenantType) {
			r.ReindexElasticsearch(consortiumName, tenantType)
		})
	},
}

func (r *Run) ReindexElasticsearch(consortiumName string, tenantType constant.TenantType) {
	vaultRootToken := r.GetVaultRootToken()

	for _, value := range r.Config.ManagementStep.GetTenants(false, consortiumName, tenantType) {
		mapEntry := value.(map[string]any)

		existingTenant := mapEntry["name"].(string)
		if !helpers.HasTenant(existingTenant) {
			continue
		}

		tenantType := mapEntry["description"].(string)
		if viper.IsSet(field.Consortiums) && tenantType != fmt.Sprintf("%s-%s", consortiumName, constant.Central) {
			continue
		}

		slog.Info(r.Config.Action.Name, "text", fmt.Sprintf("REINDEXING ELASTICSEARCH FOR %s TENANT", existingTenant))
		keycloakAccessToken := r.Config.KeycloakStep.GetKeycloakAccessToken(vaultRootToken, existingTenant)
		r.Config.SearchStep.ReindexInventoryRecords(existingTenant, keycloakAccessToken)
		r.Config.SearchStep.ReindexInstanceRecords(existingTenant, keycloakAccessToken)
	}
}

func init() {
	rootCmd.AddCommand(reindexElasticsearchCmd)
}
