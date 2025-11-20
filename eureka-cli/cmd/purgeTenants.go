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
	"github.com/folio-org/eureka-cli/models"
	"github.com/spf13/cobra"
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
		slog.Info(run.Config.Action.Name, slog.String("text", "PURGING TENANTS"), slog.String("Using Kong Gateway", run.Config.Action.Param.GatewayURL))

		slog.Info(run.Config.Action.Name, "text", "Purging tenant entitlements")
		for _, tenantID := range run.Config.Action.Param.TenantIDs {
			for key, value := range map[string][]string{tenantID: run.Config.Action.Param.ApplicationNames} {
				requestURL, err := url.JoinPath(run.Config.Action.Param.GatewayURL, "/entitlements")
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

				var decodedResponse models.TenantEntitlementResponse
				err = run.Config.HTTPClient.DeleteWithPayloadReturnStruct(fmt.Sprintf("%s%s", requestURL, "?purge=true"), payload, map[string]string{}, &decodedResponse)
				if err != nil {
					slog.Warn(run.Config.Action.Name, "text", "Purge of tenant entitlements was unsuccessful", "tenant", key, "error", err)
				} else {
					slog.Info(run.Config.Action.Name, "text", "Purged tenant entitlements", "tenant", key, "applications", value, "flowId", decodedResponse.FlowID)
				}
			}
		}

		slog.Info(run.Config.Action.Name, "text", "Purging tenants")
		for _, tenantID := range run.Config.Action.Param.TenantIDs {
			requestURL, err := url.JoinPath(run.Config.Action.Param.GatewayURL, "/tenants", tenantID)
			if err != nil {
				return err
			}

			err = run.Config.HTTPClient.Delete(fmt.Sprintf("%s%s", requestURL, "?purgeKafkaTopics=true"), map[string]string{})
			if err != nil {
				slog.Warn(run.Config.Action.Name, "text", "Purge of tenants was unsuccessful", "tenant", tenantID, "error", err)
			}

			slog.Info(run.Config.Action.Name, "text", "Purged tenants", "tenant", tenantID)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(purgeTenantsCmd)
	purgeTenantsCmd.PersistentFlags().StringVarP(&params.GatewayURL, action.GatewayURL.Long, action.GatewayURL.Short, "http://localhost:8000", action.GatewayURL.Description)
	purgeTenantsCmd.PersistentFlags().StringSliceVarP(&params.TenantIDs, action.TenantIDs.Long, action.TenantIDs.Short, []string{}, action.TenantIDs.Description)
	purgeTenantsCmd.PersistentFlags().StringSliceVarP(&params.ApplicationNames, action.ApplicationNames.Long, action.ApplicationNames.Short, []string{"app-combined-1.0.0"}, action.ApplicationNames.Description)
}
