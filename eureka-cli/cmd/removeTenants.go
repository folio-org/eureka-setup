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

const removeTenantsCommand string = "Remove Tenants"

// removeTenantsCmd represents the removeTenants command
var removeTenantsCmd = &cobra.Command{
	Use:   "removeTenants",
	Short: "Remove tenants",
	Long:  `Remove all tenants.`,
	Run: func(cmd *cobra.Command, args []string) {
		RemoveTenants()
	},
}

func RemoveTenants() {
	slog.Info(removeTenantsCommand, internal.GetFuncName(), "### REMOVING TENANTS ###")
	internal.RemoveTenants(removeTenantsCommand, enableDebug, false)
}

func init() {
	rootCmd.AddCommand(removeTenantsCmd)
}
