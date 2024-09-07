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
	"log/slog"
	"os/exec"
	"time"

	"github.com/folio-org/eureka-cli/internal"
	"github.com/spf13/cobra"
)

const deploySystemCommand string = "Deploy System"

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
	slog.Info(deploySystemCommand, "### DEPLOYING SYSTEM CONTAINERS ###", "")
	preparedCommands := []*exec.Cmd{
		// TODO Add an optional --no-cache flag
		exec.Command("docker", "compose", "-p", "eureka", "build", "--no-cache"),
		exec.Command("docker", "compose", "-p", "eureka", "up", "--detach"),
	}
	for _, preparedCommand := range preparedCommands {
		internal.RunCommandFromDir(deployManagementCommand, preparedCommand, internal.ComposeWorkDir)
	}
	slog.Info(deploySystemCommand, "### WAITING FOR SYSTEM TO INITIALIZE ###", "")
	time.Sleep(150 * time.Second)
}

func init() {
	rootCmd.AddCommand(deploySystemCmd)
}
