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
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/big"
	"os"

	"github.com/j011195/eureka-setup/eureka-cli/action"
	"github.com/j011195/eureka-setup/eureka-cli/constant"
	"github.com/j011195/eureka-setup/eureka-cli/errors"
	"github.com/spf13/cobra"
)

// getEdgeApiKeyCmd represents the generateApiKey command
var getEdgeApiKeyCmd = &cobra.Command{
	Use:   "getEdgeApiKey",
	Short: "Get Edge API key",
	Long:  `Get Edge API key for a tenant`,
	RunE: func(cmd *cobra.Command, args []string) error {
		run, err := New(action.GetEdgeApiKey)
		if err != nil {
			return err
		}

		return run.GetEdgeApiKey()
	},
}

func (run *Run) GetEdgeApiKey() error {
	randomStr, err := run.getRandomString(params.Length)
	if err != nil {
		return err
	}

	payload, err := json.Marshal(map[string]any{
		"s": randomStr,
		"t": params.Tenant,
		"u": params.User,
	})
	if err != nil {
		return err
	}
	apiKey := base64.URLEncoding.EncodeToString(payload)

	fmt.Println(apiKey)

	return nil
}

func (run *Run) getRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	for i := range length {
		charsetIdx, err := rand.Int(rand.Reader, big.NewInt(int64(len(constant.Charset))))
		if err != nil {
			return "", err
		}
		bytes[i] = constant.Charset[charsetIdx.Int64()]
	}

	return string(bytes), nil
}

func init() {
	rootCmd.AddCommand(getEdgeApiKeyCmd)
	getEdgeApiKeyCmd.PersistentFlags().StringVarP(&params.Tenant, action.Tenant.Long, action.Tenant.Short, "", action.Tenant.Description)
	getEdgeApiKeyCmd.PersistentFlags().StringVarP(&params.User, action.User.Long, action.User.Short, "", action.User.Description)
	getEdgeApiKeyCmd.PersistentFlags().IntVarP(&params.Length, action.Length.Long, action.Length.Short, 17, action.Length.Description)

	if err := getEdgeApiKeyCmd.MarkPersistentFlagRequired(action.Tenant.Long); err != nil {
		slog.Error(errors.MarkFlagRequiredFailed(action.Tenant, err).Error())
		os.Exit(1)
	}
	if err := getEdgeApiKeyCmd.MarkPersistentFlagRequired(action.User.Long); err != nil {
		slog.Error(errors.MarkFlagRequiredFailed(action.User, err).Error())
		os.Exit(1)
	}
}
