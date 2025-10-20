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
	"github.com/spf13/cobra"
)

// updateModuleDiscoveryCmd represents the redirect command
var updateModuleDiscoveryCmd = &cobra.Command{
	Use:   "updateModuleDiscovery",
	Short: "Update module discovery",
	Long:  `Update module discovery to point to a different sidecar URL.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := New(action.UpdateModuleDiscovery)
		if err != nil {
			return err
		}

		r.UpdateModuleDiscovery(rp.SidecarURL)

		return nil
	},
}

func (r *Run) UpdateModuleDiscovery(sidecarUrl string) {
	slog.Info(r.Config.Action.Name, "text", "UPDATING MODULE DISCOVERY URL")
	r.Config.ManagementStep.UpdateModuleDiscovery(rp.ID, sidecarUrl, rp.Restore, constant.ServerPort)
}

func init() {
	rootCmd.AddCommand(updateModuleDiscoveryCmd)
	updateModuleDiscoveryCmd.PersistentFlags().StringVarP(&rp.ID, "id", "i", "", "Module id, e.g. mod-orders:13.1.0-SNAPSHOT.1021 (required)")
	updateModuleDiscoveryCmd.PersistentFlags().StringVarP(&rp.SidecarURL, "sidecarUrl", "s", "", "Sidecar URL e.g. http://host.docker.internal:37002")
	updateModuleDiscoveryCmd.PersistentFlags().BoolVarP(&rp.Restore, "restore", "r", false, "Restore sidecar URL")
	if err := updateModuleDiscoveryCmd.MarkPersistentFlagRequired("id"); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
