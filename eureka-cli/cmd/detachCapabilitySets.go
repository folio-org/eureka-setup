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
	"fmt"
	"log/slog"

	"github.com/folio-org/eureka-cli/internal"
	"github.com/spf13/cobra"
)

const detachCapabilitySetsCommand string = "Detach Capability Sets"

// detachCapabilitySetsCmd represents the detachCapabilitySets command
var detachCapabilitySetsCmd = &cobra.Command{
	Use:   "detachCapabilitySets",
	Short: "Detach capability sets",
	Long:  `Detach all capability sets from roles.`,
	Run: func(cmd *cobra.Command, args []string) {
		DetachCapabilitySets()
	},
}

func DetachCapabilitySets() {
	vaultRootToken := GetVaultRootToken()

	for _, value := range internal.GetTenants(detachCapabilitySetsCommand, withEnableDebug, false) {
		mapEntry := value.(map[string]any)

		existingTenant := mapEntry["name"].(string)
		if !internal.HasTenant(existingTenant) {
			continue
		}

		slog.Info(detachCapabilitySetsCommand, internal.GetFuncName(), fmt.Sprintf("### DETACHING CAPABILITY SETS FROM ROLES FOR %s TENANT ###", existingTenant))
		keycloakAccessToken := internal.GetKeycloakAccessToken(detachCapabilitySetsCommand, withEnableDebug, vaultRootToken, existingTenant)
		internal.DetachCapabilitySetsFromRoles(detachCapabilitySetsCommand, withEnableDebug, false, existingTenant, keycloakAccessToken)
	}
}

func init() {
	rootCmd.AddCommand(detachCapabilitySetsCmd)
}
