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
	"strings"

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
		UpdateModuleDiscovery()
	},
}

func UpdateModuleDiscovery() {
	slog.Info(updateModuleDiscoveryCommand, internal.GetFuncName(), "### Update Module Discovery ###")
	id = strings.ReplaceAll(id, ":", "-")
	internal.UpdateModuleDiscovery(updateModuleDiscoveryCommand, enableDebug, id, location, restore, internal.DefaultServerPort)
}

func init() {
	rootCmd.AddCommand(updateModuleDiscoveryCmd)
	updateModuleDiscoveryCmd.PersistentFlags().StringVarP(&id, "id", "i", "", "Module id, e.g. mod-users:19.4.1-SNAPSHOT.323 (required)")
	updateModuleDiscoveryCmd.PersistentFlags().StringVarP(&location, "location", "l", "", "Location")
	updateModuleDiscoveryCmd.PersistentFlags().BoolVarP(&restore, "restore", "r", false, "Restore location")
	updateModuleDiscoveryCmd.MarkPersistentFlagRequired("id")
}
