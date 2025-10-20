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
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/field"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// attachCapabilitySetsCmd represents the attachCapabilitySets command
var attachCapabilitySetsCmd = &cobra.Command{
	Use:   "attachCapabilitySets",
	Short: "Attach capability sets",
	Long:  `Attach capability sets to roles.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		r := NewRun(action.AttachCapabilitySets)
		return r.PartitionErr(func(consortiumName string, tenantType constant.TenantType) error {
			err := r.AttachCapabilitySets(consortiumName, tenantType, time.Duration(0*time.Second))
			if err != nil {
				return err
			}

			return nil
		})
	},
}

func (r *Run) AttachCapabilitySets(consortiumName string, tenantType constant.TenantType, initialWaitDuration time.Duration) error {
	vaultRootToken := r.GetVaultRootToken()

	for _, value := range r.Config.ManagementStep.GetTenants(false, consortiumName, tenantType) {
		mapEntry := value.(map[string]any)

		existingTenant := mapEntry["name"].(string)
		if !helpers.HasTenant(existingTenant) {
			continue
		}
		if initialWaitDuration > 0 {
			time.Sleep(initialWaitDuration)
		}

		slog.Info(r.Config.Action.Name, "text", "POLLING FOR CAPABILITY SETS CREATION")
		err := r.pollCapabilitySetsCreation(existingTenant)
		if err != nil {
			return err
		}

		slog.Info(r.Config.Action.Name, "text", fmt.Sprintf("ATTACHING CAPABILITY SETS TO ROLES FOR %s TENANT", existingTenant))
		keycloakAccessToken := r.Config.KeycloakStep.GetKeycloakAccessToken(vaultRootToken, existingTenant)
		r.Config.ManagementStep.AttachCapabilitySetsToRoles(existingTenant, keycloakAccessToken)
	}

	return nil
}

func (r *Run) pollCapabilitySetsCreation(tenant string) error {
	consumerGroup := fmt.Sprintf("%s-%s", viper.GetString(field.EnvFolio), constant.ConsumerGroupSuffix)
	pollWaitDuration := 30 * time.Second

	var lag int
	for {
		lag, err := r.getConsumerGroupLag(tenant, consumerGroup, lag)
		if err != nil {
			return err
		}

		if lag == 0 {
			break
		}

		slog.Info(r.Config.Action.Name, "text", fmt.Sprintf("waiting for %.1f sec for %s consumer group to process, lag: %d", pollWaitDuration.Seconds(), consumerGroup, lag))
		time.Sleep(pollWaitDuration)
	}
	slog.Info(r.Config.Action.Name, "text", fmt.Sprintf("consumer group %s has no new message to process", consumerGroup))

	return nil
}

func (r *Run) getConsumerGroupLag(tenant string, consumerGroup string, initialLag int) (lag int, err error) {
	stdout, stderr, err := helpers.ExecReturnOutput(exec.Command("docker", "exec", "-i", "kafka-tools", "bash", "-c",
		fmt.Sprintf("kafka-consumer-groups.sh --bootstrap-server %s --describe --group %s | grep %s | awk '{print $6}'", constant.KafkaTCP, consumerGroup, tenant)))
	if err != nil {
		return initialLag, err
	}

	if stderr.Len() > 0 {
		if strings.Contains(stderr.String(), constant.ErrNoActiveMembers) || strings.Contains(stderr.String(), constant.ErrRebalancing) {
			time.Sleep(30 * time.Second)

			return initialLag, nil
		}
		helpers.LogErrorPrintStderrPanic(r.Config.Action, "failed to execute a container command", stderr.String())

		return 0, nil
	}

	lag, err = strconv.Atoi(helpers.GetKafkaConsumerLagFromLogLine(stdout))
	if err != nil {
		slog.Error(r.Config.Action.Name, "error", err.Error())

		return initialLag, nil
	}

	return lag, nil
}

func init() {
	rootCmd.AddCommand(attachCapabilitySetsCmd)
}
