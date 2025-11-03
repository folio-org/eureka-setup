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

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/spf13/cobra"
)

// getEdgeApiKeyCmd represents the generateApiKey command
var getEdgeApiKeyCmd = &cobra.Command{
	Use:   "getEdgeApiKey",
	Short: "Get Edge API key",
	Long:  `Get Edge API key for a tenant`,
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := New(action.GetEdgeApiKey)
		if err != nil {
			return err
		}

		return r.GetEdgeApiKey()
	},
}

func (r *Run) GetEdgeApiKey() error {
	randomStr, err := r.getRandomString(actionParams.Length)
	if err != nil {
		return err
	}

	payload, err := json.Marshal(map[string]any{
		"s": randomStr,
		"t": actionParams.Tenant,
		"u": actionParams.User,
	})
	if err != nil {
		return err
	}

	apiKey := base64.URLEncoding.EncodeToString(payload)

	fmt.Println(apiKey)

	return nil
}

func (r *Run) getRandomString(length int) (string, error) {
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
	getEdgeApiKeyCmd.PersistentFlags().StringVarP(&actionParams.Tenant, "tenant", "t", "", "Tenant (required)")
	getEdgeApiKeyCmd.PersistentFlags().StringVarP(&actionParams.User, "user", "U", "", "User (required)")
	getEdgeApiKeyCmd.PersistentFlags().IntVarP(&actionParams.Length, "length", "l", 17, "Salt length")
	if err := getEdgeApiKeyCmd.MarkPersistentFlagRequired("tenant"); err != nil {
		slog.Error("failed to mark tenant flag as required", "error", err)
		os.Exit(1)
	}
	if err := getEdgeApiKeyCmd.MarkPersistentFlagRequired("user"); err != nil {
		slog.Error("failed to mark user flag as required", "error", err)
		os.Exit(1)
	}
}
