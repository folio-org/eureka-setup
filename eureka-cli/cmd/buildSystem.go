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

	"github.com/folio-org/eureka-cli/internal"
	"github.com/spf13/cobra"
)

const (
	buildSystemCommand string = "Build System"
)

// buildSystemCmd represents the buildSystem command
var buildSystemCmd = &cobra.Command{
	Use:   "buildSystem",
	Short: "Build system",
	Long:  `Build system images.`,
	Run: func(cmd *cobra.Command, args []string) {
		start := time.Now()
		CloneUpdateRepositories()
		BuildSystem()
		slog.Info(buildSystemCommand, "Elapsed, duration", time.Since(start))
	},
}

func CloneUpdateRepositories() {
	slog.Info(buildSystemCommand, internal.GetFuncName(), "### CLONING & UPDATING REPOSITORIES ###")
	repositories := []*internal.Repository{
		internal.NewRepository(buildSystemCommand, internal.FolioKongRepositoryUrl, internal.DefaultFolioKongOutputDir, internal.DefaultFolioKongBranchName),
		internal.NewRepository(buildSystemCommand, internal.FolioKeycloakRepositoryUrl, internal.DefaultFolioKeycloakOutputDir, internal.DefaultFolioKeycloakBranchName),
	}

	slog.Info(buildSystemCommand, internal.GetFuncName(), "Cloning repositories")
	for _, repository := range repositories {
		internal.GitCloneRepository(buildSystemCommand, withEnableDebug, false, repository)
	}

	if withUpdateCloned {
		slog.Info(buildSystemCommand, internal.GetFuncName(), "Updating repositories")
		for _, repository := range repositories {
			internal.GitResetHardPullFromOriginRepository(buildSystemCommand, withEnableDebug, repository)
		}
	}
}

func BuildSystem() {
	slog.Info(buildSystemCommand, internal.GetFuncName(), "### BUILDING SYSTEM IMAGES ###")
	subCommand := []string{"compose", "--progress", "plain", "--ansi", "never", "--project-name", "eureka", "build", "--no-cache"}
	internal.RunCommandFromDir(buildSystemCommand, exec.Command("docker", subCommand...), internal.GetHomeMiscDir(buildSystemCommand))
}

func init() {
	rootCmd.AddCommand(buildSystemCmd)
	buildSystemCmd.PersistentFlags().BoolVarP(&withUpdateCloned, "updateCloned", "u", false, "Update Git cloned projects")
}
