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
	"github.com/folio-org/eureka-cli/interceptmodulesvc"
	"github.com/folio-org/eureka-cli/models"
	"github.com/spf13/cobra"
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

// TODO Set to correct private port on restore
func (run *Run) InterceptModule() error {
	moduleDiscovery, err := run.Config.ManagementSvc.GetModuleDiscovery(run.Config.Action.Params.ModuleName)
	if err != nil {
		return err
	}
	if len(moduleDiscovery.Discovery) == 0 {
		return errors.ModuleDiscoveryNotFound(run.Config.Action.Params.ModuleName)
	}
	run.Config.Action.Params.ID = moduleDiscovery.Discovery[0].ID

	slog.Info(run.Config.Action.Name, "text", "INTERCEPTING MODULE", "module", run.Config.Action.Params.ModuleName, "id", run.Config.Action.Params.ID)
	backendModules, err := run.Config.ModuleParams.ReadBackendModulesFromConfig(false, false)
	if err != nil {
		return err
	}

	instalJsonURLs := run.Config.Action.GetCombinedInstallJsonURLs()
	registryModules, err := run.Config.RegistrySvc.GetModules(instalJsonURLs, false)
	if err != nil {
		return err
	}
	run.Config.RegistrySvc.ExtractModuleNameAndVersion(registryModules)

	client, err := run.Config.DockerClient.Create()
	if err != nil {
		return err
	}
	defer run.Config.DockerClient.Close(client)
	if err := run.setVaultRootTokenIntoContext(client); err != nil {
		return err
	}

	slog.Info(run.Config.Action.Name, "text", "UNDEPLOYING DEFAULT MODULE AND SIDECAR PAIR")
	pattern := fmt.Sprintf(constant.SingleModuleOrSidecarContainerPattern, run.Config.Action.ConfigProfile, run.Config.Action.Params.ModuleName)
	if err := run.Config.ModuleSvc.UndeployModuleByNamePattern(client, pattern); err != nil {
		return err
	}

	pair, err := interceptmodulesvc.NewModulePair(run.Config.Action, run.Config.Action.Params)
	if err != nil {
		return err
	}

	globalEnv := run.Config.Action.GetConfigEnvVars(field.Env)
	sidecarEnv := run.Config.Action.GetConfigEnvVars(field.SidecarModuleEnv)
	pair.Containers = models.NewCoreAndBusinessContainers(run.Config.Action.VaultRootToken, registryModules, backendModules, globalEnv, sidecarEnv)
	if err := run.Config.ManagementSvc.UpdateModuleDiscovery(actionParams.ID, *pair.SidecarURL, actionParams.Restore); err != nil {
		return err
	}
	if run.Config.Action.Params.Restore {
		return run.Config.InterceptModuleSvc.DeployDefaultModuleAndSidecarPair(pair, client)
	}

	return run.Config.InterceptModuleSvc.DeployCustomSidecarForInterception(pair, client)
}

func init() {
	rootCmd.AddCommand(interceptModuleCmd)
	interceptModuleCmd.PersistentFlags().StringVarP(&actionParams.ModuleName, "moduleName", "n", "", "Module name, e.g. mod-orders")
	interceptModuleCmd.PersistentFlags().StringVarP(&actionParams.ModuleURL, "moduleUrl", "m", "", "Module URL, e.g. http://host.docker.internal:36002 or 36002 (if -g is used)")
	interceptModuleCmd.PersistentFlags().StringVarP(&actionParams.SidecarURL, "sidecarUrl", "s", "", "Sidecar URL e.g. http://host.docker.internal:37002 or 37002 (if -g is used)")
	interceptModuleCmd.PersistentFlags().BoolVarP(&actionParams.Restore, "restore", "r", false, "Restore module & sidecar")
	interceptModuleCmd.PersistentFlags().BoolVarP(&actionParams.DefaultGateway, "defaultGateway", "g", false, "Use default gateway in URLs, .e.g http://host.docker.internal:{{port}} will be set automatically")
	if err := interceptModuleCmd.MarkPersistentFlagRequired("moduleName"); err != nil {
		slog.Error("failed to mark moduleName flag as required", "error", err)
		os.Exit(1)
	}
}
