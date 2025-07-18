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

	"github.com/folio-org/eureka-cli/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const undeployModulesCommand = "Undeploy Modules"

// undeployModulesCmd represents the undeployModules command
var undeployModulesCmd = &cobra.Command{
	Use:   "undeployModules",
	Short: "Undeploy modules",
	Long:  `Undeploy multiple modules.`,
	Run: func(cmd *cobra.Command, args []string) {
		UndeployModules()
	},
}

func UndeployModules() {
	slog.Info(undeployModulesCommand, internal.GetFuncName(), "### REMOVING APPLICATIONS ###")
	applicationId := fmt.Sprintf("%s-%s", viper.GetString(internal.ApplicationNameKey), viper.GetString(internal.ApplicationVersionKey))
	internal.RemoveApplication(undeployModulesCommand, withEnableDebug, false, applicationId)

	slog.Info(undeployModulesCommand, internal.GetFuncName(), "### UNDEPLOYING MODULES ###")
	client := internal.CreateDockerClient(undeployModulesCommand)
	defer func() {
		_ = client.Close()
	}()

	internal.UndeployModuleByNamePattern(undeployModulesCommand, client, fmt.Sprintf(internal.ProfileContainerPattern, viper.GetString(internal.ProfileNameKey)), true)
}

func init() {
	rootCmd.AddCommand(undeployModulesCmd)
}
