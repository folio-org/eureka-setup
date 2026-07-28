/*
Copyright © 2026 Open Library Foundation

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

// buildUiCmd represents the buildUi command
var buildUiCmd = &cobra.Command{
	Use:   "buildUi",
	Short: "Build UI",
	Long:  `Build the UI image without deploying a container.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		run, err := New(action.BuildUi)
		if err != nil {
			return err
		}

		return run.BuildUi()
	},
}

// BuildUi builds the UI image for every config tenant with deploy-ui enabled.
// It only needs Docker and the config, so it can run before the platform is deployed.
func (run *Run) BuildUi() error {
	slog.Info(run.Config.Action.Name, "text", "BUILDING UI")
	for configTenant := range run.Config.Action.ConfigTenants {
		if !helpers.IsUIEnabled(configTenant, run.Config.Action.ConfigTenants) {
			slog.Info(run.Config.Action.Name, "text", "UI is not required for tenant", "tenant", configTenant)
			continue
		}
		if err := run.Config.TenantSvc.SetConfigTenantParams(configTenant); err != nil {
			return err
		}

		outputDir, err := run.Config.UISvc.CloneAndUpdateRepository(params.UpdateCloned)
		if err != nil {
			return err
		}
		if _, err := run.Config.UISvc.BuildImage(configTenant, outputDir); err != nil {
			return err
		}
	}

	return nil
}

func init() {
	rootCmd.AddCommand(buildUiCmd)
	buildUiCmd.PersistentFlags().BoolVarP(&params.UpdateCloned, action.UpdateCloned.Long, action.UpdateCloned.Short, false, action.UpdateCloned.Description)
	buildUiCmd.PersistentFlags().StringVarP(&params.PlatformLspURL, action.PlatformLspURL.Long, action.PlatformLspURL.Short, constant.DefaultPlatformLspURL, action.PlatformLspURL.Description)
	buildUiCmd.PersistentFlags().BoolVarP(&params.SingleTenant, action.SingleTenant.Long, action.SingleTenant.Short, true, action.SingleTenant.Description)
	buildUiCmd.PersistentFlags().BoolVarP(&params.LinkedData, action.LinkedData.Long, action.LinkedData.Short, false, action.LinkedData.Description)
	buildUiCmd.PersistentFlags().BoolVarP(&params.EnableECSRequests, action.EnableECSRequests.Long, action.EnableECSRequests.Short, false, action.EnableECSRequests.Description)
}
