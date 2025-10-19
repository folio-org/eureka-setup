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
	"os/exec"

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/spf13/cobra"
)

// undeployAdditionalSystemCmd represents the undeployAdditionalSystem command
var undeployAdditionalSystemCmd = &cobra.Command{
	Use:   "undeployAdditionalSystem",
	Short: "Undeploy additional system",
	Long:  `Undeploy additional system containers.`,
	Run: func(cmd *cobra.Command, args []string) {
		NewRun(action.UndeployAdditionalSystem).UndeployAdditionalSystem()
	},
}

func (r *Run) UndeployAdditionalSystem() {
	slog.Info(r.Config.Action.Name, "text", "UNDEPLOYING ADDITIONAL SYSTEM CONTAINERS")

	finalRequiredContainers := helpers.AppendAdditionalRequiredContainers(r.Config.Action, []string{})
	if len(finalRequiredContainers) == 0 {
		slog.Info(r.Config.Action.Name, "text", "No additional system containers undeployed")
		return
	}

	subCommand := append([]string{"compose", "--progress", "plain", "--ansi", "never", "--project-name", "eureka", "stop"}, finalRequiredContainers...)
	helpers.Exec(exec.Command("docker", subCommand...))

	subCommand = append([]string{"compose", "--progress", "plain", "--ansi", "never", "--project-name", "eureka", "rm", "--volumes", "--force"}, finalRequiredContainers...)
	helpers.Exec(exec.Command("docker", subCommand...))
}

func init() {
	rootCmd.AddCommand(undeployAdditionalSystemCmd)
}
