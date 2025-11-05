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
	"time"

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/helpers"
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

		if len(run.Config.Action.ConfigApplicationDependencies) > 0 {
			err = run.UndeployChildApplication()
		} else {
			err = run.UndeployApplication()
		}
		if err != nil {
			return err
		}
		helpers.LogCompletion(run.Config.Action.Name, start)

		return nil
	},
}

func (run *Run) UndeployApplication() error {
	_ = run.UndeployUi()
	if err := run.UndeployModules(false); err != nil {
		return err
	}
	if err := run.UndeployManagement(); err != nil {
		return err
	}

	return run.UndeploySystem()
}

func (run *Run) UndeployChildApplication() error {
	if err := run.ConsortiumPartition(func(consortiumName string, tenantType constant.TenantType) error {
		return run.RemoveTenantEntitlements(consortiumName, tenantType)
	}); err != nil {
		return err
	}
	if err := run.UndeployModules(true); err != nil {
		return err
	}
	if err := run.UndeployAdditionalSystem(); err != nil {
		return err
	}
	return run.ConsortiumPartition(func(consortiumName string, tenantType constant.TenantType) error {
		if err := run.DetachCapabilitySets(consortiumName, tenantType); err != nil {
			return err
		}

		return run.AttachCapabilitySets(consortiumName, tenantType, 0*time.Second)
	})
}

func init() {
	rootCmd.AddCommand(undeployApplicationCmd)
	undeployApplicationCmd.PersistentFlags().BoolVarP(&actionParams.PurgeSchemas, "purgeSchemas", "z", false, "Purge schemas in PostgreSQL on uninstallation")
}
