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

	"github.com/docker/docker/client"
	"github.com/folio-org/eureka-cli/internal"
	"github.com/spf13/cobra"
)

const getVaultRootTokenCommand string = "Get Vault Root Token"

// getVaultRootTokenCmd represents the getVaultRootToken command
var getVaultRootTokenCmd = &cobra.Command{
	Use:   "getVaultRootToken",
	Short: "Get vault root token",
	Long:  `Get vault root token from the server.`,
	Run: func(cmd *cobra.Command, args []string) {
		vaultRootToken := GetVaultRootToken()
		fmt.Println(vaultRootToken)
	},
}

func GetVaultRootToken() string {
	client := internal.CreateDockerClient(getVaultRootTokenCommand)
	defer client.Close()

	return internal.GetVaultRootToken(getVaultRootTokenCommand, client)
}

func GetVaultRootTokenWithDockerClient() (string, *client.Client) {
	client := internal.CreateDockerClient(getVaultRootTokenCommand)

	return internal.GetVaultRootToken(getVaultRootTokenCommand, client), client
}

func init() {
	rootCmd.AddCommand(getVaultRootTokenCmd)
}
