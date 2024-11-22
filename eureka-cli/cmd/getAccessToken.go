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

const getAccessTokenCommand string = "Get Access Token"

var (
	tenant string
)

// getAccessTokenCmd represents the getAccessToken command
var getAccessTokenCmd = &cobra.Command{
	Use:   "getAccessToken",
	Short: "Get access token",
	Long:  `Get a master access token from a particular realm.`,
	Run: func(cmd *cobra.Command, args []string) {
		GetAccessToken()
	},
}

func GetAccessToken() {
	slog.Info(getAccessTokenCommand, internal.GetFuncName(), "### ACQUIRING VAULT ROOT TOKEN ###")
	client := internal.CreateClient(getAccessTokenCommand)
	defer client.Close()
	vaultRootToken := internal.GetRootVaultToken(getAccessTokenCommand, client)

	slog.Info(getAccessTokenCommand, internal.GetFuncName(), "### ACQUIRING KEYCLOAK ACCESS TOKEN ###")
	accessToken := internal.GetKeycloakAccessToken(createUsersCommand, enableDebug, vaultRootToken, tenant)
	slog.Info(getAccessTokenCommand, internal.GetFuncName(), fmt.Sprintf("Access token: %s", accessToken))
}

func init() {
	rootCmd.AddCommand(getAccessTokenCmd)
	getAccessTokenCmd.PersistentFlags().StringVarP(&tenant, "realm", "r", "", "Realm (required)")
	getAccessTokenCmd.MarkPersistentFlagRequired("realm")
}
