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
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := New(action.GetVaultRootToken)
		if err != nil {
			return err
		}

		err = r.GetVaultRootToken()
		if err != nil {
			return err
		}
		fmt.Println(r.RunConfig.Action.KeycloakAccessToken)

		return nil
	},
}

func (r *Run) GetVaultRootToken() error {
	client, err := r.RunConfig.DockerClient.Create()
	if err != nil {
		return err
	}
	defer r.RunConfig.DockerClient.Close(client)

	err = r.setVaultRootTokenIntoContext(client)
	if err != nil {
		return err

	}

	return nil
}

func (r *Run) setVaultRootTokenIntoContext(client *client.Client) error {
	vaultRootToken, err := r.RunConfig.ModuleSvc.GetVaultRootToken(client)
	if err != nil {
		return err
	}
	r.RunConfig.Action.VaultRootToken = vaultRootToken

	return nil
}

func init() {
	rootCmd.AddCommand(getVaultRootTokenCmd)
}
