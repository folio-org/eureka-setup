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

	"github.com/folio-org/eureka-cli/action"
	"github.com/spf13/cobra"
)

var (
	withKongGateway      string
	withTenantIDs        []string
	withApplicationNames []string
)

// purgeTenantsCmd represents the purgeTenants command
var purgeTenantsCmd = &cobra.Command{
	Use:   "purgeTenants",
	Short: "Purge tenants",
	Long:  `Purge tenants and their entitlements.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		run, err := New(action.PurgeTenants)
		if err != nil {
			return err
		}

		slog.Info(run.RunConfig.Action.Name, slog.String("text", "PURGING TENANTS"), slog.String("Using Kong Gateway", withKongGateway))

		slog.Info(run.RunConfig.Action.Name, "text", "Purging tenant entitlements")
		for _, tenantID := range withTenantIDs {
			for key, value := range map[string][]string{tenantID: withApplicationNames} {
				requestURL, err := url.JoinPath(withKongGateway, "/entitlements")
				if err != nil {
					return err
				}

				payload, err := json.Marshal(map[string]any{
					"tenantId":     key,
					"applications": value,
				})
				if err != nil {
					return err
				}

				_ = run.RunConfig.HTTPClient.DeleteWithBody(fmt.Sprintf("%s%s", requestURL, "?purge=true"), payload, map[string]string{})

				slog.Info(run.RunConfig.Action.Name, "text", "Purged tenant entitlement with applications", "tenant", key, "applications", value)
			}
		}

		slog.Info(run.RunConfig.Action.Name, "text", "Purging tenants")
		for _, tenantID := range withTenantIDs {
			requestURL, err := url.JoinPath(withKongGateway, "/tenants", tenantID)
			if err != nil {
				return err
			}

			_ = run.RunConfig.HTTPClient.Delete(fmt.Sprintf("%s%s", requestURL, "?purgeKafkaTopics=true"), map[string]string{})

			slog.Info(run.RunConfig.Action.Name, "text", "Purged tenant", "tenant", tenantID)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(purgeTenantsCmd)
	purgeTenantsCmd.PersistentFlags().StringVarP(&withKongGateway, "gateway", "", "http://localhost:8000", "Kong Gateway")
	purgeTenantsCmd.PersistentFlags().StringSliceVarP(&withTenantIDs, "ids", "", []string{}, "Tenant ids")
	purgeTenantsCmd.PersistentFlags().StringSliceVarP(&withApplicationNames, "apps", "", []string{"app-combined-1.0.0"}, "Application names")
}
