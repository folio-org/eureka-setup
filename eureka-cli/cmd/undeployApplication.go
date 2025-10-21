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
	"github.com/folio-org/eureka-cli/field"
	"github.com/folio-org/eureka-cli/helpers"
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

		var err error
		r, err := New(action.UndeployApplication)
		if err != nil {
			return err
		}

		if len(viper.GetStringMap(field.ApplicationGatewayDependencies)) > 0 {
			err = r.UndeployChildApplication()
		} else {
			err = r.UndeployApplication()
		}
		if err != nil {
			return err
		}
		helpers.LogCompletion(r.Config.Action.Name, start)

		return nil
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
	err := r.PartitionErr(func(consortiumName string, tenantType constant.TenantType) error {
		return r.RemoveTenantEntitlements(consortiumName, tenantType)
	})
	if err != nil {
		return err
	}
	err = r.UndeployModules()
	if err != nil {
		return err
	}
	err = r.UndeployAdditionalSystem()
	if err != nil {
		return err
	}
	err = r.PartitionErr(func(consortiumName string, tenantType constant.TenantType) error {
		err := r.DetachCapabilitySets(consortiumName, tenantType)
		if err != nil {
			return err
		}
		err = r.AttachCapabilitySets(consortiumName, tenantType, 0*time.Second)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func init() {
	rootCmd.AddCommand(undeployApplicationCmd)
	undeployApplicationCmd.PersistentFlags().BoolVarP(&rp.PurgeSchemas, "purgeSchemas", "S", false, "Purge schemas in PostgreSQL on uninstallation")
}
