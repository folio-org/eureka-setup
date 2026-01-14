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
	"os"
	"time"

	"github.com/j011195/eureka-setup/eureka-cli/action"
	"github.com/j011195/eureka-setup/eureka-cli/errors"
	"github.com/spf13/cobra"
)

// buildAndPushUiCmd represents the buildAndPushUi command
var buildAndPushUiCmd = &cobra.Command{
	Use:   "buildAndPushUi",
	Short: "Build and push UI",
	Long:  `Build and push UI image to DockerHub.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		run, err := New(action.BuildAndPushUi)
		if err != nil {
			return err
		}

		return run.BuildAndPushUi()
	},
}

func (run *Run) BuildAndPushUi() error {
	start := time.Now()
	if err := run.Config.TenantSvc.SetConfigTenantParams(params.Tenant); err != nil {
		return err
	}

	slog.Info(run.Config.Action.Name, "text", "BUILDING AND PUSHING PLATFORM COMPLETE UI IMAGE TO DOCKER HUB")
	outputDir, err := run.Config.UISvc.CloneAndUpdateRepository(params.UpdateCloned)
	if err != nil {
		return err
	}

	imageName, err := run.Config.UISvc.BuildImage(params.Tenant, outputDir)
	if err != nil {
		return err
	}
	if err := run.Config.DockerClient.PushImage(params.Namespace, imageName); err != nil {
		return err
	}
	slog.Info(run.Config.Action.Name, "text", "Command completed", "duration", time.Since(start))

	return nil
}

func init() {
	rootCmd.AddCommand(buildAndPushUiCmd)
	buildAndPushUiCmd.PersistentFlags().StringVarP(&params.Namespace, action.Namespace.Long, action.Namespace.Short, "", action.Namespace.Description)
	buildAndPushUiCmd.PersistentFlags().StringVarP(&params.Tenant, action.Tenant.Long, action.Tenant.Short, "", action.Tenant.Description)
	buildAndPushUiCmd.PersistentFlags().StringVarP(&params.PlatformCompleteURL, action.PlatformCompleteURL.Long, action.PlatformCompleteURL.Short, "http://localhost:3000", action.PlatformCompleteURL.Description)
	buildAndPushUiCmd.PersistentFlags().BoolVarP(&params.SingleTenant, action.SingleTenant.Long, action.SingleTenant.Short, true, action.SingleTenant.Description)
	buildAndPushUiCmd.PersistentFlags().BoolVarP(&params.EnableECSRequests, action.EnableECSRequests.Long, action.EnableECSRequests.Short, false, action.EnableECSRequests.Description)
	buildAndPushUiCmd.PersistentFlags().BoolVarP(&params.UpdateCloned, action.UpdateCloned.Long, action.UpdateCloned.Short, false, action.UpdateCloned.Description)

	if err := buildAndPushUiCmd.MarkPersistentFlagRequired(action.Namespace.Long); err != nil {
		slog.Error(errors.MarkFlagRequiredFailed(action.Namespace, err).Error())
		os.Exit(1)
	}
	if err := buildAndPushUiCmd.MarkPersistentFlagRequired(action.Tenant.Long); err != nil {
		slog.Error(errors.MarkFlagRequiredFailed(action.Tenant, err).Error())
		os.Exit(1)
	}
}
