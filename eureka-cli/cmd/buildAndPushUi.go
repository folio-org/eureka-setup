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
	"os/exec"
	"path/filepath"
	"time"

	"github.com/folio-org/eureka-cli/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	buildAndPushUiCmdCommand string = "Build and push UI"

	kongExternalUrl     string = "http://localhost:8000"
	keycloakExternalUrl string = "http://keycloak.eureka:8080"
)

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

	setCommandFlagsFromConfigFile(buildAndPushUiCmdCommand, withTenant)

	slog.Info(buildAndPushUiCmdCommand, internal.GetFuncName(), "### BUILDING AND PUSHING PLATFORM COMPLETE UI IMAGE TO DOCKER HUB ###")
	outputDir := cloneUpdatePlatformCompleteUiRepository(buildAndPushUiCmdCommand, withEnableDebug, withUpdateCloned)
	imageName := buildPlatformCompleteUiImageLocally(buildAndPushUiCmdCommand, withSingleTenant, withEnableEcsRequests, outputDir, withTenant, withPlatformCompleteUrl)
	pushPlatformCompleteUiImageToRegistry(buildAndPushUiCmdCommand, withNamespace, imageName)

	slog.Info(buildAndPushUiCmdCommand, "Elapsed, duration", time.Since(start))
}

func setCommandFlagsFromConfigFile(commandName string, existingTenant string) {
	tenants := viper.GetStringMap(internal.TenantsKey)
	if tenants == nil {
		return
	}

	var tenant = tenants[existingTenant].(map[string]any)
	if tenant[internal.TenantsSingleTenantEntryKey] != nil {
		withSingleTenant = tenant[internal.TenantsSingleTenantEntryKey].(bool)
	}
	if tenant[internal.TenantsEnableEcsRequestEntryKey] != nil {
		withEnableEcsRequests = tenant[internal.TenantsEnableEcsRequestEntryKey].(bool)
	}
	if tenant[internal.TenantsPlatformCompleteUrlEntryKey] != nil {
		withPlatformCompleteUrl = tenant[internal.TenantsPlatformCompleteUrlEntryKey].(string)
	}

	vars := []any{existingTenant, withSingleTenant, withEnableEcsRequests, withPlatformCompleteUrl}
	slog.Info(commandName, internal.GetFuncName(), fmt.Sprintf("Setting command flags from a config file, tenant: %s, singleTenant: %t, enableEcsRequests: %t, platformCompleteUrl: %s", vars...))
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

func buildPlatformCompleteUiImageLocally(commandName string, singleTenant bool, enableEcsRequests bool, outputDir string, existingTenant string, platformCompleteUrl string) (finalImageName string) {
	finalImageName = fmt.Sprintf("platform-complete-ui-%s", existingTenant)

	slog.Info(commandName, internal.GetFuncName(), "Copying platform complete UI configs")
	configName := "stripes.config.js"
	internal.CopySingleFile(commandName, filepath.Join(outputDir, "eureka-tpl", configName), filepath.Join(outputDir, configName))

	slog.Info(commandName, internal.GetFuncName(), "Preparing platform complete UI configs")
	internal.PrepareStripesConfigJs(commandName, outputDir, existingTenant, kongExternalUrl, keycloakExternalUrl, platformCompleteUrl, singleTenant, enableEcsRequests)
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

func pushPlatformCompleteUiImageToRegistry(commandName string, namespace string, imageName string) {
	slog.Info(commandName, internal.GetFuncName(), "### PUSHING PLATFORM COMPLETE UI IMAGE TO DOCKER HUB ###")

	finalImageName := fmt.Sprintf("%s/%s", namespace, imageName)

	slog.Info(commandName, internal.GetFuncName(), "Tagging platform complete UI image")
	internal.RunCommand(commandName, exec.Command("docker", "tag", imageName, finalImageName))

	slog.Info(commandName, internal.GetFuncName(), "Pushing new platform complete UI image to DockerHub")
	internal.RunCommand(commandName, exec.Command("docker", "push", finalImageName))
}

func init() {
	rootCmd.AddCommand(buildAndPushUiCmd)
	buildAndPushUiCmd.PersistentFlags().StringVarP(&withTenant, "tenant", "t", "", "Tenant (required)")
	buildAndPushUiCmd.PersistentFlags().StringVarP(&withNamespace, "namespace", "n", "", "DockerHub namespace (required)")
	buildAndPushUiCmd.PersistentFlags().StringVarP(&withPlatformCompleteUrl, "platformCompleteUrl", "P", "http://localhost:3000", "Platform Complete UI url")
	buildAndPushUiCmd.PersistentFlags().BoolVarP(&withUpdateCloned, "updateCloned", "u", false, "Update Git cloned projects")
	buildAndPushUiCmd.PersistentFlags().BoolVarP(&withSingleTenant, "singleTenant", "T", true, "Use for Single Tenant workflow")
	buildAndPushUiCmd.PersistentFlags().BoolVarP(&withEnableEcsRequests, "enableEcsRequests", "e", false, "Enable ECS requests")
	buildAndPushUiCmd.MarkPersistentFlagRequired("tenant")
	buildAndPushUiCmd.MarkPersistentFlagRequired("namespace")
}
