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
	"slices"

	"github.com/folio-org/eureka-cli/internal"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	deployUiCommand     string = "Deploy UI"
	platformCompleteDir string = "platform-complete"

	// TODO Make configurable
	//branchName plumbing.ReferenceName = "snapshot"
	branchName plumbing.ReferenceName = "R1-2024"
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

	slog.Info(deployUiCommand, "### CLONING PLATFORM COMPLETE UI FROM A SNAPSHOT BRANCH ###", "")
	outputDir := fmt.Sprintf("%s/%s", internal.DockerComposeWorkDir, platformCompleteDir)
	internal.GitCloneRepository(deployUiCommand, enableDebug, internal.PlatformCompleteRepositoryUrl, branchName, outputDir, false)

	for _, value := range internal.GetTenants(deployUiCommand, enableDebug, false) {
		mapEntry := value.(map[string]interface{})
		tenant := mapEntry["name"].(string)

		if !slices.Contains(viper.GetStringSlice(internal.TenantsKey), tenant) {
			continue
		}

		slog.Info(deployUiCommand, "### UPDATING KEYCLOAK PUBLIC CLIENT", "")
		masterAccessToken := internal.GetKeycloakMasterRealmAccessToken(createUsersCommand, enableDebug)
		internal.UpdateKeycloakPublicClientParams(deployUiCommand, enableDebug, tenant, masterAccessToken)

		slog.Info(deployUiCommand, "### COPYING PLATFORM COMPLETE UI CONFIGS ###", "")
		files, err := filepath.Glob("eureka-tpl/*")
		if err != nil {
			// Handle error if the pattern doesn't match any files
			slog.Info("Failed to glob files: %v", err)
		}
		if len(files) == 0 {
			slog.Info("No files matched in eureka-tpl/*. Not copying.")
		} else {
			// Append the destination directory to the list of files
			args := append([]string{"-R", "-f"}, files...)
			args = append(args, ".")
			internal.RunCommandFromDir(deployUiCommand, exec.Command("cp", args...), outputDir)
		}

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
