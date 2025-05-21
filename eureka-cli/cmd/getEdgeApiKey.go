/*
Copyright © 2025 EPAM_Systems/Thunderjet/Boburbek_Kadirkhodjaev

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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/folio-org/eureka-cli/internal"
	"github.com/spf13/cobra"
)

const (
	getEdgeApiKeyCommand string = "Get Edge Api Key"

	charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

// getEdgeApiKeyCmd represents the generateApiKey command
var getEdgeApiKeyCmd = &cobra.Command{
	Use:   "getEdgeApiKey",
	Short: "Get edge api key",
	Long:  `Get edge api key for a tenant`,
	Run: func(cmd *cobra.Command, args []string) {
		GetEdgeApiKey()
	},
}

func GetEdgeApiKey() {
	apiKeyBytes, err := json.Marshal(map[string]any{"s": getRandString(10, charset), "t": withTenant, "u": withUser})
	if err != nil {
		slog.Error(getEdgeApiKeyCommand, internal.GetFuncName(), "json.Marshal error")
		panic(err)
	}

	apiKey := base64.URLEncoding.EncodeToString(apiKeyBytes)

	fmt.Println(apiKey)
}

func getRandString(length int, charset string) string {
	randBytes := make([]byte, length)
	for i := range randBytes {
		randBytes[i] = charset[seededRand.Intn(len(charset))]
	}

	return string(randBytes)
}

func init() {
	rootCmd.AddCommand(getEdgeApiKeyCmd)
	getEdgeApiKeyCmd.PersistentFlags().StringVarP(&withTenant, "tenant", "t", "", "Tenant (required)")
	getEdgeApiKeyCmd.PersistentFlags().StringVarP(&withUser, "user", "U", "", "User (required)")
	getEdgeApiKeyCmd.MarkPersistentFlagRequired("tenant")
	getEdgeApiKeyCmd.MarkPersistentFlagRequired("user")
}
