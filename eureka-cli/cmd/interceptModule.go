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
	"log/slog"
	"os"

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/errors"
	"github.com/folio-org/eureka-cli/field"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/folio-org/eureka-cli/interceptmodulesvc"
	"github.com/folio-org/eureka-cli/models"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// interceptModuleCmd represents the interceptModule command
var interceptModuleCmd = &cobra.Command{
	Use:   "interceptModule",
	Short: "Intercept module",
	Long:  `Intercept/redirect module traffic to IntelliJ.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		run, err := New(action.InterceptModule)
		if err != nil {
			return err
		}

		return run.InterceptModule()
	},
}

func (run *Run) InterceptModule() error {
	if err := run.setKeycloakMasterAccessTokenIntoContext(constant.ClientCredentials); err != nil {
		return err
	}
	if err := run.setModuleDiscoveryDataIntoContext(); err != nil {
		return err
	}

	slog.Info(run.Config.Action.Name, "text", "INTERCEPTING MODULE", "module", params.ModuleName, "id", params.ID)
	backendModules, err := run.Config.ModuleProps.ReadBackendModules(false, false)
	if err != nil {
		return err
	}

	instalJsonURLs := run.Config.Action.GetCombinedInstallJsonURLs()
	modules, err := run.Config.RegistrySvc.GetModules(instalJsonURLs, true, false)
	if err != nil {
		return err
	}
	run.Config.RegistrySvc.ExtractModuleMetadata(modules)

	client, err := run.Config.DockerClient.Create()
	if err != nil {
		return err
	}
	defer run.Config.DockerClient.Close(client)
	if err := run.setVaultRootTokenIntoContext(client); err != nil {
		return err
	}

	pair, err := interceptmodulesvc.NewModulePair(run.Config.Action, run.Config.Action.Param)
	if err != nil {
		return err
	}

	pair.Containers = &models.Containers{
		Modules:        modules,
		BackendModules: backendModules,
		IsManagement:   false,
	}
	if params.Restore {
		return run.Config.InterceptModuleSvc.DeployDefaultModuleAndSidecarPair(pair, client)
	} else {
		return run.Config.InterceptModuleSvc.DeployCustomSidecarForInterception(pair, client)
	}
}

func init() {
	rootCmd.AddCommand(interceptModuleCmd)
	interceptModuleCmd.PersistentFlags().StringVarP(&params.ModuleName, action.ModuleName.Long, action.ModuleName.Short, "", action.ModuleName.Description)
	if err := interceptModuleCmd.RegisterFlagCompletionFunc(action.ModuleName.Long, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return helpers.GetBackendModuleNames(viper.GetStringMap(field.BackendModules)), cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		slog.Error(errors.RegisterFlagCompletionFailed(err).Error())
		os.Exit(1)
	}
	interceptModuleCmd.PersistentFlags().StringVarP(&params.ModuleURL, action.ModuleURL.Long, action.ModuleURL.Short, "", action.ModuleURL.Description)
	interceptModuleCmd.PersistentFlags().StringVarP(&params.SidecarURL, action.SidecarURL.Long, action.SidecarURL.Short, "", action.SidecarURL.Description)
	interceptModuleCmd.PersistentFlags().BoolVarP(&params.Restore, action.Restore.Long, action.Restore.Short, false, action.Restore.Description)
	interceptModuleCmd.PersistentFlags().BoolVarP(&params.DefaultGateway, action.DefaultGateway.Long, action.DefaultGateway.Short, false, action.DefaultGateway.Description)
	interceptModuleCmd.PersistentFlags().BoolVarP(&params.SkipRegistry, action.SkipRegistry.Long, action.SkipRegistry.Short, false, action.SkipRegistry.Description)
	if err := interceptModuleCmd.MarkPersistentFlagRequired(action.ModuleName.Long); err != nil {
		slog.Error(errors.MarkFlagRequiredFailed(action.ModuleName, err).Error())
		os.Exit(1)
	}
}
