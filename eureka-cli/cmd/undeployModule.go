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
	"os"

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/field"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// undeployModuleCmd represents the undeployModule command
var undeployModuleCmd = &cobra.Command{
	Use:   "undeployModule",
	Short: "Undeploy module",
	Long:  `Undeploy a single module.`,
	Run: func(cmd *cobra.Command, args []string) {
		NewRun(action.UndeployModule).UndeployModule()
	},
}

func (r *Run) UndeployModule() {
	slog.Info(r.Config.Action.Name, "text", "UNDEPLOYING MODULE")

	client := r.Config.DockerClient.Create()
	defer func() {
		_ = client.Close()
	}()

	r.Config.ModuleStep.UndeployModuleByNamePattern(client, fmt.Sprintf(constant.SingleModuleOrSidecarContainerPattern, viper.GetString(field.ProfileName), rp.ModuleName), true)
}

func init() {
	rootCmd.AddCommand(undeployModuleCmd)
	undeployModuleCmd.PersistentFlags().StringVarP(&rp.ModuleName, "moduleName", "m", "", "Module name, e.g. mod-orders (required)")
	if err := undeployModuleCmd.MarkPersistentFlagRequired("moduleName"); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
