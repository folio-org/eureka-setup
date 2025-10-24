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

// undeploySystemCmd represents the undeploySystem command
var undeploySystemCmd = &cobra.Command{
	Use:   "undeploySystem",
	Short: "Undeploy system",
	Long:  `Undeploy all system containers.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := New(action.UndeploySystem)
		if err != nil {
			return err
		}

		return r.UndeploySystem()
	},
}

func (r *Run) UndeploySystem() error {
	slog.Info(r.Config.Action.Name, "text", "UNDEPLOYING SYSTEM CONTAINERS")
	preparedCommand := exec.Command("docker", "compose", "--progress", "plain", "--ansi", "never", "--project-name", "eureka", "down", "--volumes", "--remove-orphans")
	return helpers.Exec(preparedCommand)
}

func init() {
	rootCmd.AddCommand(undeploySystemCmd)
}
