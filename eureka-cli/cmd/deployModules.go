/*
Copyright © 2026 Open Library Foundation

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
	"net/http"
	"strings"
	"time"

	"github.com/folio-org/eureka-setup/eureka-cli/action"
	"github.com/folio-org/eureka-setup/eureka-cli/constant"
	"github.com/folio-org/eureka-setup/eureka-cli/errors"
	"github.com/folio-org/eureka-setup/eureka-cli/helpers"
	"github.com/folio-org/eureka-setup/eureka-cli/models"
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
	modules, err := run.Config.RegistrySvc.GetModules(true, true)
	if err != nil {
		return err
	}

	// Resolve native framework metadata first
	run.Config.RegistrySvc.ResolveModuleMetadata(modules)

	// --- UNIVERSAL PROFILE MODULE INJECTION PATCH START ---
	// Safely inject custom registry modules AFTER metadata resolution to bypass strict regex limitations
	for name, bMod := range backendModules {
		found := false
		// Match against names cleanly resolved by the framework
		for _, fm := range modules.FolioModules {
			if fm.Metadata.Name == name {
				found = true
				break
			}
		}
		if !found {
			for _, em := range modules.EurekaModules {
				if em.Metadata.Name == name {
					found = true
					break
				}
			}
		}

		if !found {
			version := "1.0.0"
			if bMod.ModuleVersion != nil && *bMod.ModuleVersion != "" {
				version = *bMod.ModuleVersion
			} else {
				discoveredVersion, lookupErr := run.discoverLatestRegistryVersion(name)
				if lookupErr == nil && discoveredVersion != "" {
					version = discoveredVersion
				}
			}

			syntheticID := fmt.Sprintf("%s-%s", name, version)
			syntheticProxy := &models.ProxyModule{
				ID:     syntheticID,
				Action: "enable",
				Metadata: models.ProxyModuleMetadata{
					Name:        name,
					SidecarName: name + "-sc",
					Version:     &version,
				},
			}
			modules.FolioModules = append(modules.FolioModules, syntheticProxy)
			slog.Info(run.Config.Action.Name, "text", "Injected custom proxy module into registry orchestration layer", "module", name, "id", syntheticID)
		}
	}
	// --- UNIVERSAL PROFILE MODULE INJECTION PATCH END ---

	client, err := run.Config.DockerClient.Create()
	if err != nil {
		return err
	}
	defer run.Config.DockerClient.Close(client)
	if err := run.setVaultRootTokenIntoContext(client); err != nil {
		return err
	}

	slog.Info(run.Config.Action.Name, "text", "PREPARING SIDECAR IMAGE")
	containers := &models.Containers{
		Modules:        modules,
		BackendModules: backendModules,
		IsManagement:   false,
	}
	sidecarImage, pullSidecarImage, err := run.Config.ModuleSvc.GetSidecarImage(containers.Modules.EurekaModules)
	if err != nil {
		return err
	}

	slog.Info(run.Config.Action.Name, "text", "Using sidecar image", "image", sidecarImage, "pullImage", pullSidecarImage)
	if pullSidecarImage {
		err = run.Config.ModuleSvc.PullModule(client, sidecarImage)
		if err != nil {
			return err
		}
	}

	slog.Info(run.Config.Action.Name, "text", "DEPLOYING MODULES")
	sidecarResources := helpers.CreateResources(false, run.Config.Action.ConfigSidecarModuleResources)
	newlyDeployed, totalMatched, err := run.Config.ModuleSvc.DeployModules(client, containers, sidecarImage, sidecarResources)
	if err != nil {
		return err
	}
	if totalMatched == 0 {
		return errors.ModulesNotDeployed(totalMatched)
	}
	if len(newlyDeployed) == 0 {
		slog.Info(run.Config.Action.Name, "text", "All modules already deployed, skipping healthchecks")
	} else {
		time.Sleep(constant.DeployModulesWait)

		slog.Info(run.Config.Action.Name, "text", "WAITING FOR MODULES TO BECOME READY")
		if err := run.CheckDeployedModuleReadiness(constant.Module, newlyDeployed); err != nil {
			return err
		}
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

func (run *Run) discoverLatestRegistryVersion(moduleName string) (string, error) {
	type RegistryItem struct {
		ID string `json:"id"`
	}

	baseRegistryURL := strings.TrimSpace(run.Config.Action.ConfigRegistryURL)
	if baseRegistryURL == "" {
		return "", fmt.Errorf("registry URL is unconfigured in profile")
	}

	endpoint := fmt.Sprintf("%s/_/proxy/modules", strings.TrimSuffix(baseRegistryURL, "/"))

	ctxClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := ctxClient.Get(endpoint)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var items []RegistryItem
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return "", err
	}

	prefix := moduleName + "-"
	latestVersion := ""

	for _, item := range items {
		if strings.HasPrefix(item.ID, prefix) {
			latestVersion = strings.TrimPrefix(item.ID, prefix)
		}
	}

	return latestVersion, nil
}

func init() {
	rootCmd.AddCommand(deployModulesCmd)
	deployModulesCmd.PersistentFlags().BoolVarP(&params.SkipRegistry, action.SkipRegistry.Long, action.SkipRegistry.Short, false, action.SkipRegistry.Description)
}