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
	deployUiCommand     string = "Deploy UI"
	platformCompleteDir string = "platform-complete"

	snapshotBranchName plumbing.ReferenceName = "snapshot"
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
	slog.Info(deployUiCommand, "### CLONING & UPDATING UI ###", "")

	slog.Info(deployUiCommand, fmt.Sprintf("Cloning %s from a %s branch", platformCompleteDir, snapshotBranchName), "")
	outputDir := fmt.Sprintf("%s/%s", internal.DockerComposeWorkDir, platformCompleteDir)
	internal.GitCloneRepository(deployUiCommand, enableDebug, internal.PlatformCompleteRepositoryUrl, snapshotBranchName, outputDir, false)

	slog.Info(deployUiCommand, fmt.Sprintf("Pulling updates for %s from origin", platformCompleteDir), "")
	internal.GitResetHardPullFromOriginRepository(deployUiCommand, enableDebug, internal.PlatformCompleteRepositoryUrl, snapshotBranchName, outputDir)

	slog.Info(deployUiCommand, "### ACQUIRING KEYCLOAK MASTER ACCESS TOKEN ###", "")
	masterAccessToken := internal.GetKeycloakMasterRealmAccessToken(createUsersCommand, enableDebug)

	for _, value := range internal.GetTenants(deployUiCommand, enableDebug, false) {
		mapEntry := value.(map[string]interface{})
		tenant := mapEntry["name"].(string)

		if !internal.HasTenant(tenant) || !internal.DeployUi(tenant) {
			continue
		}

		slog.Info(deployUiCommand, "### UPDATING KEYCLOAK PUBLIC CLIENT", "")
		internal.UpdateKeycloakPublicClientParams(deployUiCommand, enableDebug, tenant, masterAccessToken)

		slog.Info(deployUiCommand, "### COPYING PLATFORM COMPLETE UI CONFIGS ###", "")
		internal.RunCommandFromDir(deployUiCommand, exec.Command("cp", "-R", "-f", "eureka-tpl/*", "."), outputDir)

		slog.Info(deployUiCommand, "### PREPARING PLATFORM COMPLETE UI CONFIGS ###", "")
		internal.PrepareStripesConfigJs(deployUiCommand, outputDir, tenant)
		internal.PreparePackageJson(deployUiCommand, outputDir, tenant)

		slog.Info(deployUiCommand, "### BUILDING PLATFORM COMPLETE UI FROM A DOCKERFILE ###", "")
		internal.RunCommandFromDir(deployUiCommand, exec.Command("docker", "build", "--tag", "platform-complete-ui",
			"--build-arg", fmt.Sprintf("OKAPI_URL=%s", viper.GetString(internal.ResourcesKongKey)),
			"--build-arg", fmt.Sprintf("TENANT_ID=%s", tenant),
			"--file", "./docker/Dockerfile",
			"--no-cache",
			".",
		), outputDir)

		slog.Info(deployUiCommand, "### RUNNING PLATFORM COMPLETE UI CONTAINER ###", "")
		containerName := fmt.Sprintf("eureka-platform-complete-ui-%s", tenant)
		internal.RunCommand(deployUiCommand, exec.Command("docker", "run", "--name", containerName,
			"--hostname", containerName,
			"--publish", "3000:80",
			"--restart", "unless-stopped",
			"--detach",
			"platform-complete-ui:latest",
		))

		slog.Info(deployUiCommand, "### CONNECTING PLATFORM COMPLETE UI CONTAINER TO EUREKA NETWORK ###", "")
		internal.RunCommand(deployUiCommand, exec.Command("docker", "network", "connect", "eureka", containerName))
	}
}

func init() {
	rootCmd.AddCommand(deployUiCmd)
}
