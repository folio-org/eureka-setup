/*
Copyright © 2024 EPAM_Systems/Thunderjet/Boburbek_Kadirkhodjaev

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

	"github.com/folio-org/eureka-cli/internal"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/spf13/cobra"
)

const (
	deploySystemCommand string = "Deploy System"
	folioKeycloakDir    string = "folio-keycloak"
	folioKongDir        string = "folio-kong"

	masterBranchName plumbing.ReferenceName = "master"
)

// deploySystemCmd represents the deploySystem command
var deploySystemCmd = &cobra.Command{
	Use:   "deploySystem",
	Short: "Undeploy system",
	Long:  `Undeploy all system containers.`,
	Run: func(cmd *cobra.Command, args []string) {
		DeploySystem()
	},
}

func DeploySystem() {
	folioKeycloakOutputDir := fmt.Sprintf("%s/%s", internal.DockerComposeWorkDir, folioKeycloakDir)
	folioKongOutputDir := fmt.Sprintf("%s/%s", internal.DockerComposeWorkDir, folioKongDir)

	slog.Info(deploySystemCommand, "### CLONING & UPDATING SYSTEM COMPONENTS ###", "")

	slog.Info(deploySystemCommand, fmt.Sprintf("Cloning %s from a %s branch", folioKeycloakDir, masterBranchName), "")
	internal.GitCloneRepository(deploySystemCommand, enableDebug, internal.FolioKeycloakRepositoryUrl, masterBranchName, folioKeycloakOutputDir, false)

	slog.Info(deploySystemCommand, fmt.Sprintf("Pulling updates for %s from origin", folioKeycloakDir), "")
	internal.GitResetHardPullFromOriginRepository(deploySystemCommand, enableDebug, internal.FolioKeycloakRepositoryUrl, masterBranchName, folioKeycloakOutputDir)

	slog.Info(deploySystemCommand, fmt.Sprintf("Cloning %s from a %s branch", folioKongDir, masterBranchName), "")
	internal.GitCloneRepository(deploySystemCommand, enableDebug, internal.FolioKongRepositoryUrl, masterBranchName, folioKongOutputDir, false)

	slog.Info(deploySystemCommand, fmt.Sprintf("Pulling updates for %s from origin", folioKongDir), "")
	internal.GitResetHardPullFromOriginRepository(deploySystemCommand, enableDebug, internal.FolioKongRepositoryUrl, masterBranchName, folioKongOutputDir)

	slog.Info(deploySystemCommand, "### DEPLOYING SYSTEM CONTAINERS ###", "")
	// TODO Add an optional --no-cache flag
	preparedCommands := []*exec.Cmd{exec.Command("docker", "compose", "-p", "eureka", "build", "--no-cache"),
		exec.Command("docker", "compose", "-p", "eureka", "up", "--detach"),
	}
	for _, preparedCommand := range preparedCommands {
		internal.RunCommandFromDir(deployManagementCommand, preparedCommand, internal.DockerComposeWorkDir)
	}
	slog.Info(deploySystemCommand, "### WAITING FOR SYSTEM TO INITIALIZE ###", "")
	time.Sleep(15 * time.Second)
	slog.Info(deployModulesCommand, "System has initialized", "")
}

func init() {
	rootCmd.AddCommand(deploySystemCmd)
}
