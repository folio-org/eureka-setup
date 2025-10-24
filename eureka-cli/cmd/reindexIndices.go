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

// reindexIndicesCmd represents the reindexIndices command
var reindexIndicesCmd = &cobra.Command{
	Use:   "reindexIndices",
	Short: "Reindex elasticsearch",
	Long:  `Reindex elasticsearch indices.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := New(action.ReindexIndices)
		if err != nil {
			return err
		}

		return r.PartitionErr(func(consortiumName string, tenantType constant.TenantType) error {
			return r.ReindexIndices(consortiumName, tenantType)
		})
	},
}

func (r *Run) ReindexIndices(consortiumName string, tenantType constant.TenantType) error {
	vaultRootToken, err := r.GetVaultRootToken()
	if err != nil {
		return err
	}

	tt, err := r.Config.ManagementSvc.GetTenants(consortiumName, tenantType)
	if err != nil {
		return err
	}

	for _, value := range tt {
		mapEntry := value.(map[string]any)

		existingTenant := mapEntry["name"].(string)
		if !helpers.HasTenant(existingTenant) {
			continue
		}

		tenantType := mapEntry["description"].(string)
		if viper.IsSet(field.Consortiums) && tenantType != fmt.Sprintf("%s-%s", consortiumName, constant.Central) {
			continue
		}

		slog.Info(r.Config.Action.Name, "text", "REINDEXING INDICES FOR TENANT", "tenant", existingTenant)
		keycloakAccessToken, err := r.Config.KeycloakSvc.GetKeycloakAccessToken(vaultRootToken, existingTenant)
		if err != nil {
			return err
		}

		err = r.Config.SearchSvc.ReindexInventoryRecords(existingTenant, keycloakAccessToken)
		if err != nil {
			return err
		}

		err = r.Config.SearchSvc.ReindexInstanceRecords(existingTenant, keycloakAccessToken)
		if err != nil {
			return err
		}
	}

	return nil
}

func init() {
	rootCmd.AddCommand(reindexIndicesCmd)
}
