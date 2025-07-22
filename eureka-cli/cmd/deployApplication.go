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
	"fmt"
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
		start := time.Now()
		if len(viper.GetStringMap(internal.ApplicationGatewayDependenciesKey)) > 0 {
			DeployChildApplication()
		} else {
			DeployApplication()
		}
		slog.Info(deployApplicationCommand, "Elapsed, duration", time.Since(start))
	},
}

func DeployApplication() {
	DeploySystem()
	DeployManagement()
	DeployModules()
	CreateTenants()
	RunByConsortiumAndTenantType(deployApplicationCommand, func(consortium string, tenantType internal.TenantType) {
		waitDuration := 10 * time.Second

		CreateTenantEntitlements(consortium, tenantType)
		CreateRoles(consortium, tenantType)
		CreateUsers(consortium, tenantType)
		AttachCapabilitySets(consortium, tenantType, waitDuration)

		if consortium != internal.NoneConsortium {
			slog.Info(deployApplicationCommand, internal.GetFuncName(), fmt.Sprintf("Waiting for %d duration", waitDuration))
			time.Sleep(waitDuration)
		}
	})
	CreateConsortium()
	DeployUi()
	UpdateKeycloakPublicClients()
	if internal.HasModule(internal.ModSearchModuleName) {
		RunByConsortiumAndTenantType(deployApplicationCommand, func(consortium string, tenantType internal.TenantType) {
			ReindexElasticsearch(consortium, tenantType)
		})
	}
}

func DeployChildApplication() {
	DeployAdditionalSystem()
	DeployModules()
	RunByConsortiumAndTenantType(deployApplicationCommand, func(consortium string, tenantType internal.TenantType) {
		CreateTenantEntitlements(consortium, tenantType)
		DetachCapabilitySets(consortium, tenantType)
		AttachCapabilitySets(consortium, tenantType, 0*time.Second)
	})
}

func init() {
	rootCmd.AddCommand(deployApplicationCmd)
	deployApplicationCmd.PersistentFlags().BoolVarP(&withBuildImages, "buildImages", "b", false, "Build Docker images")
	deployApplicationCmd.PersistentFlags().BoolVarP(&withUpdateCloned, "updateCloned", "u", false, "Update Git cloned projects")
	deployApplicationCmd.PersistentFlags().BoolVarP(&withOnlyRequired, "onlyRequired", "R", false, "Use only required system containers")
}
