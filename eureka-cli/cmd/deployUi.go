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

	"github.com/folio-org/eureka-cli/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	deployUiCommand string = "Deploy UI"

	kongExternalUrl             string = "http://localhost:8000"
	keycloakExternalUrl         string = "http://keycloak.eureka:8080"
	platformCompleteExternalUrl string = "http://localhost:3000"
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
	slog.Info(deployUiCommand, internal.GetFuncName(), "### DEPLOYING UI ###")

	for _, value := range internal.GetTenants(deployUiCommand, withEnableDebug, false) {
		existingTenant := value.(map[string]any)["name"].(string)
		if !internal.HasTenant(existingTenant) || !internal.CanDeployUi(existingTenant) {
			continue
		}

		setCommandFlagsFromConfigFile(deployUiCommand, existingTenant)

		finalImageName := preparePlatformCompleteUiImage(withEnableDebug, withBuildImages, withUpdateCloned, withSingleTenant, withEnableEcsRequests, existingTenant)
		deployPlatformCompleteUiContainer(deployUiCommand, existingTenant, finalImageName)
	}
}

func preparePlatformCompleteUiImage(enabledDebug, buildImages, updateCloned, singleTenant, enableEcsRequests bool, existingTenant string) (finalImageName string) {
	imageName := fmt.Sprintf("platform-complete-ui-%s", existingTenant)
	if buildImages {
		outputDir := cloneUpdatePlatformCompleteUiRepository(deployUiCommand, enabledDebug, updateCloned)
		return buildPlatformCompleteUiImageLocally(deployUiCommand, singleTenant, enableEcsRequests, outputDir, existingTenant)
	}

	return forcePullPlatformCompleteUiImageFromRegistry(deployUiCommand, imageName)
}

func forcePullPlatformCompleteUiImageFromRegistry(commandName string, imageName string) (finalImageName string) {
	slog.Info(commandName, internal.GetFuncName(), "### PULLING PLATFORM COMPLETE UI IMAGE FROM DOCKER HUB ###")
	if !viper.IsSet(internal.NamespacesPlatformCompleteUiKey) {
		internal.LogErrorPanic(commandName, fmt.Sprintf("cmd.deployUi - Cannot run %s image, key %s is not set in current config file", imageName, internal.NamespacesPlatformCompleteUiKey))
		return ""
	}

	finalImageName = fmt.Sprintf("%s/%s", viper.GetString(internal.NamespacesPlatformCompleteUiKey), imageName)

	slog.Info(commandName, internal.GetFuncName(), "Removing old platform complete UI image")
	internal.RunCommand(commandName, exec.Command("docker", "image", "rm", "--force", finalImageName))

	slog.Info(commandName, internal.GetFuncName(), "Pulling new platform complete UI image from DockerHub")
	internal.RunCommand(commandName, exec.Command("docker", "image", "pull", finalImageName))

	return finalImageName
}

func deployPlatformCompleteUiContainer(commandName string, existingTenant string, finalImageName string) {
	slog.Info(commandName, internal.GetFuncName(), fmt.Sprintf("Deploying platform complete UI container for %s tenant", existingTenant))
	containerName := fmt.Sprintf("eureka-platform-complete-ui-%s", existingTenant)

	internal.RunCommand(commandName, exec.Command("docker", "run", "--name", containerName,
		"--hostname", containerName,
		"--publish", "3000:80",
		"--restart", "unless-stopped",
		"--cpus", "1",
		"--memory", "35m",
		"--memory-swap", "-1",
		"--detach",
		finalImageName,
	))

	slog.Info(commandName, internal.GetFuncName(), fmt.Sprintf("Connecting platform complete UI container for %s tenant to %s network", existingTenant, internal.DefaultNetworkId))
	internal.RunCommand(commandName, exec.Command("docker", "network", "connect", internal.DefaultNetworkId, containerName))
}

func init() {
	rootCmd.AddCommand(deployUiCmd)
	deployUiCmd.PersistentFlags().BoolVarP(&withBuildImages, "buildImages", "b", false, "Build Docker images")
	deployUiCmd.PersistentFlags().BoolVarP(&withUpdateCloned, "updateCloned", "u", false, "Update Git cloned projects")
	deployUiCmd.PersistentFlags().BoolVarP(&withSingleTenant, "singleTenant", "T", true, "Use for Single Tenant workflow")
	deployUiCmd.PersistentFlags().BoolVarP(&withEnableEcsRequests, "enableEcsRequests", "e", false, "Enable ECS requests")
}
