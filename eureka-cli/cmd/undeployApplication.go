/*
Copyright © 2026 Open Library Foundation

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
	"time"

	"github.com/folio-org/eureka-setup/eureka-cli/action"
	"github.com/folio-org/eureka-setup/eureka-cli/constant"
	"github.com/folio-org/eureka-setup/eureka-cli/helpers"
	"github.com/spf13/cobra"
)

// undeployApplicationCmd represents the undeployApplication command
var undeployApplicationCmd = &cobra.Command{
	Use:     "undeployApplication",
	Short:   "Undeploy application",
	Long:    `Undeploy platform application.`,
	Aliases: []string{"undeploy"},
	RunE: func(cmd *cobra.Command, args []string) error {
		start := time.Now()

		var err error
		run, err := New(action.UndeployApplication)
		if err != nil {
			return err
		}

		switch {
		case cmd.Flags().Changed(action.ApplicationName.Long):
			err = run.UndeployLocalApplication()
		case run.Config.Action.IsChildApp():
			err = run.UndeployChildApplication()
		default:
			err = run.UndeployApplication()
		}
		if err != nil {
			return err
		}
		slog.Info(run.Config.Action.Name, "text", "Command completed", "duration", time.Since(start))

		return nil
	},
}

func (run *Run) UndeployApplication() error {
	if err := run.UndeployUI(); err != nil {
		slog.Warn(run.Config.Action.Name, "text", "UI undeploy was unsuccessful", "error", err)
	}
	if err := run.UndeployModules(false); err != nil {
		return err
	}
	if err := run.UndeployManagement(); err != nil {
		return err
	}

	return run.UndeploySystem()
}

// UndeployLocalApplication tears down a local application created by runLocalModule
func (run *Run) UndeployLocalApplication() error {
	if err := run.setKeycloakMasterAccessTokenIntoContext(constant.ClientCredentials); err != nil {
		return err
	}
	appName := params.ApplicationName

	app, err := run.Config.ManagementSvc.GetLatestApplicationByName(appName)
	if err != nil {
		return err
	}
	if app == nil {
		slog.Info(run.Config.Action.Name, "text", "No local application found, nothing to undeploy", "name", appName)
		return nil
	}
	appID := helpers.GetString(app, "id")

	slog.Info(run.Config.Action.Name, "text", "UNDEPLOYING LOCAL APPLICATION", "name", appName, "id", appID)
	if err := run.Config.ManagementSvc.RemoveTenantEntitlementsForApplication(constant.NoneConsortium, constant.All, appID, params.PurgeSchemas); err != nil {
		slog.Warn(run.Config.Action.Name, "text", "Remove local application tenant entitlements was unsuccessful", "error", err)
	}

	client, err := run.Config.DockerClient.Create()
	if err != nil {
		return err
	}
	defer run.Config.DockerClient.Close(client)

	for _, value := range helpers.GetAnySlice(app, "modules") {
		entry, ok := value.(map[string]any)
		if !ok {
			continue
		}
		moduleName := helpers.GetString(entry, "name")
		moduleID := helpers.GetString(entry, "id")

		if err := run.Config.ManagementSvc.RemoveModuleDiscovery(moduleID); err != nil {
			slog.Warn(run.Config.Action.Name, "text", "Remove module discovery was unsuccessful", "module", moduleName, "error", err)
		}

		pattern := fmt.Sprintf(constant.SingleModuleOrSidecarContainerPattern, run.Config.Action.ConfigProfileName, moduleName)
		if err := run.Config.ModuleSvc.UndeployModuleByNamePattern(client, pattern); err != nil {
			slog.Warn(run.Config.Action.Name, "text", "Undeploy module containers was unsuccessful", "module", moduleName, "error", err)
		}
	}

	slog.Info(run.Config.Action.Name, "text", "REMOVING LOCAL APPLICATIONS", "name", appName)
	return run.Config.ManagementSvc.RemoveApplications(appName, "")
}

func (run *Run) UndeployChildApplication() error {
	if err := run.ConsortiumPartition(func(consortiumName string, tenantType constant.TenantType) error {
		if err := run.RemoveTenantEntitlements(consortiumName, tenantType); err != nil {
			slog.Warn(run.Config.Action.Name, "text", "Remove tenant entitlement was unsuccessful", "error", err)
		}

		return nil
	}); err != nil {
		return err
	}
	if err := run.UndeployModules(true); err != nil {
		return err
	}
	if err := run.UndeployAdditionalSystem(); err != nil {
		return err
	}
	if params.SkipCapabilitySets {
		return nil
	}
	return run.ConsortiumPartition(func(consortiumName string, tenantType constant.TenantType) error {
		if err := run.DetachCapabilitySets(consortiumName, tenantType); err != nil {
			return err
		}

		return run.AttachCapabilitySets(consortiumName, tenantType, 0*time.Second, false)
	})
}

func init() {
	rootCmd.AddCommand(undeployApplicationCmd)
	undeployApplicationCmd.PersistentFlags().StringVarP(&params.ApplicationName, action.ApplicationName.Long, action.ApplicationName.Short, "", action.ApplicationName.Description)
	undeployApplicationCmd.PersistentFlags().BoolVarP(&params.PurgeSchemas, action.PurgeSchemas.Long, action.PurgeSchemas.Short, false, action.PurgeSchemas.Description)
	undeployApplicationCmd.PersistentFlags().BoolVarP(&params.SkipCapabilitySets, action.SkipCapabilitySets.Long, action.SkipCapabilitySets.Short, false, action.SkipCapabilitySets.Description)
}
