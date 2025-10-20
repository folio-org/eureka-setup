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
	Run: func(cmd *cobra.Command, args []string) {
		start := time.Now()

		r := NewRun(action.UndeployApplication)
		if len(viper.GetStringMap(field.ApplicationGatewayDependencies)) > 0 {
			r.UndeployChildApplication()
		} else {
			r.UndeployApplication()
		}

		slog.Info(r.Config.Action.Name, "text", fmt.Sprintf("Elapsed, duration %.1f", time.Since(start).Minutes()))
	},
}

func (r *Run) UndeployApplication() {
	r.UndeployUi()
	r.UndeployModules()
	r.UndeployManagement()

	err := r.UndeploySystem()
	if err != nil {
		slog.Error(r.Config.Action.Name, "error", err.Error())
		os.Exit(1)
	}
}

func (r *Run) UndeployChildApplication() {
	r.Partition(func(consortiumName string, tenantType constant.TenantType) {
		r.RemoveTenantEntitlements(consortiumName, tenantType)
	})

	r.UndeployModules()

	err := r.UndeployAdditionalSystem()
	if err != nil {
		slog.Error(r.Config.Action.Name, "error", err.Error())
		os.Exit(1)
	}
	r.Partition(func(consortiumName string, tenantType constant.TenantType) {
		r.DetachCapabilitySets(consortiumName, tenantType)

		err := r.AttachCapabilitySets(consortiumName, tenantType, 0*time.Second)
		if err != nil {
			slog.Error(r.Config.Action.Name, "error", err.Error())
			os.Exit(1)
		}
	})
}

func init() {
	rootCmd.AddCommand(undeployApplicationCmd)
	undeployApplicationCmd.PersistentFlags().BoolVarP(&rp.PurgeSchemas, "purgeSchemas", "S", false, "Purge schemas in PostgreSQL on uninstallation")
}
