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

// detachCapabilitySetsCmd represents the detachCapabilitySets command
var detachCapabilitySetsCmd = &cobra.Command{
	Use:   "detachCapabilitySets",
	Short: "Detach capability sets",
	Long:  `Detach all capability sets from roles.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		run, err := New(action.DetachCapabilitySets)
		if err != nil {
			return err
		}

		return run.ConsortiumPartition(func(consortiumName string, tenantType constant.TenantType) error {
			return run.DetachCapabilitySets(consortiumName, tenantType)
		})
	},
}

func (run *Run) DetachCapabilitySets(consortiumName string, tenantType constant.TenantType) error {
	return run.TenantPartition(consortiumName, tenantType, func(configTenant, tenantType string) error {
		slog.Info(run.Config.Action.Name, "text", "DETACHING CAPABILITY SETS", "tenant", configTenant)
		if err := run.Config.KeycloakSvc.DetachCapabilitySetsFromRoles(configTenant); err != nil {
			slog.Warn(run.Config.Action.Name, "text", "Capability sets detachment was unsuccessful", "tenant", configTenant, "error", err)
		}

		return nil
	})
}

func init() {
	rootCmd.AddCommand(detachCapabilitySetsCmd)
}
