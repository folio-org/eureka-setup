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

	"github.com/j011195/eureka-setup/eureka-cli/action"
	"github.com/j011195/eureka-setup/eureka-cli/constant"
	"github.com/j011195/eureka-setup/eureka-cli/errors"
	"github.com/j011195/eureka-setup/eureka-cli/helpers"
	"github.com/j011195/eureka-setup/eureka-cli/models"
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
	slog.Info(run.Config.Action.Name, "text", "READING BACKEND MODULES")
	backendModules, err := run.Config.ModuleProps.ReadBackendModules(false, true)
	if err != nil {
		return err
	}

	slog.Info(run.Config.Action.Name, "text", "READING FRONTEND MODULES")
	frontendModules, err := run.Config.ModuleProps.ReadFrontendModules(true)
	if err != nil {
		return err
	}

	slog.Info(run.Config.Action.Name, "text", "READING BACKEND MODULE REGISTRIES")
	installJsonURLs := run.Config.Action.GetCombinedInstallJsonURLs()
	modules, err := run.Config.RegistrySvc.GetModules(installJsonURLs, true, true)
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
	containers := &models.Containers{
		Modules:        modules,
		BackendModules: backendModules,
		IsManagement:   false,
	}
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
	sidecarResources := helpers.CreateResources(false, run.Config.Action.ConfigSidecarModuleResources)
	deployedModules, err := run.Config.ModuleSvc.DeployModules(client, containers, sidecarImage, sidecarResources)
	if err != nil {
		return err
	}
	if len(deployedModules) == 0 {
		return errors.ModulesNotDeployed(len(deployedModules))
	}
	time.Sleep(constant.DeployModulesWait)

	slog.Info(run.Config.Action.Name, "text", "WAITING FOR MODULES TO BECOME READY")
	if err := run.CheckDeployedModuleReadiness(constant.Module, deployedModules); err != nil {
		return err
	}

	slog.Info(run.Config.Action.Name, "text", "CREATING APPLICATION")
	if err := run.setKeycloakMasterAccessTokenIntoContext(constant.ClientCredentials); err != nil {
		return err
	}

	return run.Config.ManagementSvc.CreateApplication(&models.RegistryExtract{
		Modules:           modules,
		BackendModules:    backendModules,
		FrontendModules:   frontendModules,
		ModuleDescriptors: make(map[string]any),
	})
}

func init() {
	rootCmd.AddCommand(deployModulesCmd)
	deployModulesCmd.PersistentFlags().BoolVarP(&params.SkipRegistry, action.SkipRegistry.Long, action.SkipRegistry.Short, false, action.SkipRegistry.Description)
}
