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
	slog.Info(deployUiCommand, internal.GetFuncName(), "### CLONING & UPDATING UI ###")

	slog.Info(deployUiCommand, internal.GetFuncName(), fmt.Sprintf("Cloning %s from a %s branch", platformCompleteDir, defaultStripesBranch))
	outputDir := fmt.Sprintf("%s/%s", internal.DockerComposeWorkDir, platformCompleteDir)
	stripesBranch := internal.GetStripesBranch(deployUiCommand, defaultStripesBranch)
	internal.GitCloneRepository(deployUiCommand, enableDebug, internal.PlatformCompleteRepositoryUrl, stripesBranch, outputDir, false)

	if updateCloned {
		slog.Info(deployUiCommand, internal.GetFuncName(), fmt.Sprintf("Pulling updates for %s from origin", platformCompleteDir))
		internal.GitResetHardPullFromOriginRepository(deployUiCommand, enableDebug, internal.PlatformCompleteRepositoryUrl, defaultStripesBranch, outputDir)
	}

	slog.Info(deployUiCommand, internal.GetFuncName(), "### DEPLOYING UI ###")

	slog.Info(deployUiCommand, internal.GetFuncName(), "Acquiring keycloak master access token")
	masterAccessToken := internal.GetKeycloakMasterRealmAccessToken(createUsersCommand, enableDebug)

	for _, value := range internal.GetTenants(deployUiCommand, enableDebug, false) {
		mapEntry := value.(map[string]interface{})
		tenant := mapEntry["name"].(string)

		if !internal.HasTenant(tenant) || !internal.DeployUi(tenant) {
			continue
		}

		slog.Info(deployUiCommand, internal.GetFuncName(), "Updating keycloak public client")
		internal.UpdateKeycloakPublicClientParams(deployUiCommand, enableDebug, tenant, masterAccessToken, platformCompleteExternalUrl)

		slog.Info(deployUiCommand, internal.GetFuncName(), "Copying platform complete UI configs")
		configName := "stripes.config.js"
		internal.CopySingleFile(deployUiCommand, fmt.Sprintf("%s/eureka-tpl/%s", outputDir, configName), fmt.Sprintf("%s/%s", outputDir, configName))

		slog.Info(deployUiCommand, internal.GetFuncName(), "Preparing platform complete UI config")
		internal.PrepareStripesConfigJs(deployUiCommand, outputDir, tenant, kongExternalUrl, keycloakExternalUrl, platformCompleteExternalUrl)
		internal.PreparePackageJson(deployUiCommand, outputDir, tenant)

		slog.Info(deployUiCommand, internal.GetFuncName(), "Building platform complete UI from a Dockerfile")
		internal.RunCommandFromDir(deployUiCommand, exec.Command("docker", "build", "--tag", "platform-complete-ui",
			"--build-arg", fmt.Sprintf("OKAPI_URL=%s", kongExternalUrl),
			"--build-arg", fmt.Sprintf("TENANT_ID=%s", tenant),
			"--file", "./docker/Dockerfile",
			"--progress", "plain",
			"--no-cache",
			".",
		), outputDir)

		slog.Info(deployUiCommand, internal.GetFuncName(), "Running platform complete UI container")
		containerName := fmt.Sprintf("eureka-platform-complete-ui-%s", tenant)
		internal.RunCommand(deployUiCommand, exec.Command("docker", "run", "--name", containerName,
			"--hostname", containerName,
			"--publish", "3000:80",
			"--restart", "unless-stopped",
			"--detach",
			"platform-complete-ui:latest",
		))

		slog.Info(deployUiCommand, internal.GetFuncName(), "Connecting platform complete UI container to Eureka network")
		internal.RunCommand(deployUiCommand, exec.Command("docker", "network", "connect", "eureka", containerName))
	}
}

func init() {
	rootCmd.AddCommand(deployUiCmd)
	deployUiCmd.PersistentFlags().BoolVarP(&updateCloned, "updateCloned", "u", false, "Update cloned projects")
}
