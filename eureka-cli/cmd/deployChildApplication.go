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
	"time"

	"github.com/spf13/cobra"
)

const deployChildApplicationCommand string = "Deploy Child Application"

// deployChildApplicationCmd represents the deployChildApplication command
var deployChildApplicationCmd = &cobra.Command{
	Use:   "deployChildApplication",
	Short: "Deploy child application",
	Long:  `Deploy platform child application.`,
	Run: func(cmd *cobra.Command, args []string) {
		DeployChildApplication()
	},
}

func DeployChildApplication() {
	start := time.Now()
	DeployModules()
	CreateTenantEntitlements()
	DetachCapabilitySets()
	AttachCapabilitySets()
	slog.Info(deployChildApplicationCommand, "Elapsed, duration", time.Since(start))
}

func init() {
	rootCmd.AddCommand(deployChildApplicationCmd)
}
