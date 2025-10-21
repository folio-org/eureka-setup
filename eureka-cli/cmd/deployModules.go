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
	"github.com/spf13/viper"
)

// deployModulesCmd represents the deployModules command
var deployModulesCmd = &cobra.Command{
	Use:   "deployModules",
	Short: "Deploy modules",
	Long:  `Deploy multiple module versions.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		startPort := viper.GetInt(field.ApplicationPortStart)
		endPort := viper.GetInt(field.ApplicationPortEnd)
		r, err := NewCustom(action.DeployModules, startPort, endPort)
		if err != nil {
			return err
		}

		err = r.DeployModules()
		if err != nil {
			return err
		}

		return nil
	},
}

func (r *Run) DeployModules() error {
	registryURL := viper.GetString(field.RegistryURL)
	env := helpers.GetConfigEnvVars(field.Env)
	sidecarEnv := helpers.GetConfigEnvVars(field.SidecarModuleEnv)

	slog.Info(r.Config.Action.Name, "text", "READING BACKEND MODULES FROM CONFIG")
	backendModulesMap, err := r.Config.ModuleParams.GetBackendModulesFromConfig(false, true, viper.GetStringMap(field.BackendModules))
	if err != nil {
		return err
	}

	slog.Info(r.Config.Action.Name, "text", "READING FRONTEND MODULES FROM CONFIG")
	frontendModulesMap := r.Config.ModuleParams.GetFrontendModulesFromConfig(true, viper.GetStringMap(field.FrontendModules), viper.GetStringMap(field.CustomFrontendModules))

	slog.Info(r.Config.Action.Name, "text", "READING BACKEND MODULE REGISTRIES")
	instalJsonURLs := map[string]string{
		constant.FolioRegistry:  viper.GetString(field.InstallFolio),
		constant.EurekaRegistry: viper.GetString(field.InstallEureka),
	}
	registryModules, err := r.Config.RegistrySvc.GetModules(instalJsonURLs, true)
	if err != nil {
		return err
	}

	slog.Info(r.Config.Action.Name, "text", "EXTRACTING MODULE NAME AND VERSION")
	r.Config.RegistrySvc.ExtractModuleNameAndVersion(registryModules, true)

	vaultRootToken, client, err := r.GetVaultRootTokenWithDockerClient()
	if err != nil {
		return err
	}
	defer r.Config.DockerClient.Close(client)

	slog.Info(r.Config.Action.Name, "text", "CREATING APPLICATIONS")
	registryURLs := map[string]string{constant.FolioRegistry: registryURL, constant.EurekaRegistry: registryURL}
	registerModuleExtract := models.NewRegistryModuleExtract(registryURLs, registryModules, backendModulesMap, frontendModulesMap)
	err = r.Config.ManagementSvc.CreateApplications(registerModuleExtract)
	if err != nil {
		return err
	}

	slog.Info(r.Config.Action.Name, "text", "PULLING SIDECAR IMAGE")
	registryHosts := map[string]string{constant.FolioRegistry: "", constant.EurekaRegistry: ""}
	containers := models.NewCoreAndBusinessContainers(vaultRootToken, registryHosts, registryModules, backendModulesMap, env, sidecarEnv)
	sidecarImage, pullSidecarImage, err := r.Config.ModuleSvc.GetSidecarImage(containers.RegistryModules[constant.EurekaRegistry])
	if err != nil {
		return err
	}

	slog.Info(r.Config.Action.Name, "text", "Using sidecar image", "image", sidecarImage)
	sidecarResources := helpers.CreateResources(false, viper.GetStringMap(field.SidecarModuleResources))
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
	var wg sync.WaitGroup
	errCh := make(chan error, len(deployedModules))

	wg.Add(len(deployedModules))
	for deployedModule := range deployedModules {
		go r.Config.ModuleSvc.PerformModuleReadinessCheck(&wg, errCh, deployedModule, deployedModules[deployedModule])
	}
	wg.Wait()
	close(errCh)

	select {
	case err := <-errCh:
		return err
	default:
	}

	slog.Info(r.Config.Action.Name, "text", "All modules are ready")

	return nil
}

func init() {
	rootCmd.AddCommand(deployModulesCmd)
}
