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
	"os/exec"

	"github.com/j011195/eureka-setup/eureka-cli/action"
	"github.com/spf13/cobra"
)

// listSystemCmd represents the listSystem command
var listSystemCmd = &cobra.Command{
	Use:   "listSystem",
	Short: "List system containers",
	Long:  `List all system containers.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		run, err := New(action.ListSystem)
		if err != nil {
			return err
		}

		return run.ListSystem()
	},
}

func (run *Run) ListSystem() error {
	return run.Config.ExecSvc.Exec(exec.Command("docker", "compose", "--project-name", "eureka", "ps", "--all"))
}

func init() {
	rootCmd.AddCommand(listSystemCmd)
}
