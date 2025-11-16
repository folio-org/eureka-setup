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
		run, err := New(action.CreateConsortiums)
		if err != nil {
			return err
		}

		return run.CreateConsortium()
	},
}

func (run *Run) CreateConsortium() error {
	if !action.IsSet(field.Consortiums) {
		return nil
	}
	if err := run.GetVaultRootToken(); err != nil {
		return err
	}

	for consortium, properties := range run.Config.Action.ConfigConsortiums {
		entry := properties.(map[string]any)
		if !helpers.GetBool(entry, field.ConsortiumCreateConsortiumEntry) {
			slog.Info(run.Config.Action.Name, "text", "IGNORING CREATION OF CONSORTIUM", "consortium", consortium)
			continue
		}

		centralTenant := run.Config.ConsortiumSvc.GetConsortiumCentralTenant(consortium)
		if centralTenant == "" {
			return errors.ConsortiumMissingCentralTenant(consortium)
		}
		_, err := run.GetKeycloakAccessToken(constant.DefaultToken, centralTenant)
		if err != nil {
			return err
		}

		slog.Info(run.Config.Action.Name, "text", "CREATING CONSORTIUM", "consortium", consortium)
		consortiumID, err := run.Config.ConsortiumSvc.CreateConsortium(centralTenant, consortium)
		if err != nil {
			return err
		}
		consortiumTenants := run.Config.ConsortiumSvc.GetSortedConsortiumTenants(consortium)
		consortiumUsers := run.Config.ConsortiumSvc.GetConsortiumUsers(consortium)

		slog.Info(run.Config.Action.Name, "text", "ADDING TENANTS TO CONSORTIUM", "tenants", consortiumTenants, "count", len(consortiumTenants), "consortium", consortium)
		adminUsername := run.Config.ConsortiumSvc.GetAdminUsername(centralTenant, consortiumUsers)
		if err := run.Config.ConsortiumSvc.CreateConsortiumTenants(centralTenant, consortiumID, consortiumTenants, adminUsername); err != nil {
			return err
		}
		if !helpers.GetBool(entry, field.ConsortiumEnableCentralOrderingEntry) {
			slog.Warn(run.Config.Action.Name, "text", "Ignoring enablement of central ordering", "tenant", centralTenant, "consortium", consortium)
			continue
		}

		slog.Info(run.Config.Action.Name, "text", "ENABLING CENTRAL ORDERING", "tenant", centralTenant, "consortium", consortium)
		if err := run.Config.ConsortiumSvc.EnableCentralOrdering(centralTenant); err != nil {
			return err
		}
	}

	return nil
}

func init() {
	rootCmd.AddCommand(createConsortiumsCmd)
}
