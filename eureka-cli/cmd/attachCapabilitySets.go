/*
Copyright © 2026 Open Library Foundation

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
	"os"
	"path/filepath"
	"time"

	"github.com/folio-org/eureka-setup/eureka-cli/action"
	"github.com/folio-org/eureka-setup/eureka-cli/constant"
	"github.com/folio-org/eureka-setup/eureka-cli/helpers"
	"github.com/spf13/cobra"
)

type capabilitySetsRecord struct {
	Tenant string `json:"tenant"`
	Total  int    `json:"total"`
}

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
			return run.AttachCapabilitySets(consortiumName, tenantType, time.Duration(0*time.Second), false)
		})
	},
}

func (run *Run) AttachCapabilitySets(consortiumName string, tenantType constant.TenantType, initialWait time.Duration, forceRefresh bool) error {
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

		slog.Info(run.Config.Action.Name, "text", "ATTACHING CAPABILITY SETS", "tenant", configTenant)
		homeDir, err := helpers.GetHomeDirPath()
		if err != nil {
			return err
		}
		filePath := filepath.Join(homeDir, fmt.Sprintf(constant.CapabilitySetsFilePattern, configTenant))

		if forceRefresh {
			if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
				slog.Warn(run.Config.Action.Name, "text", "Could not delete capability sets file", "tenant", configTenant, "error", err)
			}
		}

		skipPoll := false
		var record capabilitySetsRecord
		if err := helpers.ReadJSONFromFile(filePath, &record); err == nil && record.Total > 0 {
			liveCount, countErr := run.Config.KeycloakSvc.CountCapabilitySets(configTenant)
			if countErr != nil {
				slog.Warn(run.Config.Action.Name, "text", "Could not count capability sets, polling broker", "tenant", configTenant, "error", countErr)
			} else if liveCount == record.Total {
				slog.Info(run.Config.Action.Name, "text", "Capability sets match persisted count, skipping Kafka poll",
					"tenant", configTenant, "total", liveCount)
				skipPoll = true
			} else {
				slog.Info(run.Config.Action.Name, "text", "Capability sets count changed, polling broker",
					"tenant", configTenant, "persisted", record.Total, "live", liveCount)
			}
		} else {
			slog.Info(run.Config.Action.Name, "text", "No capability sets file found, polling broker", "tenant", configTenant)
		}
		if !skipPoll {
			topicConfigTenant := run.Config.Action.GetKafkaTopicConfigTenant(configTenant)
			slog.Info(run.Config.Action.Name, "text", "POLLING FOR CAPABILITY SETS CREATION", "topicConfigTenant", topicConfigTenant)
			if err := run.Config.KafkaSvc.PollConsumerGroup(topicConfigTenant); err != nil {
				return err
			}
		}
		if err := run.Config.KeycloakSvc.AttachCapabilitySetsToRoles(configTenant); err != nil {
			return err
		}

		count, err := run.Config.KeycloakSvc.CountCapabilitySets(configTenant)
		if err != nil {
			slog.Warn(run.Config.Action.Name, "text", "Could not count capability sets, skipping persistence", "tenant", configTenant, "error", err)
			return nil
		}
		if err := helpers.WriteJSONToFile(filePath, capabilitySetsRecord{Tenant: configTenant, Total: count}); err != nil {
			slog.Warn(run.Config.Action.Name, "text", "Could not write capability sets file", "tenant", configTenant, "error", err)
		} else {
			slog.Info(run.Config.Action.Name, "text", "Persisted capability sets count", "tenant", configTenant, "total", count)
		}

		return nil
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
