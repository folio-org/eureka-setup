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
	"os"
	"time"

	"github.com/folio-org/eureka-cli/action"
	"github.com/spf13/cobra"
)

// buildAndPushUiCmd represents the buildAndPushUi command
var buildAndPushUiCmd = &cobra.Command{
	Use:   "buildAndPushUi",
	Short: "Build and push UI",
	Long:  `Build and push UI image to DockerHub.`,
	Run: func(cmd *cobra.Command, args []string) {
		NewRun(action.BuildAndPushUi).BuildAndPushUi()
	},
}

func (r *Run) BuildAndPushUi() {
	start := time.Now()

	r.Config.TenantStep.SetDefaultConfigTenantParams(&rp, rp.Tenant)

	slog.Info(r.Config.Action.Name, "text", "BUILDING AND PUSHING PLATFORM COMPLETE UI IMAGE TO DOCKER HUB")
	outputDir := r.Config.UIStep.CloneAndUpdateUIRepository(rp.UpdateCloned)
	imageName := r.Config.UIStep.BuildImage(&rp, outputDir, rp.Tenant)
	r.Config.DockerClient.PushImage(rp.Namespace, imageName)

	slog.Info(r.Config.Action.Name, "text", fmt.Sprintf("Elapsed, duration %.1f", time.Since(start).Minutes()))
}

func init() {
	rootCmd.AddCommand(buildAndPushUiCmd)
	buildAndPushUiCmd.PersistentFlags().StringVarP(&rp.Namespace, "namespace", "n", "", "DockerHub namespace (required)")
	buildAndPushUiCmd.PersistentFlags().StringVarP(&rp.Tenant, "tenant", "t", "", "Tenant (required)")
	buildAndPushUiCmd.PersistentFlags().StringVarP(&rp.PlatformCompleteURL, "PlatformCompleteURL", "P", "http://localhost:3000", "Platform Complete UI url")
	buildAndPushUiCmd.PersistentFlags().BoolVarP(&rp.SingleTenant, "singleTenant", "T", true, "Use for Single Tenant workflow")
	buildAndPushUiCmd.PersistentFlags().BoolVarP(&rp.EnableECSRequests, "enableEcsRequests", "e", false, "Enable ECS requests")
	buildAndPushUiCmd.PersistentFlags().BoolVarP(&rp.UpdateCloned, "updateCloned", "u", false, "Update Git cloned projects")
	if err := buildAndPushUiCmd.MarkPersistentFlagRequired("namespace"); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	if err := buildAndPushUiCmd.MarkPersistentFlagRequired("tenant"); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
