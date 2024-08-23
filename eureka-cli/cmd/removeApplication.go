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

const removeApplicationCommand string = "Remove Application"

// removeApplicationCmd represents the removeApplication command
var removeApplicationCmd = &cobra.Command{
	Use:   "removeApplication",
	Short: "Remove application",
	Long:  `Remove an application.`,
	Run: func(cmd *cobra.Command, args []string) {
		RemoveApplication()
	},
}

func RemoveApplication() {
	slog.Info(createApplicationCommand, internal.MessageKey, "### REMOVING TENANT ENTITLEMENTS ###")
	internal.RemoveTenantEntitlements(removeApplicationCommand, enableDebug)

	slog.Info(createApplicationCommand, internal.MessageKey, "### REMOVING TENANTS ###")
	internal.RemoveTenants(removeApplicationCommand, enableDebug)
}

func init() {
	rootCmd.AddCommand(removeApplicationCmd)
}
