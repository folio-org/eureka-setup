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

	"github.com/folio-org/eureka-cli/internal"
	"github.com/spf13/cobra"
)

const listSystemCommand string = "List System"

// listSystemCmd represents the listSystem command
var listSystemCmd = &cobra.Command{
	Use:   "listSystem",
	Short: "List system containers",
	Long:  `List all system containers.`,
	Run: func(cmd *cobra.Command, args []string) {
		ListSystem()
	},
}

func ListSystem() {
	slog.Info(listSystemCommand, internal.GetFuncName(), "### LISTING SYSTEM CONTAINERS ###")
	internal.RunCommand(listSystemCommand, exec.Command("docker", "compose", "--project-name", "eureka", "ps", "--all"))
}

func init() {
	rootCmd.AddCommand(listSystemCmd)
}
