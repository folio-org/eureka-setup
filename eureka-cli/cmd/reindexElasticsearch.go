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
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := New(action.ReindexElasticsearch)
		if err != nil {
			return err
		}

		r.Partition(func(consortiumName string, tenantType constant.TenantType) {
			r.ReindexElasticsearch(consortiumName, tenantType)
		})

		return nil
	},
}

func (r *Run) ReindexElasticsearch(consortiumName string, tenantType constant.TenantType) error {
	vaultRootToken := r.GetVaultRootToken()

	foundTenants, _ := r.Config.ManagementStep.GetTenants(consortiumName, tenantType)

	for _, value := range foundTenants {
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
		keycloakAccessToken, err := r.Config.KeycloakStep.GetKeycloakAccessToken(vaultRootToken, existingTenant)
		if err != nil {
			return err
		}

		err = r.Config.SearchStep.ReindexInventoryRecords(existingTenant, keycloakAccessToken)
		if err != nil {
			return err
		}

		err = r.Config.SearchStep.ReindexInstanceRecords(existingTenant, keycloakAccessToken)
		if err != nil {
			return err
		}
	}

	return nil
}

func init() {
	rootCmd.AddCommand(reindexElasticsearchCmd)
}
