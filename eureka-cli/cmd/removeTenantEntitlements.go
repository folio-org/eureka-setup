/*
Copyright © 2024 EPAM_Systems/Thunderjet/Boburbek_Kadirkhodjaev

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

const removeTenantEntitlementsCommand string = "Remove Tenant Entitlements"

// removeTenantEntitlementsCmd represents the removeTenantEntitlements command
var removeTenantEntitlementsCmd = &cobra.Command{
	Use:   "removeTenantEntitlements",
	Short: "Remove tenant entitlements",
	Long:  `Remove all tenant entitlements.`,
	Run: func(cmd *cobra.Command, args []string) {
		RemoveTenantEntitlements()
	},
}

func RemoveTenantEntitlements() {
	slog.Info(removeTenantEntitlementsCommand, internal.GetFuncName(), "### REMOVING TENANT ENTITLEMENTS ###")
	internal.RemoveTenantEntitlements(removeTenantEntitlementsCommand, withEnableDebug, false, withPurgeSchemas)
}

func init() {
	rootCmd.AddCommand(removeTenantEntitlementsCmd)
	removeTenantEntitlementsCmd.PersistentFlags().BoolVarP(&withPurgeSchemas, "purgeSchemas", "P", false, "Purge schemas in PostgreSQL on uninstallation")
}
