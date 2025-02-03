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
	"fmt"
	"log/slog"
	"os/exec"

	"github.com/folio-org/eureka-cli/internal"
	"github.com/spf13/cobra"
)

const buildAndPushUiCmdCommand string = "Build and push UI"

// buildAndPushUiCmd represents the buildAndPushUi command
var buildAndPushUiCmd = &cobra.Command{
	Use:   "buildAndPushUi",
	Short: "Build and push UI",
	Long:  `Build and push UI image to DockerHub.`,
	Run: func(cmd *cobra.Command, args []string) {
		BuildAndPushUi()
	},
}

func BuildAndPushUi() {
	slog.Info(deployUiCommand, internal.GetFuncName(), "### CLONING & UPDATING UI ###")

	slog.Info(deployUiCommand, internal.GetFuncName(), fmt.Sprintf("Cloning %s from a %s branch", platformCompleteDir, defaultStripesBranch))
	outputDir := fmt.Sprintf("%s/%s", internal.DockerComposeWorkDir, platformCompleteDir)
	stripesBranch := internal.GetStripesBranch(deployUiCommand, defaultStripesBranch)
	internal.GitCloneRepository(deployUiCommand, enableDebug, internal.PlatformCompleteRepositoryUrl, stripesBranch, outputDir, false)

	if updateCloned {
		slog.Info(deployUiCommand, internal.GetFuncName(), fmt.Sprintf("Pulling updates for %s from origin", platformCompleteDir))
		internal.GitResetHardPullFromOriginRepository(deployUiCommand, enableDebug, internal.PlatformCompleteRepositoryUrl, defaultStripesBranch, outputDir)
	}

	slog.Info(deployUiCommand, internal.GetFuncName(), "### BUILDING AND PUSHING UI IMAGE TO DOCKER HUB ###")

	slog.Info(buildAndPushUiCmdCommand, internal.GetFuncName(), "Copying platform complete UI configs")
	configName := "stripes.config.js"
	internal.CopySingleFile(buildAndPushUiCmdCommand, fmt.Sprintf("%s/eureka-tpl/%s", outputDir, configName), fmt.Sprintf("%s/%s", outputDir, configName))

	slog.Info(buildAndPushUiCmdCommand, internal.GetFuncName(), "Preparing platform complete UI config")
	internal.PrepareStripesConfigJs(buildAndPushUiCmdCommand, outputDir, tenant, kongExternalUrl, keycloakExternalUrl, platformCompleteExternalUrl)
	internal.PreparePackageJson(buildAndPushUiCmdCommand, outputDir, tenant)

	imageName := fmt.Sprintf("platform-complete-ui-%s", tenant)

	slog.Info(buildAndPushUiCmdCommand, internal.GetFuncName(), "Building platform complete UI from a Dockerfile")
	internal.RunCommandFromDir(buildAndPushUiCmdCommand, exec.Command("docker", "build", "--tag", imageName,
		"--build-arg", fmt.Sprintf("OKAPI_URL=%s", kongExternalUrl),
		"--build-arg", fmt.Sprintf("TENANT_ID=%s", tenant),
		"--file", "./docker/Dockerfile",
		"--progress", "plain",
		"--no-cache",
		".",
	), outputDir)

	slog.Info(buildAndPushUiCmdCommand, internal.GetFuncName(), "Tagging platform complete image UI")
	internal.RunCommand(buildAndPushUiCmdCommand, exec.Command("docker", "tag", imageName, fmt.Sprintf("%s/%s", namespace, imageName)))

	slog.Info(buildAndPushUiCmdCommand, internal.GetFuncName(), "Pushing platform complete UI image to DockerHub")
	internal.RunCommand(buildAndPushUiCmdCommand, exec.Command("docker", "push", fmt.Sprintf("%s/%s:latest", namespace, imageName)))
}

func init() {
	rootCmd.AddCommand(buildAndPushUiCmd)
	buildAndPushUiCmd.PersistentFlags().StringVarP(&tenant, "tenant", "t", "", "Tenant (required)")
	buildAndPushUiCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "DockerHub namespace (required)")
	buildAndPushUiCmd.PersistentFlags().BoolVarP(&updateCloned, "updateCloned", "u", false, "Update cloned projects")
	buildAndPushUiCmd.MarkPersistentFlagRequired("tenant")
	buildAndPushUiCmd.MarkPersistentFlagRequired("namespace")
}
