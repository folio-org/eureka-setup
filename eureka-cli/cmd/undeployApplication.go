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
	"time"

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/field"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// undeployApplicationCmd represents the undeployApplication command
var undeployApplicationCmd = &cobra.Command{
	Use:   "undeployApplication",
	Short: "Undeploy application",
	Long:  `Undeploy platform application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		start := time.Now()

		r, err := New(action.UndeployApplication)
		if err != nil {
			return err
		}

		if len(viper.GetStringMap(field.ApplicationGatewayDependencies)) > 0 {
			r.UndeployChildApplication()
		} else {
			r.UndeployApplication()
		}

		slog.Info(r.Config.Action.Name, "text", fmt.Sprintf("Elapsed, duration %.1f", time.Since(start).Minutes()))

		return err
	},
}

func (r *Run) UndeployApplication() error {
	err := r.UndeployUi()
	if err != nil {
		return err
	}

	err = r.UndeployModules()
	if err != nil {
		return err
	}

	err = r.UndeployManagement()
	if err != nil {
		return err
	}

	err = r.UndeploySystem()
	if err != nil {
		return err
	}

	return nil
}

func (r *Run) UndeployChildApplication() error {
	r.Partition(func(consortiumName string, tenantType constant.TenantType) {
		r.RemoveTenantEntitlements(consortiumName, tenantType)
	})

	r.UndeployModules()

	err := r.UndeployAdditionalSystem()
	if err != nil {
		return err
	}

	r.PartitionErr(func(consortiumName string, tenantType constant.TenantType) error {
		r.DetachCapabilitySets(consortiumName, tenantType)

		err := r.AttachCapabilitySets(consortiumName, tenantType, 0*time.Second)
		if err != nil {
			return err
		}

		return nil
	})

	return nil
}

func init() {
	rootCmd.AddCommand(undeployApplicationCmd)
	undeployApplicationCmd.PersistentFlags().BoolVarP(&rp.PurgeSchemas, "purgeSchemas", "S", false, "Purge schemas in PostgreSQL on uninstallation")
}
