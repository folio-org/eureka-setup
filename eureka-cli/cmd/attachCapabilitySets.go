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
	"time"

	"github.com/j011195/eureka-setup/eureka-cli/action"
	"github.com/j011195/eureka-setup/eureka-cli/constant"
	"github.com/spf13/cobra"
)

// attachCapabilitySetsCmd represents the attachCapabilitySets command
var attachCapabilitySetsCmd = &cobra.Command{
	Use:   "attachCapabilitySets",
	Short: "Attach capability sets",
	Long:  `Attach capability sets to roles.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		run, err := New(action.AttachCapabilitySets)
		if err != nil {
			return err
		}

		return run.ConsortiumPartition(func(consortiumName string, tenantType constant.TenantType) error {
			return run.AttachCapabilitySets(consortiumName, tenantType, time.Duration(0*time.Second))
		})
	},
}

func (run *Run) AttachCapabilitySets(consortiumName string, tenantType constant.TenantType, initialWait time.Duration) error {
	if err := run.setKeycloakMasterAccessTokenIntoContext(constant.Password); err != nil {
		return err
	}

	return run.TenantPartition(consortiumName, tenantType, func(configTenant, tenantType string) error {
		if initialWait > 0 {
			time.Sleep(initialWait)
		}
		if err := run.updateRealmAccessTokenSettingsAndRelogin(configTenant); err != nil {
			return err
		}
		topicConfigTenant := run.Config.Action.GetKafkaTopicConfigTenant(configTenant)

		slog.Info(run.Config.Action.Name, "text", "POLLING FOR CAPABILITY SETS CREATION", "topicConfigTenant", topicConfigTenant)
		if err := run.Config.KafkaSvc.PollConsumerGroup(topicConfigTenant); err != nil {
			return err
		}

		slog.Info(run.Config.Action.Name, "text", "ATTACHING CAPABILITY SETS", "tenant", configTenant)
		return run.Config.KeycloakSvc.AttachCapabilitySetsToRoles(configTenant)
	})
}

func (run *Run) updateRealmAccessTokenSettingsAndRelogin(configTenant string) error {
	if err := run.Config.KeycloakSvc.UpdateRealmAccessTokenSettings(configTenant, constant.KeycloakTenantRealmAccessTokenLifespan); err != nil {
		return err
	}
	if err := run.setKeycloakAccessTokenIntoContext(configTenant); err != nil {
		return err
	}
	slog.Info(run.Config.Action.Name, "text", "New access token was set into context", "tenant", configTenant)

	return nil
}

func init() {
	rootCmd.AddCommand(attachCapabilitySetsCmd)
}
