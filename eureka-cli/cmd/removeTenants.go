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

// removeTenantsCmd represents the removeTenants command
var removeTenantsCmd = &cobra.Command{
	Use:   "removeTenants",
	Short: "Remove tenants",
	Long:  `Remove all tenants.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := New(action.RemoveTenants)
		if err != nil {
			return err
		}

		r.Partition(func(consortiumName string, tenantType constant.TenantType) {
			_ = r.RemoveUsers(consortiumName, tenantType)
			_ = r.RemoveRoles(consortiumName, tenantType)
			_ = r.RemoveTenantEntitlements(consortiumName, tenantType)
			_ = r.RemoveTenants(consortiumName, tenantType)
		})

		return nil
	},
}

func (r *Run) RemoveTenants(consortiumName string, tenantType constant.TenantType) error {
	slog.Info(r.Config.Action.Name, "text", "REMOVING TENANTS")
	return r.Config.ManagementSvc.RemoveTenants(consortiumName, tenantType)
}

func init() {
	rootCmd.AddCommand(removeTenantsCmd)
}
