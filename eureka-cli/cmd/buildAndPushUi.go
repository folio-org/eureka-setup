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

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/spf13/cobra"
)

// buildAndPushUiCmd represents the buildAndPushUi command
var buildAndPushUiCmd = &cobra.Command{
	Use:   "buildAndPushUi",
	Short: "Build and push UI",
	Long:  `Build and push UI image to DockerHub.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := New(action.BuildAndPushUi)
		if err != nil {
			return err
		}

		return r.BuildAndPushUi()
	},
}

func (r *Run) BuildAndPushUi() error {
	start := time.Now()

	err := r.Config.TenantSvc.SetDefaultConfigTenantParams(ap.Tenant)
	if err != nil {
		return err
	}

	slog.Info(r.Config.Action.Name, "text", "BUILDING AND PUSHING PLATFORM COMPLETE UI IMAGE TO DOCKER HUB")
	outputDir, err := r.Config.UISvc.CloneAndUpdateUIRepository(ap.UpdateCloned)
	if err != nil {
		return err
	}

	imageName, err := r.Config.UISvc.BuildImage(outputDir, ap.Tenant)
	if err != nil {
		return err
	}

	err = r.Config.DockerClient.PushImage(ap.Namespace, imageName)
	if err != nil {
		return err
	}
	helpers.LogCompletion(r.Config.Action.Name, start)

	return nil
}

func init() {
	rootCmd.AddCommand(buildAndPushUiCmd)
	buildAndPushUiCmd.PersistentFlags().StringVarP(&ap.Namespace, "namespace", "n", "", "DockerHub namespace (required)")
	buildAndPushUiCmd.PersistentFlags().StringVarP(&ap.Tenant, "tenant", "t", "", "Tenant (required)")
	buildAndPushUiCmd.PersistentFlags().StringVarP(&ap.PlatformCompleteURL, "PlatformCompleteURL", "P", "http://localhost:3000", "Platform Complete UI url")
	buildAndPushUiCmd.PersistentFlags().BoolVarP(&ap.SingleTenant, "singleTenant", "T", true, "Use for Single Tenant workflow")
	buildAndPushUiCmd.PersistentFlags().BoolVarP(&ap.EnableECSRequests, "enableEcsRequests", "e", false, "Enable ECS requests")
	buildAndPushUiCmd.PersistentFlags().BoolVarP(&ap.UpdateCloned, "updateCloned", "u", false, "Update Git cloned projects")
	if err := buildAndPushUiCmd.MarkPersistentFlagRequired("namespace"); err != nil {
		slog.Error("failed to mark namespace flag as required", "error", err)
		os.Exit(1)
	}
	if err := buildAndPushUiCmd.MarkPersistentFlagRequired("tenant"); err != nil {
		slog.Error("failed to mark tenant flag as required", "error", err)
		os.Exit(1)
	}
}
