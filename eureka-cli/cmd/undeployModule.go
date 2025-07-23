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

const undeployModuleCommand = "Undeploy Module"

// undeployModuleCmd represents the undeployModule command
var undeployModuleCmd = &cobra.Command{
	Use:   "undeployModule",
	Short: "Undeploy module",
	Long:  `Undeploy a single module.`,
	Run: func(cmd *cobra.Command, args []string) {
		UndeployModule()
	},
}

func UndeployModule() {
	slog.Info(undeployModuleCommand, internal.GetFuncName(), "### UNDEPLOYING MODULE ###")
	client := internal.CreateDockerClient(undeployModuleCommand)
	defer func() {
		_ = client.Close()
	}()

	internal.UndeployModuleByNamePattern(undeployModuleCommand, client, fmt.Sprintf(internal.SingleModuleOrSidecarContainerPattern, viper.GetString(internal.ProfileNameKey), withModuleName), true)
}

func init() {
	rootCmd.AddCommand(undeployModuleCmd)
	undeployModuleCmd.PersistentFlags().StringVarP(&withModuleName, "moduleName", "m", "", "Module name, e.g. mod-orders (required)")
	if err := undeployModuleCmd.MarkPersistentFlagRequired("moduleName"); err != nil {
		slog.Error(undeployModuleCommand, internal.GetFuncName(), "undeployModuleCmd.MarkPersistentFlagRequired error")
		panic(err)
	}
}
