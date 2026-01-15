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

	"github.com/folio-org/eureka-setup/eureka-cli/action"
	"github.com/folio-org/eureka-setup/eureka-cli/constant"
	"github.com/folio-org/eureka-setup/eureka-cli/helpers"
	"github.com/spf13/cobra"
)

// deployUiCmd represents the deployUi command
var deployUiCmd = &cobra.Command{
	Use:   "deployUi",
	Short: "Deploy UI",
	Long:  `Deploy the UI container.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		run, err := New(action.DeployUi)
		if err != nil {
			return err
		}

		return run.ConsortiumPartition(func(consortiumName string, tenantType constant.TenantType) error {
			return run.DeployUi(consortiumName, tenantType)
		})
	},
}

func (run *Run) DeployUi(consortiumName string, tenantType constant.TenantType) error {
	return run.TenantPartition(consortiumName, tenantType, func(configTenant, tenantType string) error {
		slog.Info(run.Config.Action.Name, "text", "DEPLOYING UI")
		if helpers.IsUIEnabled(configTenant, run.Config.Action.ConfigTenants) {
			if err := run.Config.TenantSvc.SetConfigTenantParams(configTenant); err != nil {
				return err
			}

			finalImageName, err := run.Config.UISvc.PrepareImage(configTenant)
			if err != nil {
				return err
			}

			externalPort, err := helpers.GetPortFromURL(params.PlatformCompleteURL)
			if err != nil {
				return err
			}
			if err := run.Config.UISvc.DeployContainer(configTenant, finalImageName, externalPort); err != nil {
				return err
			}
		}

		return nil
	})
}

func init() {
	rootCmd.AddCommand(deployUiCmd)
	deployUiCmd.PersistentFlags().BoolVarP(&params.BuildImages, action.BuildImages.Long, action.BuildImages.Short, false, action.BuildImages.Description)
	deployUiCmd.PersistentFlags().BoolVarP(&params.UpdateCloned, action.UpdateCloned.Long, action.UpdateCloned.Short, false, action.UpdateCloned.Description)
	deployUiCmd.PersistentFlags().BoolVarP(&params.SingleTenant, action.SingleTenant.Long, action.SingleTenant.Short, true, action.SingleTenant.Description)
	deployUiCmd.PersistentFlags().BoolVarP(&params.EnableECSRequests, action.EnableECSRequests.Long, action.EnableECSRequests.Short, false, action.EnableECSRequests.Description)
}
