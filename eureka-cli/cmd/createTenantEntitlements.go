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
	"github.com/folio-org/eureka-cli/tenanttype"
	"github.com/spf13/cobra"
)

// createTenantEntitlementsCmd represents the createTenantEntitlements command
var createTenantEntitlementsCmd = &cobra.Command{
	Use:   "createTenantEntitlements",
	Short: "Create tenant entitlements",
	Long:  `Create all tenant entitlements.`,
	Run: func(cmd *cobra.Command, args []string) {
		r := NewRun(action.CreateTenantEntitlements)
		r.PartitionByConsortiumAndTenantType(func(consortiumName string, tenantType tenanttype.TenantType) {
			r.CreateTenantEntitlements(consortiumName, tenantType)
		})
	},
}

func (r *Run) CreateTenantEntitlements(consortiumName string, tenantType tenanttype.TenantType) {
	slog.Info(r.Config.Action.Name, "text", "CREATING TENANT ENTITLEMENTS")
	r.Config.ManagementStep.CreateTenantEntitlement(consortiumName, tenantType)
}

func init() {
	rootCmd.AddCommand(createTenantEntitlementsCmd)
}
