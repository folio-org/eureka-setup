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
	"github.com/folio-org/eureka-cli/gitrepository"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/spf13/cobra"
)

// buildSystemCmd represents the buildSystem command
var buildSystemCmd = &cobra.Command{
	Use:   "buildSystem",
	Short: "Build system",
	Long:  `Build system images.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		start := time.Now()

		r, err := New(action.BuildSystem)
		if err != nil {
			return err
		}

		err = r.CloneUpdateRepositories()
		if err != nil {
			return err
		}
		err = r.BuildSystem()
		if err != nil {
			return err
		}
		helpers.LogCompletion(r.RunConfig.Action.Name, start)

		return nil
	},
}

func (r *Run) CloneUpdateRepositories() error {
	slog.Info(r.RunConfig.Action.Name, "text", "CLONING & UPDATING REPOSITORIES")

	kongRepository, err := r.RunConfig.GitClient.KongRepository()
	if err != nil {
		return err
	}

	keycloakRepository, err := r.RunConfig.GitClient.KeycloakRepository()
	if err != nil {
		return err
	}

	repositories := []*gitrepository.GitRepository{kongRepository, keycloakRepository}

	slog.Info(r.RunConfig.Action.Name, "text", "Cloning repositories", "repositories", repositories)
	for _, repository := range repositories {
		_ = r.RunConfig.GitClient.Clone(repository)
	}

	if actionParams.UpdateCloned {
		slog.Info(r.RunConfig.Action.Name, "text", "Updating repositories", "repositories", repositories)
		for _, repository := range repositories {
			err = r.RunConfig.GitClient.ResetHardPullFromOrigin(repository)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *Run) BuildSystem() error {
	slog.Info(r.RunConfig.Action.Name, "text", "BUILDING SYSTEM IMAGES")
	subCommand := []string{"compose", "--progress", "plain", "--ansi", "never", "--project-name", "eureka", "build", "--no-cache"}
	dir, err := helpers.GetHomeMiscDir(r.RunConfig.Action.Name)
	if err != nil {
		return err
	}

	return helpers.ExecFromDir(exec.Command("docker", subCommand...), dir)
}

func init() {
	rootCmd.AddCommand(buildSystemCmd)
	buildSystemCmd.PersistentFlags().BoolVarP(&actionParams.UpdateCloned, "updateCloned", "u", false, "Update Git cloned projects")
}
