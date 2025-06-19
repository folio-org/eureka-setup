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

const createTenantsCommand string = "Create Tenants"

// createTenantsCmd represents the createTenants command
var createTenantsCmd = &cobra.Command{
	Use:   "createTenants",
	Short: "Create tenants",
	Long:  `Create all tenants.`,
	Run: func(cmd *cobra.Command, args []string) {
		CreateTenants()
	},
}

func CreateTenants() {
	slog.Info(createTenantsCommand, internal.GetFuncName(), "### CREATING TENANTS ###")
	internal.CreateTenants(createTenantsCommand, withEnableDebug)
}

func init() {
	rootCmd.AddCommand(createTenantsCmd)
}
