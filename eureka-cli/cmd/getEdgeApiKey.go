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

	"github.com/folio-org/eureka-cli/internal"
	"github.com/spf13/cobra"
)

const (
	getEdgeApiKeyCommand string = "Get Edge Api Key"

	charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

// getEdgeApiKeyCmd represents the generateApiKey command
var getEdgeApiKeyCmd = &cobra.Command{
	Use:   "getEdgeApiKey",
	Short: "Get Edge API key",
	Long:  `Get Edge API key for a tenant`,
	Run: func(cmd *cobra.Command, args []string) {
		GetEdgeApiKey()
	},
}

func GetEdgeApiKey() {
	apiKeyBytes, err := json.Marshal(map[string]any{"s": getRandomString(withLength), "t": withTenant, "u": withUser})
	if err != nil {
		slog.Error(getEdgeApiKeyCommand, internal.GetFuncName(), "json.Marshal error")
		panic(err)
	}

	apiKey := base64.URLEncoding.EncodeToString(apiKeyBytes)

	fmt.Println(apiKey)
}

func getRandomString(length int) string {
	bytes := make([]byte, length)
	for i := range length {
		charsetIdx, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			slog.Error(getEdgeApiKeyCommand, internal.GetFuncName(), "rand.Int error")
			panic(err)
		}

		bytes[i] = charset[charsetIdx.Int64()]
	}

	return string(bytes)
}

func init() {
	rootCmd.AddCommand(getEdgeApiKeyCmd)
	getEdgeApiKeyCmd.PersistentFlags().StringVarP(&withTenant, "tenant", "t", "", "Tenant (required)")
	getEdgeApiKeyCmd.PersistentFlags().StringVarP(&withUser, "user", "U", "", "User (required)")
	getEdgeApiKeyCmd.PersistentFlags().IntVarP(&withLength, "length", "l", 17, "Salt length")
	getEdgeApiKeyCmd.MarkPersistentFlagRequired("tenant")
	getEdgeApiKeyCmd.MarkPersistentFlagRequired("user")
}
