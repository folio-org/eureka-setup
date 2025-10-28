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

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/helpers"
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
		r, err := New(action.DeployApplication)
		if err != nil {
			return err
		}

		if len(r.RunConfig.Action.ConfigApplicationDependencies) > 0 {
			err = r.DeployChildApplication()
		} else {
			err = r.DeployApplication()
		}
		if err != nil {
			return err
		}
		helpers.LogCompletion(r.RunConfig.Action.Name, start)

		return nil
	},
}

func (r *Run) DeployApplication() error {
	err := r.DeploySystem()
	if err != nil {
		return err
	}
	err = r.PingGateway()
	if err != nil {
		return err
	}
	err = r.DeployManagement()
	if err != nil {
		return err
	}
	err = r.DeployModules()
	if err != nil {
		return err
	}
	err = r.CreateTenants()
	if err != nil {
		return err
	}
	err = r.ConsortiumPartitionErr(func(consortiumName string, tenantType constant.TenantType) error {
		err = r.CreateTenantEntitlements(consortiumName, tenantType)
		if err != nil {
			return err
		}
		err = r.CreateRoles(consortiumName, tenantType)
		if err != nil {
			return err
		}
		err = r.CreateUsers(consortiumName, tenantType)
		if err != nil {
			return err
		}
		err := r.AttachCapabilitySets(consortiumName, tenantType, constant.DeployApplicationPartitionWait)
		if err != nil {
			return err
		}
		if consortiumName != constant.NoneConsortium {
			slog.Info(r.RunConfig.Action.Name, "text", "Waiting for duration", "duration_seconds", constant.DeployApplicationPartitionWait.Seconds())
			time.Sleep(constant.DeployApplicationPartitionWait)
		}

		return nil
	})
	if err != nil {
		return err
	}
	err = r.CreateConsortium()
	if err != nil {
		return err
	}
	err = r.DeployUi()
	if err != nil {
		return err
	}
	err = r.UpdateKeycloakPublicClients()
	if err != nil {
		return err
	}
	if helpers.IsModuleEnabled(constant.ModSearchModule, r.RunConfig.Action.ConfigBackendModules) {
		err = r.ConsortiumPartitionErr(func(consortiumName string, tenantType constant.TenantType) error {
			return r.ReindexIndices(consortiumName, tenantType)
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Run) DeployChildApplication() error {
	err := r.DeployAdditionalSystem()
	if err != nil {
		return err
	}
	err = r.DeployModules()
	if err != nil {
		return err
	}
	err = r.ConsortiumPartitionErr(func(consortiumName string, tenantType constant.TenantType) error {
		err = r.CreateTenantEntitlements(consortiumName, tenantType)
		if err != nil {
			return err
		}
		err = r.DetachCapabilitySets(consortiumName, tenantType)
		if err != nil {
			return err
		}
		err := r.AttachCapabilitySets(consortiumName, tenantType, 0*time.Second)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func init() {
	rootCmd.AddCommand(deployApplicationCmd)
	deployApplicationCmd.PersistentFlags().BoolVarP(&actionParams.BuildImages, "buildImages", "b", false, "Build Docker images")
	deployApplicationCmd.PersistentFlags().BoolVarP(&actionParams.UpdateCloned, "updateCloned", "u", false, "Update Git cloned projects")
	deployApplicationCmd.PersistentFlags().BoolVarP(&actionParams.OnlyRequired, "onlyRequired", "R", false, "Use only required system containers")
}
