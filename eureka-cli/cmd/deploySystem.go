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
	if err := run.CloneUpdateRepositories(); err != nil {
		return err
	}
	if actionParams.BuildImages {
		if err := run.BuildSystem(); err != nil {
			return err
		}
	}

	subCommand := []string{"compose", "--progress", "plain", "--ansi", "never", "--project-name", "eureka", "up", "--detach"}
	if actionParams.OnlyRequired {
		initialRequiredContainers := constant.GetInitialRequiredContainers()
		finalRequiredContainers := helpers.AppendRequiredContainers(run.Config.Action.Name, initialRequiredContainers, run.Config.Action.ConfigBackendModules)
		subCommand = append(subCommand, finalRequiredContainers...)
	}

	slog.Info(run.Config.Action.Name, "text", "DEPLOYING SYSTEM CONTAINERS")
	dir, err := helpers.GetHomeMiscDir(run.Config.Action.Name)
	if err != nil {
		return err
	}
	if err := run.Config.ExecSvc.ExecFromDir(exec.Command("docker", subCommand...), dir); err != nil {
		return err
	}
	slog.Info(run.Config.Action.Name, "text", "WAITING FOR SYSTEM CONTAINERS TO BECOME READY")
	time.Sleep(constant.DeploySystemWait)
	slog.Info(run.Config.Action.Name, "text", "All system containers are ready")

	return nil
}

func init() {
	rootCmd.AddCommand(deploySystemCmd)
	deploySystemCmd.PersistentFlags().BoolVarP(&actionParams.BuildImages, "buildImages", "b", false, "Build Docker images")
	deploySystemCmd.PersistentFlags().BoolVarP(&actionParams.UpdateCloned, "updateCloned", "u", false, "Update Git cloned projects")
	deploySystemCmd.PersistentFlags().BoolVarP(&actionParams.OnlyRequired, "onlyRequired", "q", false, "Use only required system containers")
}
