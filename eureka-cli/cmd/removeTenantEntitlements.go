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
	"log/slog"

	"github.com/folio-org/eureka-setup/eureka-cli/action"
	"github.com/folio-org/eureka-setup/eureka-cli/constant"
	"github.com/spf13/cobra"
)

// removeTenantEntitlementsCmd represents the removeTenantEntitlements command
var removeTenantEntitlementsCmd = &cobra.Command{
	Use:   "removeTenantEntitlements",
	Short: "Remove tenant entitlements",
	Long:  `Remove all tenant entitlements.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		run, err := New(action.RemoveTenantEntitlements)
		if err != nil {
			return err
		}

		return run.ConsortiumPartition(func(consortiumName string, tenantType constant.TenantType) error {
			return run.RemoveTenantEntitlements(consortiumName, tenantType)
		})
	},
}

func (run *Run) RemoveTenantEntitlements(consortiumName string, tenantType constant.TenantType) error {
	slog.Info(run.Config.Action.Name, "text", "REMOVING TENANT ENTITLEMENTS")
	if err := run.setKeycloakMasterAccessTokenIntoContext(constant.ClientCredentials); err != nil {
		return err
	}

	return run.Config.ManagementSvc.RemoveTenantEntitlements(consortiumName, tenantType, params.PurgeSchemas)
}

func init() {
	rootCmd.AddCommand(removeTenantEntitlementsCmd)
	removeTenantEntitlementsCmd.PersistentFlags().BoolVarP(&params.PurgeSchemas, action.PurgeSchemas.Long, action.PurgeSchemas.Short, false, action.PurgeSchemas.Description)
}
