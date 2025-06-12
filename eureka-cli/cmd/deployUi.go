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
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	deployUiCommand             string = "Deploy UI"
	platformCompleteDir         string = "platform-complete"
	kongExternalUrl             string = "http://localhost:8000"
	keycloakExternalUrl         string = "http://keycloak.eureka:8080"
	platformCompleteExternalUrl string = "http://localhost:3000"

	defaultStripesBranch plumbing.ReferenceName = "snapshot"
)

// deployUiCmd represents the deployUi command
var deployUiCmd = &cobra.Command{
	Use:   "deployUi",
	Short: "Deploy UI",
	Long:  `Deploy the UI container.`,
	Run: func(cmd *cobra.Command, args []string) {
		DeployUi()
	},
}

func DeployUi() {
	var outputDir string
	if withBuildImages {
		outputDir = cloneUpdateUi()
	}

	slog.Info(deployUiCommand, internal.GetFuncName(), "### DEPLOYING UI ###")
	keycloakMasterAccessToken := internal.GetKeycloakMasterAccessToken(createUsersCommand, withEnableDebug)

	for _, value := range internal.GetTenants(deployUiCommand, withEnableDebug, false) {
		mapEntry := value.(map[string]any)

		existingTenant := mapEntry["name"].(string)
		if !internal.HasTenant(existingTenant) || !internal.CanDeployUi(existingTenant) {
			continue
		}

		slog.Info(deployUiCommand, internal.GetFuncName(), "Updating keycloak public client")
		internal.UpdateKeycloakPublicClientParams(deployUiCommand, withEnableDebug, existingTenant, keycloakMasterAccessToken, platformCompleteExternalUrl)

		imageName := fmt.Sprintf("platform-complete-ui-%s", existingTenant)
		finalImageName, shouldReturn := prepareImage(outputDir, existingTenant, imageName)
		if shouldReturn {
			return
		}

		slog.Info(deployUiCommand, internal.GetFuncName(), "Running platform complete UI container")
		containerName := fmt.Sprintf("eureka-platform-complete-ui-%s", existingTenant)
		internal.RunCommand(deployUiCommand, exec.Command("docker", "run", "--name", containerName,
			"--hostname", containerName,
			"--publish", "3000:80",
			"--restart", "unless-stopped",
			"--cpus", "1",
			"--memory", "35m",
			"--memory-swap", "-1",
			"--detach",
			finalImageName,
		))

		slog.Info(deployUiCommand, internal.GetFuncName(), "Connecting platform complete UI container to Eureka network")
		internal.RunCommand(deployUiCommand, exec.Command("docker", "network", "connect", "eureka", containerName))
	}
}

func prepareImage(outputDir string, existingTenant string, imageName string) (string, bool) {
	if withBuildImages {
		buildImageLocally(outputDir, existingTenant, imageName)

		return imageName, false
	}

	if !viper.IsSet(internal.NamespacesPlatformCompleteUiKey) {
		errorMessage := fmt.Sprintf("cmd.deployUi - Cannot run %s image, key %s is not set in current config file", imageName, internal.NamespacesPlatformCompleteUiKey)
		internal.LogErrorPanic(deployUiCommand, errorMessage)

		return "", true
	}

	return pullImageFromRegistry(imageName), false
}

func buildImageLocally(outputDir string, existingTenant string, imageName string) {
	slog.Info(deployUiCommand, internal.GetFuncName(), "Copying platform complete UI configs")
	configName := "stripes.config.js"
	internal.CopySingleFile(deployUiCommand, fmt.Sprintf("%s/eureka-tpl/%s", outputDir, configName), fmt.Sprintf("%s/%s", outputDir, configName))

	slog.Info(deployUiCommand, internal.GetFuncName(), "Preparing platform complete UI config")
	internal.PrepareStripesConfigJs(deployUiCommand, outputDir, existingTenant, kongExternalUrl, keycloakExternalUrl, platformCompleteExternalUrl, withEnableEcsRequests)
	internal.PreparePackageJson(deployUiCommand, outputDir, existingTenant)

	slog.Info(deployUiCommand, internal.GetFuncName(), "Building platform complete UI from a Dockerfile")
	internal.RunCommandFromDir(deployUiCommand, exec.Command("docker", "build", "--tag", imageName,
		"--build-arg", fmt.Sprintf("OKAPI_URL=%s", kongExternalUrl),
		"--build-arg", fmt.Sprintf("TENANT_ID=%s", existingTenant),
		"--file", "./docker/Dockerfile",
		"--progress", "plain",
		"--no-cache",
		".",
	), outputDir)
}

func pullImageFromRegistry(imageName string) (finalImageName string) {
	finalImageName = fmt.Sprintf("%s/%s", viper.GetString(internal.NamespacesPlatformCompleteUiKey), imageName)

	slog.Info(deployUiCommand, internal.GetFuncName(), "Removing old platform complete UI image")
	internal.RunCommand(deployUiCommand, exec.Command("docker", "image", "rm", "--force", finalImageName))

	slog.Info(deployUiCommand, internal.GetFuncName(), "Pulling new platform complete UI image from DockerHub")
	internal.RunCommand(deployUiCommand, exec.Command("docker", "image", "pull", finalImageName))

	return finalImageName
}

func cloneUpdateUi() (outputDir string) {
	slog.Info(deployUiCommand, internal.GetFuncName(), "### CLONING & UPDATING UI ###")

	slog.Info(deployUiCommand, internal.GetFuncName(), fmt.Sprintf("Cloning %s from a %s branch", platformCompleteDir, defaultStripesBranch))
	outputDir = fmt.Sprintf("%s/%s", internal.GetHomeMiscDir(deployUiCommand), platformCompleteDir)
	stripesBranch := internal.GetStripesBranch(deployUiCommand, defaultStripesBranch)
	internal.GitCloneRepository(deployUiCommand, withEnableDebug, internal.PlatformCompleteRepositoryUrl, stripesBranch, outputDir, false)

	if withUpdateCloned {
		slog.Info(deployUiCommand, internal.GetFuncName(), fmt.Sprintf("Pulling updates for %s from origin", platformCompleteDir))
		internal.GitResetHardPullFromOriginRepository(deployUiCommand, withEnableDebug, internal.PlatformCompleteRepositoryUrl, defaultStripesBranch, outputDir)
	}

	return outputDir
}

func init() {
	rootCmd.AddCommand(deployUiCmd)
	deployUiCmd.PersistentFlags().BoolVarP(&withBuildImages, "buildImages", "b", false, "Build Docker images")
	deployUiCmd.PersistentFlags().BoolVarP(&withUpdateCloned, "updateCloned", "u", false, "Update Git cloned projects")
	deployUiCmd.PersistentFlags().BoolVarP(&withEnableEcsRequests, "enableEcsRequests", "e", false, "Enable ECS requests")
}
