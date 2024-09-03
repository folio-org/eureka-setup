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
	"slices"

	"github.com/folio-org/eureka-cli/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	slog.Info(detachCapabilitySetsCommand, "### ACQUIRING VAULT ROOT TOKEN ###", "")
	client := internal.CreateClient(detachCapabilitySetsCommand)
	defer client.Close()
	vaultRootToken := internal.GetRootVaultToken(detachCapabilitySetsCommand, client)

	for _, value := range internal.GetTenants(detachCapabilitySetsCommand, enableDebug, false) {
		mapEntry := value.(map[string]interface{})
		tenant := mapEntry["name"].(string)

		if !slices.Contains(viper.GetStringSlice(internal.TenantsKey), tenant) {
			continue
		}

		slog.Info(detachCapabilitySetsCommand, "### ACQUIRING KEYCLOAK ACCESS TOKEN ###", "")
		accessToken := internal.GetKeycloakAccessToken(detachCapabilitySetsCommand, enableDebug, vaultRootToken, tenant)

		slog.Info(detachCapabilitySetsCommand, "### DETACHING CAPABILITY SETS FROM ROLES ###", "")
		internal.DetachCapabilitySetsFromRoles(detachCapabilitySetsCommand, enableDebug, false, tenant, accessToken)
	}
}

func init() {
	rootCmd.AddCommand(detachCapabilitySetsCmd)
}
