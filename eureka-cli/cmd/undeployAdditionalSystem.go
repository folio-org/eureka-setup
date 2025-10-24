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
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := New(action.UndeployAdditionalSystem)
		if err != nil {
			return err
		}

		return r.UndeployAdditionalSystem()
	},
}

func (r *Run) UndeployAdditionalSystem() error {
	slog.Info(r.Config.Action.Name, "text", "UNDEPLOYING ADDITIONAL SYSTEM CONTAINERS")

	finalRequiredContainers := helpers.AppendAdditionalRequiredContainers(r.Config.Action, []string{})
	if len(finalRequiredContainers) == 0 {
		slog.Info(r.Config.Action.Name, "text", "No additional system containers undeployed")
		return nil
	}

	subCommand := append([]string{"compose", "--progress", "plain", "--ansi", "never", "--project-name", "eureka", "stop"}, finalRequiredContainers...)
	err := helpers.Exec(exec.Command("docker", subCommand...))
	if err != nil {
		return err
	}

	subCommand = append([]string{"compose", "--progress", "plain", "--ansi", "never", "--project-name", "eureka", "rm", "--volumes", "--force"}, finalRequiredContainers...)
	err = helpers.Exec(exec.Command("docker", subCommand...))
	if err != nil {
		return err
	}

	return nil
}

func init() {
	rootCmd.AddCommand(undeployAdditionalSystemCmd)
}
