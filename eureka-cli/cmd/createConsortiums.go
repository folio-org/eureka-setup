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
	"github.com/folio-org/eureka-cli/errors"
	"github.com/folio-org/eureka-cli/field"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/spf13/cobra"
)

// createConsortiumCmd represents the createConsortiums command
var createConsortiumsCmd = &cobra.Command{
	Use:   "createConsortiums",
	Short: "Create consortiums",
	Long:  `Create consortiums with multiple tenants.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := New(action.CreateConsortiums)
		if err != nil {
			return err
		}

		return r.CreateConsortium()
	},
}

func (r *Run) CreateConsortium() error {
	if !action.IsSet(field.Consortiums) {
		return nil
	}

	err := r.GetVaultRootToken()
	if err != nil {
		return err
	}

	for consortium, properties := range r.RunConfig.Action.ConfigConsortiums {
		mapEntry := properties.(map[string]any)
		if !helpers.GetBool(mapEntry, field.ConsortiumCreateConsortiumEntry) {
			slog.Info(r.RunConfig.Action.Name, "text", "IGNORING CREATION OF CONSORTIUM", "consortium", consortium)
			continue
		}

		centralTenant := r.RunConfig.ConsortiumSvc.GetConsortiumCentralTenant(consortium)
		if centralTenant == "" {
			return errors.ConsortiumMissingCentralTenant(consortium)
		}

		consortiumTenants := r.RunConfig.ConsortiumSvc.GetSortedConsortiumTenants(consortium)
		consortiumUsers := r.RunConfig.ConsortiumSvc.GetConsortiumUsers(consortium)
		keycloakAccessToken, err := r.RunConfig.KeycloakSvc.GetKeycloakAccessToken(centralTenant)
		if err != nil {
			return err
		}
		r.RunConfig.Action.KeycloakAccessToken = keycloakAccessToken

		slog.Info(r.RunConfig.Action.Name, "text", "CREATING CONSORTIUM", "consortium", consortium)
		consortiumID, err := r.RunConfig.ConsortiumSvc.CreateConsortium(centralTenant, consortium)
		if err != nil {
			return err
		}

		slog.Info(r.RunConfig.Action.Name, "text", "ADDING TENANTS TO CONSORTIUM", "tenants", consortiumTenants, "tenantCount", len(consortiumTenants), "consortium", consortium)
		adminUsername := r.RunConfig.ConsortiumSvc.GetAdminUsername(centralTenant, consortiumUsers)
		err = r.RunConfig.ConsortiumSvc.CreateConsortiumTenants(centralTenant, consortiumID, consortiumTenants, adminUsername)
		if err != nil {
			return err
		}

		if !helpers.GetBool(mapEntry, field.ConsortiumEnableCentralOrderingEntry) {
			slog.Info(r.RunConfig.Action.Name, "text", "IGNORING ENABLEMENT OF CENTRAL ORDERING FOR TENANT IN CONSORTIUM", "tenant", centralTenant, "consortium", consortium)
			continue
		}

		slog.Info(r.RunConfig.Action.Name, "text", "ENABLING CENTRAL ORDERING FOR TENANT IN CONSORTIUM", "tenant", centralTenant, "consortium", consortium)
		err = r.RunConfig.ConsortiumSvc.EnableCentralOrdering(centralTenant)
		if err != nil {
			return err
		}
	}

	return nil
}

func init() {
	rootCmd.AddCommand(createConsortiumsCmd)
}
