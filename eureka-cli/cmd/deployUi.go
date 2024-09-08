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

const deployUiCommand string = "Deploy UI"

// deployUiCmd represents the deployUi command
var deployUiCmd = &cobra.Command{
	Use:   "deployUi",
	Short: "Deploy UI",
	Long:  `Deploy the UI container.`,
	Run: func(cmd *cobra.Command, args []string) {
		DeployUi()
	},
}

func DeployUi() {
	// TODO Deploy UI module

	slog.Info(createUsersCommand, "### ACQUIRING KEYCLOAK MASTER ACCESS TOKEN ###", "")
	masterAccessToken := internal.GetKeycloakMasterRealmAccessToken(createUsersCommand, enableDebug)

	for _, value := range internal.GetTenants(deployUiCommand, enableDebug, false) {
		mapEntry := value.(map[string]interface{})
		tenant := mapEntry["name"].(string)

		if !slices.Contains(viper.GetStringSlice(internal.TenantsKey), tenant) {
			continue
		}

		slog.Info(deployUiCommand, "### UPDATING KEYCLOAK PUBLIC CLIENT", "")
		internal.UpdateKeycloakPublicClientParams(deployUiCommand, enableDebug, tenant, masterAccessToken)
	}

}

func init() {
	rootCmd.AddCommand(deployUiCmd)
}
