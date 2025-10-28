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
	"github.com/spf13/cobra"
)

// removeUsersCmd represents the removeUsers command
var removeUsersCmd = &cobra.Command{
	Use:   "removeUsers",
	Short: "Create users",
	Long:  `Create all users.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := New(action.RemoveUsers)
		if err != nil {
			return err
		}

		return r.ConsortiumPartitionErr(func(consortiumName string, tenantType constant.TenantType) error {
			return r.RemoveUsers(consortiumName, tenantType)
		})
	},
}

func (r *Run) RemoveUsers(consortiumName string, tenantType constant.TenantType) error {
	return r.TenantPartition(consortiumName, tenantType, func(configTenant string, tenantType constant.TenantType) error {
		slog.Info(r.RunConfig.Action.Name, "text", "REMOVING USERS FOR TENANT", "tenant", configTenant)
		keycloakAccessToken, err := r.RunConfig.KeycloakSvc.GetKeycloakAccessToken(configTenant)
		if err != nil {
			return err
		}
		r.RunConfig.Action.KeycloakAccessToken = keycloakAccessToken

		return r.RunConfig.KeycloakSvc.RemoveUsers(configTenant)
	})
}

func init() {
	rootCmd.AddCommand(removeUsersCmd)
}
