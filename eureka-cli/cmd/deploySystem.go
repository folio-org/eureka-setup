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
	"strings"
	"time"

	"github.com/folio-org/eureka-setup/eureka-cli/action"
	"github.com/folio-org/eureka-setup/eureka-cli/constant"
	"github.com/folio-org/eureka-setup/eureka-cli/helpers"
	"github.com/spf13/cobra"
)

// deploySystemCmd represents the deploySystem command
var deploySystemCmd = &cobra.Command{
	Use:   "deploySystem",
	Short: "Deploy system",
	Long:  `Deploy all system containers.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		run, err := New(action.DeploySystem)
		if err != nil {
			return err
		}

		return run.DeploySystem()
	},
}

func (run *Run) DeploySystem() error {
	slog.Info(run.Config.Action.Name, "text", "DEPLOYING SYSTEM CONTAINERS")
	if err := run.CloneUpdateRepositories(); err != nil {
		return err
	}
	if params.BuildImages {
		if err := run.BuildSystem(); err != nil {
			return err
		}
	}

	subCommand := []string{"compose", "--progress", "plain", "--ansi", "never", "--project-name", "eureka", "up", "--detach"}
	if params.OnlyRequired {
		initialRequiredContainers := constant.GetInitialRequiredContainers()
		finalRequiredContainers := helpers.AppendRequiredContainers(run.Config.Action.Name, initialRequiredContainers, run.Config.Action.ConfigBackendModules)
		subCommand = append(subCommand, finalRequiredContainers...)
	}

	return run.dockerComposeUp(subCommand, constant.DeploySystemWait, "system")
}

func (run *Run) dockerComposeUp(subCommand []string, wait time.Duration, label string) error {
	homeDir, err := helpers.GetHomeMiscDir()
	if err != nil {
		return err
	}
	dockerCmd := exec.Command("docker", subCommand...)
	dockerCmd.Dir = homeDir

	stdout, stderr, err := run.Config.ExecSvc.ExecReturnOutput(dockerCmd)
	if err != nil {
		return err
	}

	combined := stdout.String() + stderr.String()
	if strings.Contains(combined, " Started") || strings.Contains(combined, " Created") {
		slog.Info(run.Config.Action.Name, "text", "WAITING FOR "+strings.ToUpper(label)+" CONTAINERS TO BECOME READY")
		time.Sleep(wait)
		slog.Info(run.Config.Action.Name, "text", fmt.Sprintf("All %s containers are ready", label))
	} else {
		slog.Info(run.Config.Action.Name, "text", fmt.Sprintf("All %s containers already running, skipping wait", label))
	}

	return nil
}

func init() {
	rootCmd.AddCommand(deploySystemCmd)
	deploySystemCmd.PersistentFlags().BoolVarP(&params.BuildImages, action.BuildImages.Long, action.BuildImages.Short, false, action.BuildImages.Description)
	deploySystemCmd.PersistentFlags().BoolVarP(&params.UpdateCloned, action.UpdateCloned.Long, action.UpdateCloned.Short, false, action.UpdateCloned.Description)
	deploySystemCmd.PersistentFlags().BoolVarP(&params.OnlyRequired, action.OnlyRequired.Long, action.OnlyRequired.Short, false, action.OnlyRequired.Description)
}
