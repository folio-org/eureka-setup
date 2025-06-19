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

	"github.com/folio-org/eureka-cli/internal"
	"github.com/spf13/cobra"
)

const updateModuleDiscoveryCommand = "Update Module Discovery"

// updateModuleDiscoveryCmd represents the redirect command
var updateModuleDiscoveryCmd = &cobra.Command{
	Use:   "updateModuleDiscovery",
	Short: "Update module discovery",
	Long:  `Update module discovery to point to a different sidecar URL.`,
	Run: func(cmd *cobra.Command, args []string) {
		UpdateModuleDiscovery(withSidecarUrl)
	},
}

func UpdateModuleDiscovery(sidecarUrl string) {
	slog.Info(updateModuleDiscoveryCommand, internal.GetFuncName(), "### UPDATING MODULE DISCOVERY URL ###")
	internal.UpdateModuleDiscovery(updateModuleDiscoveryCommand, withEnableDebug, withId, sidecarUrl, withRestore, internal.DefaultServerPort)
}

func init() {
	rootCmd.AddCommand(updateModuleDiscoveryCmd)
	updateModuleDiscoveryCmd.PersistentFlags().StringVarP(&withId, "id", "i", "", "Module id, e.g. mod-orders:13.1.0-SNAPSHOT.1021 (required)")
	updateModuleDiscoveryCmd.PersistentFlags().StringVarP(&withSidecarUrl, "sidecarUrl", "s", "", "Sidecar URL e.g. http://host.docker.internal:37002")
	updateModuleDiscoveryCmd.PersistentFlags().BoolVarP(&withRestore, "restore", "r", false, "Restore sidecar URL")
	updateModuleDiscoveryCmd.MarkPersistentFlagRequired("id")
}
