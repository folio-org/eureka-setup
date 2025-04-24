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

	"github.com/docker/docker/api/types/filters"
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
	internal.RemoveApplications(undeployModulesCommand, enableDebug, false)

	slog.Info(undeployModulesCommand, internal.GetFuncName(), "### UNDEPLOYING MODULES ###")
	client := internal.CreateDockerClient(undeployModulesCommand)
	defer client.Close()

	filters := filters.NewArgs(filters.KeyValuePair{Key: "name", Value: fmt.Sprintf(internal.MultipleModulesContainerPattern, viper.GetString(internal.ProfileNameKey))})
	deployedModules := internal.GetDeployedModules(undeployModulesCommand, client, filters)
	for _, deployedModule := range deployedModules {
		internal.UndeployModule(undeployModulesCommand, client, deployedModule)
	}
}

func init() {
	rootCmd.AddCommand(undeployModulesCmd)
}
