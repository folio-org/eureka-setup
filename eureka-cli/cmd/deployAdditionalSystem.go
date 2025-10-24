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
	"time"

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/spf13/cobra"
)

// deployAdditionalSystemCmd represents the deployAdditionalSystem command
var deployAdditionalSystemCmd = &cobra.Command{
	Use:   "deployAdditionalSystem",
	Short: "Deploy additional system",
	Long:  `Deploy additional system containers.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := New(action.DeployAdditionalSystem)
		if err != nil {
			return err
		}

		return r.DeployAdditionalSystem()
	},
}

func (r *Run) DeployAdditionalSystem() error {
	slog.Info(r.Config.Action.Name, "text", "DEPLOYING ADDITIONAL SYSTEM CONTAINERS")

	finalRequiredContainers := helpers.AppendAdditionalRequiredContainers(r.Config.Action, []string{})
	if len(finalRequiredContainers) == 0 {
		slog.Info(r.Config.Action.Name, "text", "No additional system containers deployed")
		return nil
	}

	dir, err := helpers.GetHomeMiscDir(r.Config.Action)
	if err != nil {
		return err
	}

	subCommand := append([]string{"compose", "--progress", "plain", "--ansi", "never", "--project-name", "eureka", "up", "--detach"}, finalRequiredContainers...)
	err = helpers.ExecFromDir(exec.Command("docker", subCommand...), dir)
	if err != nil {
		return err
	}

	slog.Info(r.Config.Action.Name, "text", "WAITING FOR ADDITIONAL SYSTEM CONTAINERS TO BECOME READY")
	time.Sleep(constant.DeployAdditionalSystemWait)
	slog.Info(r.Config.Action.Name, "text", "All additional system containers are ready")

	return nil
}

func init() {
	rootCmd.AddCommand(deployAdditionalSystemCmd)
}
