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

	"github.com/folio-org/eureka-setup/eureka-cli/action"
	"github.com/spf13/cobra"
)

// undeploySystemCmd represents the undeploySystem command
var undeploySystemCmd = &cobra.Command{
	Use:   "undeploySystem",
	Short: "Undeploy system",
	Long:  `Undeploy all system containers.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		run, err := New(action.UndeploySystem)
		if err != nil {
			return err
		}

		return run.UndeploySystem()
	},
}

func (run *Run) UndeploySystem() error {
	slog.Info(run.Config.Action.Name, "text", "UNDEPLOYING SYSTEM CONTAINERS")
	preparedCommand := exec.Command("docker", "compose", "--progress", "plain", "--ansi", "never", "--project-name", "eureka", "down", "--volumes", "--remove-orphans")
	return run.Config.ExecSvc.Exec(preparedCommand)
}

func init() {
	rootCmd.AddCommand(undeploySystemCmd)
}
