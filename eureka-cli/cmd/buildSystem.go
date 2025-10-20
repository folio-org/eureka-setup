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
	"time"

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/gitclient"
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

		r := NewRun(action.BuildSystem)
		r.CloneUpdateRepositories()
		err := r.BuildSystem()
		if err != nil {
			return err
		}

		slog.Info(r.Config.Action.Name, "text", fmt.Sprintf("Elapsed, duration %.1f", time.Since(start).Minutes()))

		return nil
	},
}

func (r *Run) CloneUpdateRepositories() {
	slog.Info(r.Config.Action.Name, "text", "CLONING & UPDATING REPOSITORIES")
	repositories := []*gitclient.GitRepository{
		r.Config.GitClient.KongRepository(),
		r.Config.GitClient.KeycloakRepository(),
	}

	slog.Info(r.Config.Action.Name, "text", "Cloning repositories")
	for _, repository := range repositories {
		r.Config.GitClient.Clone(false, repository)
	}

	if rp.UpdateCloned {
		slog.Info(r.Config.Action.Name, "text", "Updating repositories")
		for _, repository := range repositories {
			r.Config.GitClient.ResetHardPullFromOrigin(repository)
		}
	}
}

func (r *Run) BuildSystem() error {
	slog.Info(r.Config.Action.Name, "text", "BUILDING SYSTEM IMAGES")
	subCommand := []string{"compose", "--progress", "plain", "--ansi", "never", "--project-name", "eureka", "build", "--no-cache"}
	return helpers.ExecFromDir(exec.Command("docker", subCommand...), helpers.GetHomeMiscDir(r.Config.Action))
}

func init() {
	rootCmd.AddCommand(buildSystemCmd)
	buildSystemCmd.PersistentFlags().BoolVarP(&rp.UpdateCloned, "updateCloned", "u", false, "Update Git cloned projects")
}
