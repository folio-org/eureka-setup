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
	"github.com/spf13/cobra"
)

// deployUiCmd represents the deployUi command
var deployUiCmd = &cobra.Command{
	Use:   "deployUi",
	Short: "Deploy UI",
	Long:  `Deploy the UI container.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := New(action.DeployUi)
		if err != nil {
			return err
		}

		return r.DeployUi()
	},
}

func (r *Run) DeployUi() error {
	slog.Info(r.Config.Action.Name, "text", "DEPLOYING UI")

	tt, err := r.Config.ManagementSvc.GetTenants(constant.NoneConsortium, constant.All)
	if err != nil {
		return err
	}

	for _, value := range tt {
		existingTenant := value.(map[string]any)["name"].(string)
		if !helpers.HasTenant(existingTenant) || !helpers.IsUIEnabled(existingTenant) {
			continue
		}

		err := r.Config.TenantSvc.SetDefaultConfigTenantParams(existingTenant)
		if err != nil {
			return err
		}

		finalImageName, err := r.Config.UISvc.PrepareUIImage(existingTenant)
		if err != nil {
			return err
		}

		externalPort, err := helpers.ExtractPortFromURL(ap.PlatformCompleteURL)
		if err != nil {
			return err
		}

		err = r.Config.UISvc.DeployContainer(existingTenant, finalImageName, externalPort)
		if err != nil {
			return err
		}
	}

	return nil
}

func init() {
	rootCmd.AddCommand(deployUiCmd)
	deployUiCmd.PersistentFlags().BoolVarP(&ap.BuildImages, "buildImages", "b", false, "Build Docker images")
	deployUiCmd.PersistentFlags().BoolVarP(&ap.UpdateCloned, "updateCloned", "u", false, "Update Git cloned projects")
	deployUiCmd.PersistentFlags().BoolVarP(&ap.SingleTenant, "singleTenant", "T", true, "Use for Single Tenant workflow")
	deployUiCmd.PersistentFlags().BoolVarP(&ap.EnableECSRequests, "enableEcsRequests", "e", false, "Enable ECS requests")
}
