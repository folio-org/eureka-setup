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
	"time"

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/errors"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/spf13/cobra"
)

// attachCapabilitySetsCmd represents the attachCapabilitySets command
var attachCapabilitySetsCmd = &cobra.Command{
	Use:   "attachCapabilitySets",
	Short: "Attach capability sets",
	Long:  `Attach capability sets to roles.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := New(action.AttachCapabilitySets)
		if err != nil {
			return err
		}

		return r.ConsortiumPartitionErr(func(consortiumName string, tenantType constant.TenantType) error {
			return r.AttachCapabilitySets(consortiumName, tenantType, time.Duration(0*time.Second))
		})
	},
}

func (r *Run) AttachCapabilitySets(consortiumName string, tenantType constant.TenantType, initialWait time.Duration) error {
	// TODO Abstract
	err := r.GetVaultRootToken()
	if err != nil {
		return err
	}

	resp, err := r.RunConfig.ManagementSvc.GetTenants(consortiumName, tenantType)
	if err != nil {
		return err
	}

	for _, value := range resp {
		mapEntry := value.(map[string]any)
		configTenant := mapEntry["name"].(string)
		hasTenant := helpers.HasTenant(configTenant, r.RunConfig.Action.ConfigTenants)
		if !hasTenant {
			continue
		}
		if initialWait > 0 {
			time.Sleep(initialWait)
		}

		slog.Info(r.RunConfig.Action.Name, "text", "POLLING FOR CAPABILITY SETS CREATION")
		err := r.pollCapabilitySetsCreation(configTenant)
		if err != nil {
			return err
		}

		slog.Info(r.RunConfig.Action.Name, "text", "ATTACHING CAPABILITY SETS TO ROLES FOR TENANT", "tenant", configTenant)
		keycloakAccessToken, err := r.RunConfig.KeycloakSvc.GetKeycloakAccessToken(configTenant)
		if err != nil {
			return err
		}
		r.RunConfig.Action.KeycloakAccessToken = keycloakAccessToken

		err = r.RunConfig.KeycloakSvc.AttachCapabilitySetsToRoles(configTenant)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Run) pollCapabilitySetsCreation(tenantName string) error {
	consumerGroup := fmt.Sprintf("%s-%s", r.RunConfig.Action.ConfigEnvFolio, constant.ConsumerGroupSuffix)
	retryCount := 0
	slog.Info(r.RunConfig.Action.Name, "text", "Checking Kafka readiness before polling consumer groups")
	if err := r.RunConfig.KafkaSvc.CheckReadiness(); err != nil {
		slog.Info(r.RunConfig.Action.Name, "text", "Kafka not fully ready, proceeding with polling", "error", err)
	}

	var lag int
	for {
		lag, err := r.RunConfig.KafkaSvc.GetConsumerGroupLag(tenantName, consumerGroup, lag)
		if err != nil {
			retryCount++
			if retryCount >= constant.ConsumerGroupRebalanceRetries {
				return errors.ConsumerGroupRebalanceTimeout(consumerGroup, err)
			}

			slog.Info(r.RunConfig.Action.Name, "text", "Retry: Error polling consumer group, retrying", "retryCount", retryCount, "maxRetries", constant.ConsumerGroupRebalanceRetries, "waitSeconds", constant.AttachCapabilitySetsRebalanceWait.Seconds())
			time.Sleep(constant.AttachCapabilitySetsRebalanceWait)
			continue
		}

		retryCount = 0
		if lag == 0 {
			break
		}
		slog.Info(r.RunConfig.Action.Name, "text", "Waiting for consumer group to process", "waitSeconds", constant.AttachCapabilitySetsPollWait.Seconds(), "consumerGroup", consumerGroup, "lag", lag)
		time.Sleep(constant.AttachCapabilitySetsPollWait)
	}
	slog.Info(r.RunConfig.Action.Name, "text", "Consumer group has no new message to process", "consumerGroup", consumerGroup)

	return nil
}

func init() {
	rootCmd.AddCommand(attachCapabilitySetsCmd)
}
