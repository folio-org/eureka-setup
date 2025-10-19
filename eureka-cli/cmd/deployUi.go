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

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/folio-org/eureka-cli/tenanttype"
	"github.com/spf13/cobra"
)

// deployUiCmd represents the deployUi command
var deployUiCmd = &cobra.Command{
	Use:   "deployUi",
	Short: "Deploy UI",
	Long:  `Deploy the UI container.`,
	Run: func(cmd *cobra.Command, args []string) {
		NewRun(action.DeployUi).DeployUi()
	},
}

func (r *Run) DeployUi() {
	slog.Info(r.Config.Action.Name, "text", "DEPLOYING UI")

	for _, value := range r.Config.ManagementStep.GetTenants(false, constant.NoneConsortium, tenanttype.All) {
		existingTenant := value.(map[string]any)["name"].(string)
		if !helpers.HasTenant(existingTenant) || !helpers.IsUIEnabled(existingTenant) {
			continue
		}

		r.Config.TenantStep.SetDefaultConfigTenantParams(&rp, existingTenant)

		finalImageName := r.Config.UIStep.PrepareUIImage(&rp, existingTenant)
		externalPort := helpers.ExtractPortFromURL(r.Config.Action, rp.PlatformCompleteURL)
		r.Config.UIStep.DeployContainer(existingTenant, finalImageName, externalPort)
	}
}

func init() {
	rootCmd.AddCommand(deployUiCmd)
	deployUiCmd.PersistentFlags().BoolVarP(&rp.BuildImages, "buildImages", "b", false, "Build Docker images")
	deployUiCmd.PersistentFlags().BoolVarP(&rp.UpdateCloned, "updateCloned", "u", false, "Update Git cloned projects")
	deployUiCmd.PersistentFlags().BoolVarP(&rp.SingleTenant, "singleTenant", "T", true, "Use for Single Tenant workflow")
	deployUiCmd.PersistentFlags().BoolVarP(&rp.EnableECSRequests, "enableEcsRequests", "e", false, "Enable ECS requests")
}
