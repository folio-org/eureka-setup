/*
Copyright © 2024 EPAM_Systems/Thunderjet/Boburbek_Kadirkhodjaev

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
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/folio-org/eureka-cli/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	attachCapabilitySetsCommand string = "Attach Capability Sets"
	NewLinePattern              string = `\r?\n`
	KafkaUrl                    string = "kafka.eureka:9092"
	ConsumerGroupSuffix         string = "mod-roles-keycloak-capability-group"
)

// attachCapabilitySetsCmd represents the attachCapabilitySets command
var attachCapabilitySetsCmd = &cobra.Command{
	Use:   "attachCapabilitySets",
	Short: "Attach capability sets",
	Long:  `Attach capability sets to roles.`,
	Run: func(cmd *cobra.Command, args []string) {
		AttachCapabilitySets()
	},
}

func AttachCapabilitySets() {
	vaultRootToken := GetVaultRootToken()

	for _, tenantValue := range internal.GetTenants(attachCapabilitySetsCommand, enableDebug, false) {
		tenantMapEntry := tenantValue.(map[string]any)

		existingTenant := tenantMapEntry["name"].(string)
		if !internal.HasTenant(existingTenant) {
			continue
		}

		slog.Info(attachCapabilitySetsCommand, internal.GetFuncName(), "### POLLING FOR CAPABILITY SETS CREATION ###")
		pollCapabilitySetsCreation(existingTenant)

		slog.Info(attachCapabilitySetsCommand, internal.GetFuncName(), fmt.Sprintf("### ATTACHING CAPABILITY SETS TO ROLES FOR %s TENANT ###", existingTenant))
		keycloakAccessToken := internal.GetKeycloakAccessToken(attachCapabilitySetsCommand, enableDebug, vaultRootToken, existingTenant)
		internal.AttachCapabilitySetsToRoles(attachCapabilitySetsCommand, enableDebug, existingTenant, keycloakAccessToken)
	}
}

func pollCapabilitySetsCreation(tenant string) {
	consumerGroup := fmt.Sprintf("%s-%s", viper.GetString(internal.EnvironmentFolioKey), ConsumerGroupSuffix)

	for {
		lag := getConsumerGroupLag(tenant, consumerGroup)
		if lag == 0 {
			break
		}

		slog.Info(attachCapabilitySetsCommand, internal.GetFuncName(), fmt.Sprintf("Waiting for %s consumer group to process all messages, lag: %d", consumerGroup, lag))
		time.Sleep(5 * time.Second)
	}

	slog.Info(attachCapabilitySetsCommand, internal.GetFuncName(), fmt.Sprintf("Consumer group %s has no new message to process", consumerGroup))
}

func getConsumerGroupLag(tenant string, consumerGroup string) int {
	stdout, stderr := internal.RunCommandReturnOutput(listSystemCommand, exec.Command("docker", "exec", "-i", "kafka", "bash", "-c",
		fmt.Sprintf("kafka-consumer-groups.sh --bootstrap-server %s --describe --group %s | grep %s | awk '{print $6}'", KafkaUrl, consumerGroup, tenant)))
	if stderr.Len() > 0 {
		internal.LogErrorPrintStderrPanic(attachCapabilitySetsCommand, "internal.RunCommandReturnOutput error", stderr.String())
		return 0
	}

	lag, err := strconv.Atoi(strings.TrimSpace(regexp.MustCompile(NewLinePattern).ReplaceAllString(stdout.String(), "")))
	if err != nil {
		slog.Error(attachCapabilitySetsCommand, internal.GetFuncName(), "strconv.Atoi error")
		panic(err)
	}

	return lag
}

func init() {
	rootCmd.AddCommand(attachCapabilitySetsCmd)
}
