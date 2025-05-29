/*
Copyright © 2024 EPAM_Systems/Thunderjet/Boburbek_Kadirkhodjaev

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
	"time"

	"github.com/folio-org/eureka-cli/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const undeployApplicationCommand string = "Undeploy Application"

// undeployApplicationCmd represents the undeployApplication command
var undeployApplicationCmd = &cobra.Command{
	Use:   "undeployApplication",
	Short: "Undeploy application",
	Long:  `Undeploy platform application.`,
	Run: func(cmd *cobra.Command, args []string) {
		start := time.Now()
		if len(viper.GetStringMap(internal.ApplicationGatewayDependenciesKey)) > 0 {
			UndeployChildApplication()
		} else {
			UndeployApplication()
		}
		slog.Info(undeployApplicationCommand, "Elapsed, duration", time.Since(start))
	},
}

func UndeployApplication() {
	UndeployUi()
	UndeployModules()
	UndeployManagement()
	UndeploySystem()
}

func UndeployChildApplication() {
	RemoveTenantEntitlements()
	UndeployModules()
	UndeployAdditionalSystem()
	DetachCapabilitySets()
	AttachCapabilitySets()
}

func init() {
	rootCmd.AddCommand(undeployApplicationCmd)
	undeployApplicationCmd.PersistentFlags().BoolVarP(&withPurgeSchemas, "purgeSchemas", "S", false, "Purge schemas in PostgreSQL on uninstallation")
}
