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
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/errors"
	"github.com/spf13/cobra"
)

// getKeycloakAccessTokenCmd represents the getAccessToken command
var getKeycloakAccessTokenCmd = &cobra.Command{
	Use:   "getKeycloakAccessToken",
	Short: "Get keycloak access token",
	Long:  `Get a keycloak master access token.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		run, err := New(action.GetKeycloakAccessToken)
		if err != nil {
			return err
		}
		if err := run.GetVaultRootToken(); err != nil {
			return err
		}

		accessToken, err := run.GetKeycloakAccessToken(params.TokenType, params.Tenant)
		if err != nil {
			return err
		}
		fmt.Println(accessToken)

		return nil
	},
}

func (run *Run) GetKeycloakAccessToken(tokenType, tenant string) (string, error) {
	switch tokenType {
	case constant.MasterCustomToken:
		if err := run.setKeycloakMasterAccessTokenIntoContext(constant.ClientCredentials); err != nil {
			return "", err
		}

		return run.Config.Action.KeycloakMasterAccessToken, nil
	case constant.MasterAdminCLIToken:
		if err := run.setKeycloakMasterAccessTokenIntoContext(constant.Password); err != nil {
			return "", err
		}

		return run.Config.Action.KeycloakMasterAccessToken, nil
	default:
		if tenant == "" {
			return "", errors.RequiredParameterMissing("tenant")
		}
		if err := run.setKeycloakAccessTokenIntoContext(tenant); err != nil {
			return "", err
		}

		return run.Config.Action.KeycloakAccessToken, nil
	}
}

func (run *Run) setKeycloakAccessTokenIntoContext(tenant string) error {
	accessToken, err := run.Config.KeycloakSvc.GetAccessToken(tenant)
	if err != nil {
		return err
	}
	run.Config.Action.KeycloakAccessToken = accessToken

	return nil
}

func (run *Run) setKeycloakMasterAccessTokenIntoContext(grantType constant.KeycloakGrantType) error {
	accessToken, err := run.Config.KeycloakSvc.GetMasterAccessToken(grantType)
	if err != nil {
		return err
	}
	run.Config.Action.KeycloakMasterAccessToken = accessToken

	return nil
}

func init() {
	rootCmd.AddCommand(getKeycloakAccessTokenCmd)
	getKeycloakAccessTokenCmd.PersistentFlags().StringVarP(&params.Tenant, action.Tenant.Long, action.Tenant.Short, "", action.Tenant.Description)
	getKeycloakAccessTokenCmd.PersistentFlags().StringVarP(&params.TokenType, action.TokenType.Long, action.TokenType.Short, "", action.TokenType.Description)
	if err := getKeycloakAccessTokenCmd.RegisterFlagCompletionFunc(action.TokenType.Long, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return constant.GetTokenTypes(), cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		slog.Error(errors.RegisterFlagCompletionFailed(err).Error())
		os.Exit(1)
	}
}
