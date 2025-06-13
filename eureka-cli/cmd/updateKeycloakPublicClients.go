/*
Copyright © 2025 EPAM_Systems/Thunderjet/Boburbek_Kadirkhodjaev

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

const updateKeycloakPublicClientsCommand string = "Update Keycloak Public Clients"

// updateKeycloakPublicClientsCmd represents the updateKeycloakPublicClients command
var updateKeycloakPublicClientsCmd = &cobra.Command{
	Use:   "updateKeycloakPublicClientParams",
	Short: "Update Keycloak public client params",
	Long:  `Update Keycloak public client params for each UI container.`,
	Run: func(cmd *cobra.Command, args []string) {
		UpdateKeycloakPublicClients()
	},
}

func UpdateKeycloakPublicClients() {
	slog.Info(updateKeycloakPublicClientsCommand, internal.GetFuncName(), "### UPDATING KEYCLOAK PUBLIC CLIENTS ###")
	keycloakMasterAccessToken := internal.GetKeycloakMasterAccessToken(updateKeycloakPublicClientsCommand, withEnableDebug)

	for _, value := range internal.GetTenants(updateKeycloakPublicClientsCommand, withEnableDebug, false) {
		mapEntry := value.(map[string]any)

		existingTenant := mapEntry["name"].(string)
		if !internal.HasTenant(existingTenant) || !internal.CanDeployUi(existingTenant) {
			continue
		}

		slog.Info(updateKeycloakPublicClientsCommand, internal.GetFuncName(), "Updating keycloak public client")
		internal.UpdateKeycloakPublicClientParams(updateKeycloakPublicClientsCommand, withEnableDebug, existingTenant, keycloakMasterAccessToken, platformCompleteExternalUrl)
	}
}

func init() {
	rootCmd.AddCommand(updateKeycloakPublicClientsCmd)
}
