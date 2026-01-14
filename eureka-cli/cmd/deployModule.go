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
	"github.com/j011195/eureka-setup/eureka-cli/action"
	"github.com/spf13/cobra"
)

// deployModuleCmd represents the deployModule command
var deployModuleCmd = &cobra.Command{
	Use:   "deployModule",
	Short: "Deploy module",
	Long:  `Deploy a single module.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		run, err := New(action.DeployModule)
		if err != nil {
			return err
		}

		return run.DeployModule()
	},
}

func (run *Run) DeployModule() error {
	return nil
}

func init() {
	rootCmd.AddCommand(deployModuleCmd)
}
