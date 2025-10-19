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
	"fmt"
	"log/slog"

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/field"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// createConsortiumCmd represents the createConsortiums command
var createConsortiumsCmd = &cobra.Command{
	Use:   "createConsortiums",
	Short: "Create consortiums",
	Long:  `Create consortiums with multiple tenants.`,
	Run: func(cmd *cobra.Command, args []string) {
		NewRun(action.CreateConsortiums).CreateConsortium()
	},
}

func (r *Run) CreateConsortium() {
	if !viper.IsSet(field.Consortiums) {
		return
	}

	consortiums := viper.GetStringMap(field.Consortiums)
	tenants := viper.GetStringMap(field.Tenants)
	users := viper.GetStringMap(field.Users)

	vaultRootToken := r.GetVaultRootToken()

	for consortium, properties := range consortiums {
		mapEntry := properties.(map[string]any)

		if !helpers.GetBoolKey(mapEntry, field.ConsortiumCreateConsortiumEntry) {
			slog.Info(r.Config.Action.Name, "text", fmt.Sprintf("IGNORING CREATION OF %s CONSORTIUM ", consortium))
			continue
		}

		centralTenant := r.Config.ConsortiumStep.GetConsortiumCentralTenant(consortium, tenants)
		if centralTenant == "" {
			helpers.LogErrorPanic(r.Config.Action, fmt.Errorf("%s consortium does not contain a central tenant", consortium))
			return
		}

		consortiumTenants := r.Config.ConsortiumStep.GetSortedConsortiumTenants(consortium, tenants)
		consortiumUsers := r.Config.ConsortiumStep.GetConsortiumUsers(consortium, users)
		keycloakAccessToken := r.Config.KeycloakStep.GetKeycloakAccessToken(vaultRootToken, centralTenant)

		slog.Info(r.Config.Action.Name, "text", fmt.Sprintf("CREATING %s CONSORTIUM ", consortium))
		consortiumId := r.Config.ConsortiumStep.CreateConsortium(centralTenant, keycloakAccessToken, consortium)

		slog.Info(r.Config.Action.Name, "text", fmt.Sprintf("ADDING %s (%d) TENANTS TO %s CONSORTIUM ", consortiumTenants, len(consortiumTenants), consortium))
		adminUsername := r.Config.ConsortiumStep.GetAdminUsername(centralTenant, consortiumUsers)
		r.Config.ConsortiumStep.CreateConsortiumTenants(centralTenant, keycloakAccessToken, consortiumId, consortiumTenants, adminUsername)

		if !helpers.GetBoolKey(mapEntry, field.ConsortiumEnableCentralOrderingEntry) {
			slog.Info(r.Config.Action.Name, "text", fmt.Sprintf("IGNORING ENABLEMENT OF CENTRAL ORDERING FOR %s TENANT IN %s CONSORTIUM ", centralTenant, consortium))
			continue
		}

		slog.Info(r.Config.Action.Name, "text", fmt.Sprintf("ENABLING CENTRAL ORDERING FOR %s TENANT IN %s CONSORTIUM ", centralTenant, consortium))
		r.Config.ConsortiumStep.EnableCentralOrdering(centralTenant, keycloakAccessToken)
	}
}

func init() {
	rootCmd.AddCommand(createConsortiumsCmd)
}
