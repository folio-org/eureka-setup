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
	"sync"

	"github.com/j011195/eureka-setup/eureka-cli/action"
	"github.com/j011195/eureka-setup/eureka-cli/constant"
	"github.com/j011195/eureka-setup/eureka-cli/field"
	"github.com/j011195/eureka-setup/eureka-cli/helpers"
	"github.com/j011195/eureka-setup/eureka-cli/runconfig"
)

// Run is a container that holds the RunConfig instance
type Run struct {
	Config *runconfig.RunConfig
}

func New(name string) (*Run, error) {
	gatewayURLTemplate, err := action.GetGatewayURLTemplate(name)
	if err != nil {
		return nil, err
	}
	action := action.New(name, gatewayURLTemplate, &params)

	runConfig, err := runconfig.New(action, logger)
	if err != nil {
		return nil, err
	}

	return &Run{Config: runConfig}, nil
}

func (run *Run) PingKongStatus() error {
	requestURL := run.Config.Action.GetRequestURL(constant.KongAdminPort, "/status")
	return run.Config.HTTPClient.PingRetry(requestURL)
}

func (run *Run) ConsortiumPartition(fn func(string, constant.TenantType) error) error {
	if !action.IsSet(field.Consortiums) {
		return fn(constant.NoneConsortium, constant.Default)
	}
	for consortiumName := range run.Config.Action.ConfigConsortiums {
		for _, tenantType := range constant.GetTenantTypes() {
			if err := fn(consortiumName, tenantType); err != nil {
				return err
			}
		}
	}

	return nil
}

func (run *Run) TenantPartition(consortiumName string, tenantType constant.TenantType, fn func(string, string) error) error {
	client, err := run.Config.DockerClient.Create()
	if err != nil {
		return err
	}
	if err := run.setVaultRootTokenIntoContext(client); err != nil {
		return err
	}
	if err := run.setKeycloakMasterAccessTokenIntoContext(constant.ClientCredentials); err != nil {
		return err
	}

	tenants, err := run.Config.ManagementSvc.GetTenants(consortiumName, tenantType)
	if err != nil {
		return err
	}

	for _, value := range tenants {
		entry := value.(map[string]any)
		configTenant := helpers.GetString(entry, "name")
		if !helpers.HasTenant(configTenant, run.Config.Action.ConfigTenants) {
			continue
		}
		if err := run.setKeycloakAccessTokenIntoContext(configTenant); err != nil {
			return err
		}
		configDescription := helpers.GetString(entry, "description")
		if err = fn(configTenant, configDescription); err != nil {
			return err
		}
	}

	return nil
}

func (run *Run) CheckDeployedModuleReadiness(moduleType string, modules map[string]int) error {
	var (
		wg    sync.WaitGroup
		errCh = make(chan error, len(modules))
	)

	wg.Add(len(modules))
	for deployedModule := range modules {
		go run.Config.ModuleSvc.CheckModuleReadiness(&wg, errCh, deployedModule, modules[deployedModule])
	}
	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			return err
		}
	}
	slog.Info(run.Config.Action.Name, "text", "All modules are ready", "type", moduleType)

	return nil
}
