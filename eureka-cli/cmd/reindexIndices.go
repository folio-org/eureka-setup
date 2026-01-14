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

	"github.com/j011195/eureka-setup/eureka-cli/action"
	"github.com/j011195/eureka-setup/eureka-cli/constant"
	"github.com/j011195/eureka-setup/eureka-cli/field"
	"github.com/spf13/cobra"
)

// reindexIndicesCmd represents the reindexIndices command
var reindexIndicesCmd = &cobra.Command{
	Use:   "reindexIndices",
	Short: "Reindex elasticsearch",
	Long:  `Reindex elasticsearch indices.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		run, err := New(action.ReindexIndices)
		if err != nil {
			return err
		}

		return run.ConsortiumPartition(func(consortiumName string, tenantType constant.TenantType) error {
			return run.ReindexIndices(consortiumName, tenantType)
		})
	},
}

func (run *Run) ReindexIndices(consortiumName string, tenantType constant.TenantType) error {
	return run.TenantPartition(consortiumName, tenantType, func(configTenant, tenantType string) error {
		if action.IsSet(field.Consortiums) && tenantType == fmt.Sprintf("%s-%s", consortiumName, constant.Central) {
			slog.Info(run.Config.Action.Name, "text", "REINDEXING INDICES", "tenant", configTenant)
			if err := run.Config.SearchSvc.ReindexInventoryRecords(configTenant); err != nil {
				return err
			}
			if err := run.Config.SearchSvc.ReindexInstanceRecords(configTenant); err != nil {
				return err
			}
		}

		return nil
	})
}

func init() {
	rootCmd.AddCommand(reindexIndicesCmd)
}
