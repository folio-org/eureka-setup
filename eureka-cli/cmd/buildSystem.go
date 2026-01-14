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
	"github.com/j011195/eureka-setup/eureka-cli/gitrepository"
	"github.com/j011195/eureka-setup/eureka-cli/helpers"
	"github.com/spf13/cobra"
)

// buildSystemCmd represents the buildSystem command
var buildSystemCmd = &cobra.Command{
	Use:   "buildSystem",
	Short: "Build system",
	Long:  `Build system images.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		start := time.Now()

		run, err := New(action.BuildSystem)
		if err != nil {
			return err
		}
		if err := run.CloneUpdateRepositories(); err != nil {
			return err
		}
		if err := run.BuildSystem(); err != nil {
			return err
		}
		slog.Info(run.Config.Action.Name, "text", "Command completed", "duration", time.Since(start))

		return nil
	},
}

func (run *Run) CloneUpdateRepositories() error {
	slog.Info(run.Config.Action.Name, "text", "CLONING & UPDATING REPOSITORIES")
	kongRepository, err := run.Config.GitClient.KongRepository()
	if err != nil {
		return err
	}

	keycloakRepository, err := run.Config.GitClient.KeycloakRepository()
	if err != nil {
		return err
	}

	repositories := []*gitrepository.GitRepository{kongRepository, keycloakRepository}
	slog.Info(run.Config.Action.Name, "text", "Cloning repositories", "repositories", repositories)
	for _, repository := range repositories {
		if err := run.Config.GitClient.Clone(repository); err != nil {
			slog.Warn(run.Config.Action.Name, "text", "Cloning was unsuccessful", "error", err)
		}
	}

	if params.UpdateCloned {
		slog.Info(run.Config.Action.Name, "text", "Updating repositories", "repositories", repositories)
		for _, repository := range repositories {
			if err := run.Config.GitClient.ResetHardPullFromOrigin(repository); err != nil {
				return err
			}
		}
	}

	return nil
}

func (run *Run) BuildSystem() error {
	slog.Info(run.Config.Action.Name, "text", "BUILDING SYSTEM IMAGES")
	subCommand := []string{"compose", "--progress", "plain", "--ansi", "never", "--project-name", "eureka", "build", "--no-cache"}
	homeDir, err := helpers.GetHomeMiscDir()
	if err != nil {
		return err
	}

	return run.Config.ExecSvc.ExecFromDir(exec.Command("docker", subCommand...), homeDir)
}

func init() {
	rootCmd.AddCommand(buildSystemCmd)
	buildSystemCmd.PersistentFlags().BoolVarP(&params.UpdateCloned, action.UpdateCloned.Long, action.UpdateCloned.Short, false, action.UpdateCloned.Description)
}
