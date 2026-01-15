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
	"github.com/folio-org/eureka-setup/eureka-cli/action"
	"github.com/spf13/cobra"
)

// getVaultRootTokenCmd represents the getVaultRootToken command
var getVaultRootTokenCmd = &cobra.Command{
	Use:   "getVaultRootToken",
	Short: "Get vault root token",
	Long:  `Get vault root token from the server.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		run, err := New(action.GetVaultRootToken)
		if err != nil {
			return err
		}
		if err := run.GetVaultRootToken(); err != nil {
			return err
		}
		fmt.Println(run.Config.Action.VaultRootToken)

		return nil
	},
}

func (run *Run) GetVaultRootToken() error {
	client, err := run.Config.DockerClient.Create()
	if err != nil {
		return err
	}
	defer run.Config.DockerClient.Close(client)

	return run.setVaultRootTokenIntoContext(client)
}

func (run *Run) setVaultRootTokenIntoContext(client *client.Client) error {
	rootToken, err := run.Config.ModuleSvc.GetVaultRootToken(client)
	if err != nil {
		return err
	}
	run.Config.Action.VaultRootToken = rootToken

	return nil
}

func init() {
	rootCmd.AddCommand(getVaultRootTokenCmd)
}
