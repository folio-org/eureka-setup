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

	"github.com/j011195/eureka-setup/eureka-cli/action"
	"github.com/j011195/eureka-setup/eureka-cli/constant"
	"github.com/j011195/eureka-setup/eureka-cli/helpers"
	"github.com/spf13/cobra"
)

// deployAdditionalSystemCmd represents the deployAdditionalSystem command
var deployAdditionalSystemCmd = &cobra.Command{
	Use:   "deployAdditionalSystem",
	Short: "Deploy additional system",
	Long:  `Deploy additional system containers.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		run, err := New(action.DeployAdditionalSystem)
		if err != nil {
			return err
		}

		return run.DeployAdditionalSystem()
	},
}

func (run *Run) DeployAdditionalSystem() error {
	slog.Info(run.Config.Action.Name, "text", "DEPLOYING ADDITIONAL SYSTEM CONTAINERS")
	finalRequiredContainers := helpers.AppendRequiredContainers(run.Config.Action.Name, []string{}, run.Config.Action.ConfigBackendModules)
	if len(finalRequiredContainers) == 0 {
		slog.Info(run.Config.Action.Name, "text", "No additional system containers deployed")
		return nil
	}

	homeDir, err := helpers.GetHomeMiscDir()
	if err != nil {
		return err
	}

	subCommand := append([]string{"compose", "--progress", "plain", "--ansi", "never", "--project-name", "eureka", "up", "--detach"}, finalRequiredContainers...)
	if err := run.Config.ExecSvc.ExecFromDir(exec.Command("docker", subCommand...), homeDir); err != nil {
		return err
	}
	slog.Info(run.Config.Action.Name, "text", "WAITING FOR ADDITIONAL SYSTEM CONTAINERS TO BECOME READY")
	time.Sleep(constant.DeployAdditionalSystemWait)
	slog.Info(run.Config.Action.Name, "text", "All additional system containers are ready")

	return nil
}

func init() {
	rootCmd.AddCommand(deployAdditionalSystemCmd)
}
