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

// createUsersCmd represents the createUsers command
var createUsersCmd = &cobra.Command{
	Use:   "createUsers",
	Short: "Create users",
	Long:  `Create all users.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		run, err := New(action.CreateUsers)
		if err != nil {
			return err
		}

		return run.ConsortiumPartition(func(consortiumName string, tenantType constant.TenantType) error {
			return run.CreateUsers(consortiumName, tenantType)
		})
	},
}

func (run *Run) CreateUsers(consortiumName string, tenantType constant.TenantType) error {
	return run.TenantPartition(consortiumName, tenantType, func(configTenant, tenantType string) error {
		slog.Info(run.Config.Action.Name, "text", "CREATING USERS", "tenant", configTenant)
		return run.Config.KeycloakSvc.CreateUsers(configTenant)
	})
}

func init() {
	rootCmd.AddCommand(createUsersCmd)
}
