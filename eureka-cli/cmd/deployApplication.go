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

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/field"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/folio-org/eureka-cli/tenanttype"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// deployApplicationCmd represents the deployApplication command
var deployApplicationCmd = &cobra.Command{
	Use:   "deployApplication",
	Short: "Deploy application",
	Long:  `Deploy platform application.`,
	Run: func(cmd *cobra.Command, args []string) {
		start := time.Now()

		r := NewRun(action.DeployApplication)
		if len(viper.GetStringMap(field.ApplicationGatewayDependencies)) > 0 {
			r.DeployChildApplication()
		} else {
			r.DeployApplication()
		}

		slog.Info(r.Config.Action.Name, "text", fmt.Sprintf("Elapsed, duration %.1f", time.Since(start).Minutes()))
	},
}

func (r *Run) DeployApplication() {
	r.DeploySystem()
	r.DeployManagement()
	r.DeployModules()
	r.CreateTenants()
	r.PartitionByConsortiumAndTenantType(func(consortiumName string, tenantType tenanttype.TenantType) {
		waitDuration := 10 * time.Second
		r.CreateTenantEntitlements(consortiumName, tenantType)
		r.CreateRoles(consortiumName, tenantType)
		r.CreateUsers(consortiumName, tenantType)
		r.AttachCapabilitySets(consortiumName, tenantType, waitDuration)
		if consortiumName != constant.NoneConsortium {
			slog.Info(r.Config.Action.Name, "text", fmt.Sprintf("Waiting for %.1f duration", waitDuration.Seconds()))
			time.Sleep(waitDuration)
		}
	})
	r.CreateConsortium()
	r.DeployUi()
	r.UpdateKeycloakPublicClients()
	if helpers.IsModuleEnabled(constant.ModSearchModuleName) {
		r.PartitionByConsortiumAndTenantType(func(consortiumName string, tenantType tenanttype.TenantType) {
			r.ReindexElasticsearch(consortiumName, tenantType)
		})
	}
}

func (r *Run) DeployChildApplication() {
	r.DeployAdditionalSystem()
	r.DeployModules()
	r.PartitionByConsortiumAndTenantType(func(consortiumName string, tenantType tenanttype.TenantType) {
		r.CreateTenantEntitlements(consortiumName, tenantType)
		r.DetachCapabilitySets(consortiumName, tenantType)
		r.AttachCapabilitySets(consortiumName, tenantType, 0*time.Second)
	})
}

func init() {
	rootCmd.AddCommand(deployApplicationCmd)
	deployApplicationCmd.PersistentFlags().BoolVarP(&rp.BuildImages, "buildImages", "b", false, "Build Docker images")
	deployApplicationCmd.PersistentFlags().BoolVarP(&rp.UpdateCloned, "updateCloned", "u", false, "Update Git cloned projects")
	deployApplicationCmd.PersistentFlags().BoolVarP(&rp.OnlyRequired, "onlyRequired", "R", false, "Use only required system containers")
}
