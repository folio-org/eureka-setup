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
	"log/slog"
	"time"

	"github.com/folio-org/eureka-setup/eureka-cli/action"
	"github.com/folio-org/eureka-setup/eureka-cli/constant"
	"github.com/folio-org/eureka-setup/eureka-cli/helpers"
	"github.com/spf13/cobra"
)

// deployApplicationCmd represents the deployApplication command
var deployApplicationCmd = &cobra.Command{
	Use:     "deployApplication",
	Short:   "Deploy application",
	Long:    `Deploy platform application.`,
	Aliases: []string{"deploy"},
	RunE: func(cmd *cobra.Command, args []string) error {
		start := time.Now()

		var err error
		run, err := New(action.DeployApplication)
		if err != nil {
			return err
		}

		if params.Cleanup {
			err = run.DeployApplicationWithCleanup()
		} else {
			err = run.DeployApplicationWithoutCleanup()
		}
		if err != nil {
			return err
		}
		slog.Info(run.Config.Action.Name, "text", "Command completed", "duration", time.Since(start))

		return nil
	},
}

func (run *Run) DeployApplicationWithCleanup() error {
	if run.Config.Action.IsChildApp() {
		if err := run.UndeployChildApplication(); err != nil {
			return err
		}
		return run.DeployChildApplication()
	}
	if err := run.UndeployApplication(); err != nil {
		return err
	}

	return run.DeployApplication()
}

func (run *Run) DeployApplicationWithoutCleanup() error {
	if run.Config.Action.IsChildApp() {
		return run.DeployChildApplication()
	}

	return run.DeployApplication()
}

func (run *Run) DeployApplication() error {
	if err := run.DeploySystem(); err != nil {
		return err
	}
	if err := run.PingKongStatus(); err != nil {
		return err
	}
	if err := run.DeployManagement(); err != nil {
		return err
	}
	if err := run.DeployModules(); err != nil {
		return err
	}
	if err := run.CreateTenants(); err != nil {
		return err
	}
	if err := run.ConsortiumPartition(func(consortiumName string, tenantType constant.TenantType) error {
		if err := run.CreateTenantEntitlements(consortiumName, tenantType); err != nil {
			return err
		}
		if err := run.CreateRoles(consortiumName, tenantType); err != nil {
			return err
		}
		if err := run.CreateUsers(consortiumName, tenantType); err != nil {
			return err
		}
		if err := run.AttachCapabilitySets(consortiumName, tenantType, constant.DeployApplicationPartitionWait); err != nil {
			return err
		}
		if consortiumName != constant.NoneConsortium {
			time.Sleep(constant.DeployApplicationPartitionWait)
		}

		return nil
	}); err != nil {
		return err
	}
	if err := run.CreateConsortium(); err != nil {
		return err
	}
	return run.ConsortiumPartition(func(consortiumName string, tenantType constant.TenantType) error {
		if err := run.DeployUi(consortiumName, tenantType); err != nil {
			return err
		}
		if err := run.UpdateKeycloakPublicClients(consortiumName, tenantType); err != nil {
			return err
		}
		if helpers.IsModuleEnabled(constant.ModSearchModule, run.Config.Action.ConfigBackendModules) {
			return run.ReindexIndices(consortiumName, tenantType)
		}

		return nil
	})
}

func (run *Run) DeployChildApplication() error {
	if err := run.DeployAdditionalSystem(); err != nil {
		return err
	}
	if err := run.DeployModules(); err != nil {
		return err
	}
	return run.ConsortiumPartition(func(consortiumName string, tenantType constant.TenantType) error {
		if err := run.CreateTenantEntitlements(consortiumName, tenantType); err != nil {
			return err
		}
		if err := run.DetachCapabilitySets(consortiumName, tenantType); err != nil {
			return err
		}

		return run.AttachCapabilitySets(consortiumName, tenantType, 0*time.Second)
	})
}

func init() {
	rootCmd.AddCommand(deployApplicationCmd)
	deployApplicationCmd.PersistentFlags().BoolVarP(&params.BuildImages, action.BuildImages.Long, action.BuildImages.Short, false, action.BuildImages.Description)
	deployApplicationCmd.PersistentFlags().BoolVarP(&params.UpdateCloned, action.UpdateCloned.Long, action.UpdateCloned.Short, false, action.UpdateCloned.Description)
	deployApplicationCmd.PersistentFlags().BoolVarP(&params.OnlyRequired, action.OnlyRequired.Long, action.OnlyRequired.Short, false, action.OnlyRequired.Description)
	deployApplicationCmd.PersistentFlags().BoolVarP(&params.Cleanup, action.Cleanup.Long, action.Cleanup.Short, false, action.Cleanup.Description)
	deployApplicationCmd.PersistentFlags().BoolVarP(&params.SkipRegistry, action.SkipRegistry.Long, action.SkipRegistry.Short, false, action.SkipRegistry.Description)
}
