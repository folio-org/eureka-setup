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
	withTenantIds        []string
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

		slog.Info(run.Config.Action.Name, slog.String("text", "PURGING TENANTS"), slog.String("Using Kong Gateway", withKongGateway))

		slog.Info(run.Config.Action.Name, "text", "Purging tenant entitlements")
		for _, tenantId := range withTenantIds {
			for key, value := range map[string][]string{tenantId: withApplicationNames} {
				requestURL, err := url.JoinPath(withKongGateway, "/entitlements")
				if err != nil {
					return err
				}

				bytes, err := json.Marshal(map[string]any{"tenantId": key, "applications": value})
				if err != nil {
					return err
				}

				_ = run.Config.HTTPClient.DeleteWithBody(fmt.Sprintf("%s%s", requestURL, "?purge=true"), bytes, map[string]string{})

				slog.Info(run.Config.Action.Name, "text", fmt.Sprintf("Purged %s tenant entitlement with %s applications", key, value))
			}
		}

		slog.Info(run.Config.Action.Name, "text", "Purging tenants")
		for _, tenantId := range withTenantIds {
			requestURL, err := url.JoinPath(withKongGateway, "/tenants", tenantId)
			if err != nil {
				return err
			}

			_ = run.Config.HTTPClient.Delete(fmt.Sprintf("%s%s", requestURL, "?purgeKafkaTopics=true"), map[string]string{})

			slog.Info(run.Config.Action.Name, "text", fmt.Sprintf("Purged %s tenant", tenantId))
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(purgeTenantsCmd)
	purgeTenantsCmd.PersistentFlags().StringVarP(&withKongGateway, "gateway", "", "http://localhost:8000", "Kong Gateway")
	purgeTenantsCmd.PersistentFlags().StringSliceVarP(&withTenantIds, "ids", "", []string{}, "Tenant ids")
	purgeTenantsCmd.PersistentFlags().StringSliceVarP(&withApplicationNames, "apps", "", []string{"app-combined-1.0.0"}, "Application names")
}
