/*
Copyright © 2025 EPAM_Systems/Thunderjet/Boburbek_Kadirkhodjaev

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

const deployAdditionalSystemCommand string = "Deploy Additional System"

// deployAdditionalSystemCmd represents the deployAdditionalSystem command
var deployAdditionalSystemCmd = &cobra.Command{
	Use:   "deployAdditionalSystem",
	Short: "Deploy additional system",
	Long:  `Deploy additional system containers.`,
	Run: func(cmd *cobra.Command, args []string) {
		DeployAdditionalSystem()
	},
}

func DeployAdditionalSystem() {
	slog.Info(deployAdditionalSystemCommand, internal.GetFuncName(), "### DEPLOYING ADDITIONAL SYSTEM CONTAINERS ###")

	additionalRequiredContainers := internal.GetRequiredContainers(deployAdditionalSystemCommand, []string{})
	if len(additionalRequiredContainers) == 0 {
		slog.Info(deployAdditionalSystemCommand, internal.GetFuncName(), "No additional system containers deployed")
		return
	}

	subCommand := append([]string{"compose", "--progress", "plain", "--ansi", "never", "--project-name", "eureka", "up", "--detach"}, additionalRequiredContainers...)
	internal.RunCommandFromDir(deployAdditionalSystemCommand, exec.Command("docker", subCommand...), internal.GetHomeMiscDir(deployAdditionalSystemCommand))

	slog.Info(deployAdditionalSystemCommand, internal.GetFuncName(), "### WAITING FOR SYSTEM TO INITIALIZE ###")
	time.Sleep(15 * time.Second)
	slog.Info(deployAdditionalSystemCommand, internal.GetFuncName(), "All system containers have initialized")
}

func init() {
	rootCmd.AddCommand(deployAdditionalSystemCmd)
}
