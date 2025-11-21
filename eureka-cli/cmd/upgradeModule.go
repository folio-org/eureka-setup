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
	"github.com/folio-org/eureka-cli/action"
	"github.com/spf13/cobra"
)

// upgradeModuleCmd represents the upgradeModule command
var upgradeModuleCmd = &cobra.Command{
	Use:   "upgradeModule",
	Short: "Upgrade module",
	Long:  `Upgrade a single module for the current profile.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		run, err := New(action.UpgradeModule)
		if err != nil {
			return err
		}

		return run.UpgradeModule()
	},
}

// Module
//
//	Build jar
//	  mvn clean package -DskipTests
//	Build and tag docker image
//	  docker build --tag mod-orders:13.1.0-SNAPSHOT.1092 .
//	Deploy new container
//
// Management
//
//	Extract module descriptor from the target folder
//	Install module descriptor into the current application
//	Get current application and increment version n+1
//	Upgrade the current application to the new version
func (run *Run) UpgradeModule() error {
	return nil
}

func init() {
	rootCmd.AddCommand(upgradeModuleCmd)
}
