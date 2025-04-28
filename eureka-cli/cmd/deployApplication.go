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
	"time"

	"github.com/folio-org/eureka-cli/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const deployApplicationCommand string = "Deploy Application"

// deployApplicationCmd represents the deployApplication command
var deployApplicationCmd = &cobra.Command{
	Use:   "deployApplication",
	Short: "Deploy application",
	Long:  `Deploy platform application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(viper.GetStringMap(internal.ApplicationGatewayDependenciesKey)) > 0 {
			DeployChildApplication()
			return
		}
		DeployApplication()
	},
}

func DeployApplication() {
	start := time.Now()
	DeploySystem()
	DeployManagement()
	DeployModules()
	CreateTenants()
	CreateTenantEntitlements()
	CreateRoles()
	CreateUsers()
	AttachCapabilitySets()
	DeployUi()
	slog.Info(deployApplicationCommand, "Elapsed, duration", time.Since(start))
}

func DeployChildApplication() {
	start := time.Now()
	DeployModules()
	CreateTenantEntitlements()
	DetachCapabilitySets()
	AttachCapabilitySets()
	slog.Info(deployApplicationCommand, "Elapsed, duration", time.Since(start))
}

func init() {
	rootCmd.AddCommand(deployApplicationCmd)
	deployApplicationCmd.PersistentFlags().BoolVarP(&buildImages, "buildImages", "b", false, "Build images")
	deployApplicationCmd.PersistentFlags().BoolVarP(&updateCloned, "updateCloned", "u", false, "Update cloned projects")
}
