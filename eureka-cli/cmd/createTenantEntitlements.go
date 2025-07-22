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

	"github.com/folio-org/eureka-cli/internal"
	"github.com/spf13/cobra"
)

const createTenantEntitlementsCommand string = "Create Tenant Entitlements"

// createTenantEntitlementsCmd represents the createTenantEntitlements command
var createTenantEntitlementsCmd = &cobra.Command{
	Use:   "createTenantEntitlements",
	Short: "Create tenant entitlements",
	Long:  `Create all tenant entitlements.`,
	Run: func(cmd *cobra.Command, args []string) {
		RunByConsortiumAndTenantType(createTenantEntitlementsCommand, func(consortium string, tenantType internal.TenantType) {
			CreateTenantEntitlements(consortium, tenantType)
		})
	},
}

func CreateTenantEntitlements(consortium string, tenantType internal.TenantType) {
	slog.Info(createTenantEntitlementsCommand, internal.GetFuncName(), "### CREATING TENANT ENTITLEMENTS ###")
	internal.CreateTenantEntitlement(createTenantEntitlementsCommand, withEnableDebug, consortium, tenantType)
}

func init() {
	rootCmd.AddCommand(createTenantEntitlementsCmd)
}
