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
	"path/filepath"
	"time"

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
	start := time.Now()

	slog.Info(buildAndPushUiCmdCommand, internal.GetFuncName(), "### BUILDING AND PUSHING PLATFORM COMPLETE UI IMAGE TO DOCKER HUB ###")
	outputDir := cloneUpdatePlatformCompleteUiRepository(buildAndPushUiCmdCommand, withEnableDebug, withUpdateCloned)
	imageName := buildPlatformCompleteUiImageLocally(buildAndPushUiCmdCommand, withEnableEcsRequests, outputDir, withTenant)
	pushPlatformCompleteUiImageToRegistry(buildAndPushUiCmdCommand, withNamespace, imageName)

	slog.Info(buildAndPushUiCmdCommand, "Elapsed, duration", time.Since(start))
}

func buildPlatformCompleteUiImageLocally(commandName string, enableEcsRequests bool, outputDir string, existingTenant string) (finalImageName string) {
	finalImageName = fmt.Sprintf("platform-complete-ui-%s", existingTenant)

	slog.Info(commandName, internal.GetFuncName(), "Copying platform complete UI configs")
	configName := "stripes.config.js"
	internal.CopySingleFile(commandName, filepath.Join(outputDir, "eureka-tpl", configName), filepath.Join(outputDir, configName))

	slog.Info(commandName, internal.GetFuncName(), "Preparing platform complete UI configs")
	internal.PrepareStripesConfigJs(commandName, outputDir, existingTenant, kongExternalUrl, keycloakExternalUrl, platformCompleteExternalUrl, enableEcsRequests)
	internal.PreparePackageJson(commandName, outputDir, existingTenant)

	slog.Info(commandName, internal.GetFuncName(), "Building platform complete UI from a Dockerfile")
	internal.RunCommandFromDir(commandName, exec.Command("docker", "build", "--tag", finalImageName,
		"--build-arg", fmt.Sprintf("OKAPI_URL=%s", kongExternalUrl),
		"--build-arg", fmt.Sprintf("TENANT_ID=%s", existingTenant),
		"--file", "./docker/Dockerfile",
		"--progress", "plain",
		"--no-cache",
		".",
	), outputDir)

	return finalImageName
}

func cloneUpdatePlatformCompleteUiRepository(commandName string, enableDebug bool, updateCloned bool) (outputDir string) {
	slog.Info(commandName, internal.GetFuncName(), "### CLONING & UPDATING PLATFORM COMPLETE UI REPOSITORY ###")
	branchName := internal.GetPlatformCompleteStripesBranch(commandName)

	repository := internal.NewRepository(commandName, internal.PlatformCompleteRepositoryUrl, internal.DefaultPlatformCompleteOutputDir, branchName)

	internal.GitCloneRepository(commandName, enableDebug, false, repository)
	if updateCloned {
		internal.GitResetHardPullFromOriginRepository(commandName, enableDebug, repository)
	}

	return repository.OutputDir
}

func pushPlatformCompleteUiImageToRegistry(commandName string, namespace string, imageName string) {
	slog.Info(commandName, internal.GetFuncName(), "### PUSHING PLATFORM COMPLETE UI IMAGE TO DOCKER HUB ###")

	slog.Info(commandName, internal.GetFuncName(), "Tagging platform complete UI image")
	internal.RunCommand(commandName, exec.Command("docker", "tag", imageName, fmt.Sprintf("%s/%s", namespace, imageName)))

	slog.Info(commandName, internal.GetFuncName(), "Pushing new platform complete UI image to DockerHub")
	internal.RunCommand(commandName, exec.Command("docker", "push", fmt.Sprintf("%s/%s:latest", namespace, imageName)))
}

func init() {
	rootCmd.AddCommand(buildAndPushUiCmd)
	buildAndPushUiCmd.PersistentFlags().StringVarP(&withTenant, "tenant", "t", "", "Tenant (required)")
	buildAndPushUiCmd.PersistentFlags().StringVarP(&withNamespace, "namespace", "n", "", "DockerHub namespace (required)")
	buildAndPushUiCmd.PersistentFlags().BoolVarP(&withUpdateCloned, "updateCloned", "u", false, "Update Git cloned projects")
	buildAndPushUiCmd.PersistentFlags().BoolVarP(&withEnableEcsRequests, "enableEcsRequests", "e", false, "Enable ECS requests")
	buildAndPushUiCmd.MarkPersistentFlagRequired("tenant")
	buildAndPushUiCmd.MarkPersistentFlagRequired("namespace")
}
