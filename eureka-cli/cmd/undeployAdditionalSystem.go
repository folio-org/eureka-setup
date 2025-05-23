/*
Copyright © 2025 EPAM_Systems/Thunderjet/Boburbek_Kadirkhodjaev

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
	"os/exec"

	"github.com/folio-org/eureka-cli/internal"
	"github.com/spf13/cobra"
)

const undeployAdditionalSystemCommand string = "Undeploy Additional System"

// undeployAdditionalSystemCmd represents the undeployAdditionalSystem command
var undeployAdditionalSystemCmd = &cobra.Command{
	Use:   "undeployAdditionalSystem",
	Short: "Undeploy additional system",
	Long:  `Undeploy additional system containers.`,
	Run: func(cmd *cobra.Command, args []string) {
		UndeployAdditionalSystem()
	},
}

func UndeployAdditionalSystem() {
	slog.Info(undeployAdditionalSystemCommand, internal.GetFuncName(), "### UNDEPLOYING ADDITIONAL SYSTEM CONTAINERS ###")

	additionalRequiredContainers := internal.GetRequiredContainers(undeployAdditionalSystemCommand, []string{})
	if len(additionalRequiredContainers) == 0 {
		slog.Info(undeployAdditionalSystemCommand, internal.GetFuncName(), "No addititional system containers undeployed")
		return
	}

	subCommand := append([]string{"compose", "--progress", "plain", "--ansi", "never", "--project-name", "eureka", "stop"}, additionalRequiredContainers...)
	internal.RunCommand(undeployAdditionalSystemCommand, exec.Command("docker", subCommand...))

	subCommand = append([]string{"compose", "--progress", "plain", "--ansi", "never", "--project-name", "eureka", "rm", "--volumes", "--force"}, additionalRequiredContainers...)
	internal.RunCommand(undeployAdditionalSystemCommand, exec.Command("docker", subCommand...))
}

func init() {
	rootCmd.AddCommand(undeployAdditionalSystemCmd)
}
