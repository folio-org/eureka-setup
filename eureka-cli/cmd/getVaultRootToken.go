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

	"github.com/docker/docker/client"
	"github.com/folio-org/eureka-cli/action"
	"github.com/spf13/cobra"
)

// getVaultRootTokenCmd represents the getVaultRootToken command
var getVaultRootTokenCmd = &cobra.Command{
	Use:   "getVaultRootToken",
	Short: "Get vault root token",
	Long:  `Get vault root token from the server.`,
	Run: func(cmd *cobra.Command, args []string) {
		vaultRootToken := New(action.GetVaultRootToken).GetVaultRootToken()
		fmt.Println(vaultRootToken)
	},
}

func (r *Run) GetVaultRootToken() string {
	client := r.Config.DockerClient.Create()
	defer func() {
		_ = client.Close()
	}()

	return r.Config.ModuleStep.GetVaultRootToken(client)
}

func (r *Run) GetVaultRootTokenWithDockerClient() (string, *client.Client) {
	client := r.Config.DockerClient.Create()

	return r.Config.ModuleStep.GetVaultRootToken(client), client
}

func init() {
	rootCmd.AddCommand(getVaultRootTokenCmd)
}
