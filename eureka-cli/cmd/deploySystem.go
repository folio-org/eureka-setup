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

const deploySystemCommand string = "Deploy System"

var coreRequiredContainers = []string{"postgres", "kafka", "kafka-tools", "vault", "keycloak", "keycloak-internal", "kong"}

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
	CloneUpdateRepositories()
	if withBuildImages {
		BuildSystem()
	}

	subCommand := []string{"compose", "--progress", "plain", "--ansi", "never", "--project-name", "eureka", "up", "--detach"}
	if withOnlyRequired {
		requiredContainers := internal.GetRequiredContainers(deploySystemCommand, coreRequiredContainers)
		subCommand = append(subCommand, requiredContainers...)
	}

	slog.Info(deploySystemCommand, internal.GetFuncName(), "### DEPLOYING SYSTEM CONTAINERS ###")
	internal.RunCommandFromDir(deploySystemCommand, exec.Command("docker", subCommand...), internal.GetHomeMiscDir(deploySystemCommand))

	slog.Info(deploySystemCommand, internal.GetFuncName(), "### WAITING FOR SYSTEM CONTAINERS TO INITIALIZE ###")
	time.Sleep(15 * time.Second)
	slog.Info(deploySystemCommand, internal.GetFuncName(), "All system containers have initialized")
}

func init() {
	rootCmd.AddCommand(deploySystemCmd)
	deploySystemCmd.PersistentFlags().BoolVarP(&withBuildImages, "buildImages", "b", false, "Build Docker images")
	deploySystemCmd.PersistentFlags().BoolVarP(&withUpdateCloned, "updateCloned", "u", false, "Update Git cloned projects")
	deploySystemCmd.PersistentFlags().BoolVarP(&withOnlyRequired, "onlyRequired", "R", false, "Use only required system containers")
}
