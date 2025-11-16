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
	"github.com/folio-org/eureka-cli/errors"
	"github.com/folio-org/eureka-cli/field"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// undeployModuleCmd represents the undeployModule command
var undeployModuleCmd = &cobra.Command{
	Use:   "undeployModule",
	Short: "Undeploy module",
	Long:  `Undeploy a single module.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		run, err := New(action.UndeployModule)
		if err != nil {
			return err
		}

		return run.UndeployModule()
	},
}

func (run *Run) UndeployModule() error {
	slog.Info(run.Config.Action.Name, "text", "UNDEPLOYING MODULE")
	client, err := run.Config.DockerClient.Create()
	if err != nil {
		return err
	}
	defer run.Config.DockerClient.Close(client)

	pattern := fmt.Sprintf(constant.SingleModuleOrSidecarContainerPattern, run.Config.Action.ConfigProfile, params.ModuleName)
	return run.Config.ModuleSvc.UndeployModuleByNamePattern(client, pattern)
}

func init() {
	rootCmd.AddCommand(undeployModuleCmd)
	undeployModuleCmd.PersistentFlags().StringVarP(&params.ModuleName, action.ModuleName.Long, action.ModuleName.Short, "", action.ModuleName.Description)
	if err := undeployModuleCmd.RegisterFlagCompletionFunc(action.ModuleName.Long, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return helpers.GetBackendModuleNames(viper.GetStringMap(field.BackendModules)), cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		slog.Error(errors.RegisterFlagCompletionFailed(err).Error())
		os.Exit(1)
	}
	if err := undeployModuleCmd.MarkPersistentFlagRequired(action.ModuleName.Long); err != nil {
		slog.Error(errors.MarkFlagRequiredFailed(action.ModuleName, err).Error())
		os.Exit(1)
	}
}
