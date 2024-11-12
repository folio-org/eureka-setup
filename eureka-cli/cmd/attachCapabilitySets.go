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
)

const attachCapabilitySetsCommand string = "Attach Capability Sets"

// attachCapabilitySetsCmd represents the attachCapabilitySets command
var attachCapabilitySetsCmd = &cobra.Command{
	Use:   "attachCapabilitySets",
	Short: "Attach capability sets",
	Long:  `Attach capability sets to roles.`,
	Run: func(cmd *cobra.Command, args []string) {
		AttachCapabilitySets(false)
	},
}

func AttachCapabilitySets(ranInSequence bool) {
	if ranInSequence {
		slog.Info(attachCapabilitySetsCommand, internal.GetFuncName(), "### WAITING FOR CAPABILITY AND CAPABILITY SETS TO SYNCHRONIZE ###")
		time.Sleep(60 * time.Second)
	}

	slog.Info(attachCapabilitySetsCommand, internal.GetFuncName(), "### ACQUIRING VAULT ROOT TOKEN ###")
	client := internal.CreateClient(attachCapabilitySetsCommand)
	defer client.Close()
	vaultRootToken := internal.GetRootVaultToken(attachCapabilitySetsCommand, client)

	for _, tenantValue := range internal.GetTenants(attachCapabilitySetsCommand, enableDebug, false) {
		tenantMapEntry := tenantValue.(map[string]interface{})
		tenant := tenantMapEntry["name"].(string)

		if !internal.HasTenant(tenant) {
			continue
		}

		slog.Info(attachCapabilitySetsCommand, internal.GetFuncName(), "### ACQUIRING KEYCLOAK ACCESS TOKEN ###")
		accessToken := internal.GetKeycloakAccessToken(attachCapabilitySetsCommand, enableDebug, vaultRootToken, tenant)

		slog.Info(attachCapabilitySetsCommand, internal.GetFuncName(), "### ATTACHING CAPABILITY SETS TO ROLES ###")
		internal.AttachCapabilitySetsToRoles(attachCapabilitySetsCommand, enableDebug, tenant, accessToken)
	}
}

func init() {
	rootCmd.AddCommand(attachCapabilitySetsCmd)
}
