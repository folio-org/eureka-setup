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
		r, err := New(action.DeploySystem)
		if err != nil {
			return err
		}

		return r.DeploySystem()
	},
}

func (r *Run) DeploySystem() error {
	err := r.CloneUpdateRepositories()
	if err != nil {
		return err
	}

	if ap.BuildImages {
		err := r.BuildSystem()
		if err != nil {
			return err
		}
	}

	subCommand := []string{"compose", "--progress", "plain", "--ansi", "never", "--project-name", "eureka", "up", "--detach"}
	if ap.OnlyRequired {
		initialRequiredContainers := constant.GetInitialRequiredContainers()
		finalRequiredContainers := helpers.AppendAdditionalRequiredContainers(r.Config.Action, initialRequiredContainers)
		subCommand = append(subCommand, finalRequiredContainers...)
	}

	slog.Info(r.Config.Action.Name, "text", "DEPLOYING SYSTEM CONTAINERS")
	dir, err := helpers.GetHomeMiscDir(r.Config.Action)
	if err != nil {
		return err
	}

	err = helpers.ExecFromDir(exec.Command("docker", subCommand...), dir)
	if err != nil {
		return err
	}

	slog.Info(r.Config.Action.Name, "text", "WAITING FOR SYSTEM CONTAINERS TO BECOME READY")
	time.Sleep(constant.DeploySystemWait)
	slog.Info(r.Config.Action.Name, "text", "All system containers are ready")

	return nil
}

func init() {
	rootCmd.AddCommand(deploySystemCmd)
	deploySystemCmd.PersistentFlags().BoolVarP(&ap.BuildImages, "buildImages", "b", false, "Build Docker images")
	deploySystemCmd.PersistentFlags().BoolVarP(&ap.UpdateCloned, "updateCloned", "u", false, "Update Git cloned projects")
	deploySystemCmd.PersistentFlags().BoolVarP(&ap.OnlyRequired, "onlyRequired", "R", false, "Use only required system containers")
}
