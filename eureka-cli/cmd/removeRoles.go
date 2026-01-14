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

	"github.com/j011195/eureka-setup/eureka-cli/action"
	"github.com/j011195/eureka-setup/eureka-cli/constant"
	"github.com/spf13/cobra"
)

// removeRolesCmd represents the removeRoles command
var removeRolesCmd = &cobra.Command{
	Use:   "removeRoles",
	Short: "Remove roles",
	Long:  `Remove all roles.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		run, err := New(action.RemoveRoles)
		if err != nil {
			return err
		}

		return run.ConsortiumPartition(func(consortiumName string, tenantType constant.TenantType) error {
			return run.RemoveRoles(consortiumName, tenantType)
		})
	},
}

func (run *Run) RemoveRoles(consortiumName string, tenantType constant.TenantType) error {
	return run.TenantPartition(consortiumName, tenantType, func(configTenant, tenantType string) error {
		slog.Info(run.Config.Action.Name, "text", "REMOVING ROLES", "tenant", configTenant)
		return run.Config.KeycloakSvc.RemoveRoles(configTenant)
	})
}

func init() {
	rootCmd.AddCommand(removeRolesCmd)
}
