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
		r, err := New(action.DeployModules)
		if err != nil {
			return err
		}

		return r.DeployModules()
	},
}

func (r *Run) DeployModules() error {
	env := action.GetConfigEnvVars(field.Env)
	sidecarEnv := action.GetConfigEnvVars(field.SidecarModuleEnv)

	slog.Info(r.Config.Action.Name, "text", "READING BACKEND MODULES FROM CONFIG")
	backendModules, err := r.Config.ModuleParams.ReadBackendModulesFromConfig(false)
	if err != nil {
		return err
	}

	slog.Info(r.Config.Action.Name, "text", "READING FRONTEND MODULES FROM CONFIG")
	frontendModules := r.Config.ModuleParams.ReadFrontendModulesFromConfig()

	slog.Info(r.Config.Action.Name, "text", "READING BACKEND MODULE REGISTRIES")
	instalJsonURLs := map[string]string{
		constant.FolioRegistry:  r.Config.Action.ConfigFolioRegistry,
		constant.EurekaRegistry: r.Config.Action.ConfigEurekaRegistry,
	}
	registryModules, err := r.Config.RegistrySvc.GetModules(instalJsonURLs, true)
	if err != nil {
		return err
	}
	r.Config.RegistrySvc.ExtractModuleNameAndVersion(registryModules)

	client, err := r.Config.DockerClient.Create()
	if err != nil {
		return err
	}
	defer r.Config.DockerClient.Close(client)

	slog.Info(r.Config.Action.Name, "text", "CREATING APPLICATIONS")
	registryURLs := map[string]string{
		constant.FolioRegistry:  r.Config.Action.ConfigRegistryURL,
		constant.EurekaRegistry: r.Config.Action.ConfigRegistryURL,
	}
	registerModuleExtract := models.NewRegistryModuleExtract(registryURLs, registryModules, backendModules, frontendModules)
	err = r.Config.ManagementSvc.CreateApplications(registerModuleExtract)
	if err != nil {
		return err
	}

	slog.Info(r.Config.Action.Name, "text", "PULLING SIDECAR IMAGE")
	registryHosts := map[string]string{
		constant.FolioRegistry:  "",
		constant.EurekaRegistry: "",
	}
	containers := models.NewCoreAndBusinessContainers(r.Config.Action.VaultRootToken, registryHosts, registryModules, backendModules, env, sidecarEnv)
	sidecarImage, pullSidecarImage, err := r.Config.ModuleSvc.GetSidecarImage(containers.RegistryModules[constant.EurekaRegistry])
	if err != nil {
		return err
	}

	slog.Info(r.Config.Action.Name, "text", "Using sidecar image", "image", sidecarImage)
	sidecarResources := helpers.CreateResources(false, r.Config.Action.ConfigSidecarResources)
	if pullSidecarImage {
		err = r.Config.ModuleSvc.PullModule(client, sidecarImage)
		if err != nil {
			return err
		}
	}

	slog.Info(r.Config.Action.Name, "text", "DEPLOYING MODULES")
	deployedModules, err := r.Config.ModuleSvc.DeployModules(client, containers, sidecarImage, sidecarResources)
	if err != nil {
		return err
	}
	time.Sleep(constant.DeployModulesWait)

	slog.Info(r.Config.Action.Name, "text", "WAITING FOR MODULES TO BECOME READY")
	var deployModulesWG sync.WaitGroup
	errCh := make(chan error, len(deployedModules))

	deployModulesWG.Add(len(deployedModules))
	for deployedModule := range deployedModules {
		go r.Config.ModuleSvc.CheckModuleReadiness(&deployModulesWG, errCh, deployedModule, deployedModules[deployedModule])
	}
	deployModulesWG.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			return err
		}
	}

	slog.Info(r.Config.Action.Name, "text", "All modules are ready")

	return nil
}

func init() {
	rootCmd.AddCommand(deployModulesCmd)
}
