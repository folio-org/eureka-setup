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

	"github.com/folio-org/eureka-cli/action"
	"github.com/spf13/cobra"
)

// getKeycloakAccessTokenCmd represents the getAccessToken command
var getKeycloakAccessTokenCmd = &cobra.Command{
	Use:   "getKeycloakAccessToken",
	Short: "Get keycloak access token",
	Long:  `Get a keycloak master access token.`,
	Run: func(cmd *cobra.Command, args []string) {
		r := NewRun(action.GetKeycloakAccessToken)
		vaultRootToken := r.GetVaultRootToken()
		r.GetKeycloakAccessToken(vaultRootToken)
	},
}

func (r *Run) GetKeycloakAccessToken(vaultRootToken string) {
	keycloakAccessToken := r.Config.KeycloakStep.GetKeycloakAccessToken(vaultRootToken, rp.Tenant)
	fmt.Println(keycloakAccessToken)
}

func init() {
	rootCmd.AddCommand(getKeycloakAccessTokenCmd)
	getKeycloakAccessTokenCmd.PersistentFlags().StringVarP(&rp.Tenant, "tenant", "t", "", "Tenant (required)")
	if err := getKeycloakAccessTokenCmd.MarkPersistentFlagRequired("tenant"); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
