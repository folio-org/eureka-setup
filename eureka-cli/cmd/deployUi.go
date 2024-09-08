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
	"slices"

	"github.com/folio-org/eureka-cli/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	deployUiCommand     string = "Deploy UI"
	platformCompleteDir string = "platform-complete"
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
	slog.Info(deployUiCommand, "### ACQUIRING KEYCLOAK MASTER ACCESS TOKEN ###", "")
	masterAccessToken := internal.GetKeycloakMasterRealmAccessToken(createUsersCommand, enableDebug)

	for _, value := range internal.GetTenants(deployUiCommand, enableDebug, false) {
		mapEntry := value.(map[string]interface{})
		tenant := mapEntry["name"].(string)

		if !slices.Contains(viper.GetStringSlice(internal.TenantsKey), tenant) {
			continue
		}

		slog.Info(deployUiCommand, "### UPDATING KEYCLOAK PUBLIC CLIENTS", "")
		internal.UpdateKeycloakPublicClientParams(deployUiCommand, enableDebug, tenant, masterAccessToken)

		slog.Info(deployUiCommand, "### CLONING PLATFORM COMPLETE FROM A SNAPSHOT BRANCH ###", "")
		outputDir := fmt.Sprintf("%s/%s", internal.DockerComposeWorkDir, platformCompleteDir)
		internal.GitCloneRepository(deployUiCommand, enableDebug, internal.PlatformCompleteRepositoryUrl, outputDir, false)

		slog.Info(deployUiCommand, "### BUILDING PLATFORM COMPLETE FROM A DOCKERFILE ###", "")
		internal.RunCommandFromDir(deployUiCommand, exec.Command("cp", "-R", "-f", "eureka-tpl/*", "."), outputDir)

		slog.Info(deployUiCommand, "### PREPARING PLATFORM COMPLETE CONFIGS ###", "")
		internal.PrepareStripesConfigJson(deployUiCommand, outputDir, tenant)

		slog.Info(deployUiCommand, "### BUILDING PLATFORM COMPLETE FROM A DOCKERFILE ###", "")
		internal.RunCommandFromDir(deployUiCommand, exec.Command("docker", "build", "--tag", "platform-complete-ui",
			"--build-arg", fmt.Sprintf("OKAPI_URL=%s", internal.PlatformCompleteUrl),
			"--build-arg", fmt.Sprintf("TENANT_ID=%s", tenant),
			"--file", "./docker/Dockerfile",
			"--no-cache",
			".",
		), outputDir)

		slog.Info(deployUiCommand, "### RUNNING PLATFORM COMPLETE CONTAINER ###", "")
		containerName := fmt.Sprintf("eureka-platform-complete-ui-%s", tenant)
		internal.RunCommand(deployUiCommand, exec.Command("docker", "run", "--name", containerName,
			"--hostname", containerName,
			"--publish", "80:80",
			"--restart", "unless-stopped",
			"--detach",
			"platform-complete-ui:latest",
		))

		slog.Info(deployUiCommand, "### CONNECTING PLATFORM COMPLETE CONTAINER TO EUREKA NETWORK ###", "")
		internal.RunCommand(deployUiCommand, exec.Command("docker", "network", "connect", "eureka", containerName))
	}
}

func init() {
	rootCmd.AddCommand(deployUiCmd)
}
