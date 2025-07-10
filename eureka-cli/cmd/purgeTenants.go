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
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"

	"github.com/folio-org/eureka-cli/internal"
	"github.com/spf13/cobra"
)

const commandName = "Purge Tenants"

var (
	withKongGateway      string
	withTenantIds        []string
	withApplicationNames []string
)

// purgeTenantsCmd represents the purgeTenantsCmd command
var purgeTenantsCmd = &cobra.Command{
	Use:   "purgeTenantsCmd",
	Short: "Purge tenants",
	Long:  `Purge tenants and their entitlements.`,
	Run: func(cmd *cobra.Command, args []string) {
		slog.Info(buildSystemCommand, internal.GetFuncName(), "### PURGING TENANTS ###")
		slog.Info(commandName, internal.GetFuncName(), fmt.Sprintf("Using Kong Gateway: %s", withKongGateway))

		slog.Info(commandName, internal.GetFuncName(), "Purging tenant entitlements")
		for _, tenantId := range withTenantIds {
			for key, value := range map[string][]string{tenantId: withApplicationNames} {
				requestUrl, err := url.JoinPath(withKongGateway, "/entitlements")
				if err != nil {
					slog.Error(commandName, internal.GetFuncName(), "json.Marshal error")
					panic(err)
				}

				entitlementBodyBytes, err := json.Marshal(map[string]any{"tenantId": key, "applications": value})
				if err != nil {
					slog.Error(commandName, internal.GetFuncName(), "json.Marshal error")
					panic(err)
				}

				internal.DoDeleteWithBody(commandName, fmt.Sprintf("%s%s", requestUrl, "?purge=true"), withEnableDebug, entitlementBodyBytes, true, map[string]string{})

				slog.Info(commandName, internal.GetFuncName(), fmt.Sprintf("Purged %s tenant entitlement with %s applications", key, value))
			}

		}

		slog.Info(commandName, internal.GetFuncName(), "Purging tenants")
		for _, tenantId := range withTenantIds {
			requestUrl, err := url.JoinPath(withKongGateway, "/tenants/", tenantId)
			if err != nil {
				slog.Error(commandName, internal.GetFuncName(), "json.Marshal error")
				panic(err)
			}

			internal.DoDelete(commandName, fmt.Sprintf("%s%s", requestUrl, "?purgeKafkaTopics=true"), withEnableDebug, false, map[string]string{})

			slog.Info(commandName, internal.GetFuncName(), fmt.Sprintf("Purged %s tenant", tenantId))
		}
	},
}

func init() {
	rootCmd.AddCommand(purgeTenantsCmd)
	purgeTenantsCmd.PersistentFlags().StringVarP(&withKongGateway, "gateway", "", "http://localhost:8000", "Kong Gateway")
	purgeTenantsCmd.PersistentFlags().StringSliceVarP(&withTenantIds, "ids", "", []string{}, "Tenant ids")
	purgeTenantsCmd.PersistentFlags().StringSliceVarP(&withApplicationNames, "apps", "", []string{"app-combined-1.0.0"}, "Application names")
}
