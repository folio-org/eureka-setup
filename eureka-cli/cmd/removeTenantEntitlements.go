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

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/spf13/cobra"
)

// removeTenantEntitlementsCmd represents the removeTenantEntitlements command
var removeTenantEntitlementsCmd = &cobra.Command{
	Use:   "removeTenantEntitlements",
	Short: "Remove tenant entitlements",
	Long:  `Remove all tenant entitlements.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := New(action.RemoveTenantEntitlements)
		if err != nil {
			return err
		}

		r.Partition(func(consortiumName string, tenantType constant.TenantType) {
			_ = r.RemoveUsers(consortiumName, tenantType)
			_ = r.RemoveRoles(consortiumName, tenantType)
			_ = r.RemoveTenantEntitlements(consortiumName, tenantType)
		})

		return nil
	},
}

func (r *Run) RemoveTenantEntitlements(consortiumName string, tenantType constant.TenantType) error {
	slog.Info(r.Config.Action.Name, "text", "REMOVING TENANT ENTITLEMENTS")
	return r.Config.ManagementSvc.RemoveTenantEntitlements(ap.PurgeSchemas, consortiumName, tenantType)
}

func init() {
	rootCmd.AddCommand(removeTenantEntitlementsCmd)
	removeTenantEntitlementsCmd.PersistentFlags().BoolVarP(&ap.PurgeSchemas, "purgeSchemas", "P", false, "Purge schemas in PostgreSQL on uninstallation")
}
