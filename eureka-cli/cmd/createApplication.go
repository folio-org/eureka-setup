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

const createApplicationCommand string = "Create Application"

// createApplicationCmd represents the createApplication command
var createApplicationCmd = &cobra.Command{
	Use:   "createApplication",
	Short: "Create application",
	Long:  `Create an application.`,
	Run: func(cmd *cobra.Command, args []string) {
		CreateApplication()
	},
}

func CreateApplication() {
	slog.Info(createApplicationCommand, internal.MessageKey, "### CREATING TENANTS ###")
	internal.CreateTenants(createApplicationCommand, enableDebug)

	slog.Info(createApplicationCommand, internal.MessageKey, "### CREATING TENANT ENTITLEMENTS ###")
	internal.CreateTenantEntitlement(createApplicationCommand, enableDebug)
}

func init() {
	rootCmd.AddCommand(createApplicationCmd)
}
