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
	"time"

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/errors"
	"github.com/folio-org/eureka-cli/field"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/folio-org/eureka-cli/models"
	"github.com/spf13/cobra"
)

// deployModulesCmd represents the deployModules command
var deployModulesCmd = &cobra.Command{
	Use:   "deployModules",
	Short: "Deploy modules",
	Long:  `Deploy multiple module versions.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		run, err := New(action.DeployModules)
		if err != nil {
			return err
		}

		return run.DeployModules()
	},
}

func (run *Run) DeployModules() error {
	slog.Info(run.Config.Action.Name, "text", "READING BACKEND MODULES FROM CONFIG")
	backendModules, err := run.Config.ModuleParams.ReadBackendModulesFromConfig(false, true)
	if err != nil {
		return err
	}

	slog.Info(run.Config.Action.Name, "text", "READING FRONTEND MODULES FROM CONFIG")
	frontendModules, err := run.Config.ModuleParams.ReadFrontendModulesFromConfig(true)
	if err != nil {
		return err
	}

	slog.Info(run.Config.Action.Name, "text", "READING BACKEND MODULE REGISTRIES")
	instalJsonURLs := run.Config.Action.GetCombinedInstallJsonURLs()
	modules, err := run.Config.RegistrySvc.GetModules(instalJsonURLs, true)
	if err != nil {
		return err
	}
	run.Config.RegistrySvc.ExtractModuleMetadata(modules)

	client, err := run.Config.DockerClient.Create()
	if err != nil {
		return err
	}
	defer run.Config.DockerClient.Close(client)
	if err := run.setVaultRootTokenIntoContext(client); err != nil {
		return err
	}

	slog.Info(run.Config.Action.Name, "text", "PULLING SIDECAR IMAGE")
	globalEnv := run.Config.Action.GetConfigEnvVars(field.Env)
	sidecarEnv := run.Config.Action.GetConfigEnvVars(field.SidecarModuleEnv)
	containers := models.NewCoreAndBusinessContainers(run.Config.Action.VaultRootToken, modules, backendModules, globalEnv, sidecarEnv)
	sidecarImage, pullSidecarImage, err := run.Config.ModuleSvc.GetSidecarImage(containers.Modules.EurekaModules)
	if err != nil {
		return err
	}

	slog.Info(run.Config.Action.Name, "text", "Using sidecar image", "image", sidecarImage)
	if pullSidecarImage {
		err = run.Config.ModuleSvc.PullModule(client, sidecarImage)
		if err != nil {
			return err
		}
	}

	slog.Info(run.Config.Action.Name, "text", "DEPLOYING MODULES")
	sidecarResources := helpers.CreateResources(false, run.Config.Action.ConfigSidecarResources)
	deployedModules, err := run.Config.ModuleSvc.DeployModules(client, containers, sidecarImage, sidecarResources)
	if err != nil {
		return err
	}
	if len(deployedModules) == 0 {
		return errors.ModulesNotDeployed(len(deployedModules))
	}
	time.Sleep(constant.DeployModulesWait)

	slog.Info(run.Config.Action.Name, "text", "WAITING FOR MODULES TO BECOME READY")
	var deployModulesWG sync.WaitGroup
	errCh := make(chan error, len(deployedModules))
	deployModulesWG.Add(len(deployedModules))
	for deployedModule := range deployedModules {
		go run.Config.ModuleSvc.CheckModuleReadiness(&deployModulesWG, errCh, deployedModule, deployedModules[deployedModule])
	}
	deployModulesWG.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			return err
		}
	}
	slog.Info(run.Config.Action.Name, "text", "All modules are ready")

	slog.Info(run.Config.Action.Name, "text", "CREATING APPLICATION")
	registryURLs := run.Config.Action.GetCombinedRegistryURLs()
	registerExtract := models.NewRegistryExtract(registryURLs, modules, backendModules, frontendModules)
	return run.Config.ManagementSvc.CreateApplications(registerExtract)
}

func init() {
	rootCmd.AddCommand(deployModulesCmd)
}
