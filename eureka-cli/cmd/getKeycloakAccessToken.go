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

	"github.com/folio-org/eureka-cli/internal"
	"github.com/spf13/cobra"
)

const getKeycloakAccessTokenCommand string = "Get Keycloak Access Token"

// getKeycloakAccessTokenCmd represents the getAccessToken command
var getKeycloakAccessTokenCmd = &cobra.Command{
	Use:   "getKeycloakAccessToken",
	Short: "Get keyclaok access token",
	Long:  `Get a keycloak master access token.`,
	Run: func(cmd *cobra.Command, args []string) {
		vaultRootToken := GetVaultRootToken()
		GetKeycloakAccessToken(vaultRootToken)
	},
}

func GetKeycloakAccessToken(vaultRootToken string) {
	keycloakAccessToken := internal.GetKeycloakAccessToken(getKeycloakAccessTokenCommand, withEnableDebug, vaultRootToken, withTenant)
	fmt.Println(keycloakAccessToken)
}

func init() {
	rootCmd.AddCommand(getKeycloakAccessTokenCmd)
	getKeycloakAccessTokenCmd.PersistentFlags().StringVarP(&withTenant, "tenant", "t", "", "Tenant (required)")
	getKeycloakAccessTokenCmd.MarkPersistentFlagRequired("tenant")
}
