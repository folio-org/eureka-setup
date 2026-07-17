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

	prepopulatedDescriptors := make(map[string]any)
	baseRegistryURL := strings.TrimSpace(run.Config.Action.ConfigRegistryURL)

	// --- UNIVERSAL PROFILE MODULE INJECTION PATCH START ---
	for name, bMod := range backendModules {
		found := false
		for _, fm := range modules.FolioModules {
			if fm.Metadata.Name == name || isStrictModuleID(fm.ID, name) {
				found = true
				break
			}
		}
		if !found {
			for _, em := range modules.EurekaModules {
				if em.Metadata.Name == name || isStrictModuleID(em.ID, name) {
					found = true
					break
				}
			}
		}

		if !found {
			var pristineVersion string
			if bMod.ModuleVersion != nil && *bMod.ModuleVersion != "" {
				pristineVersion = *bMod.ModuleVersion
				slog.Debug(run.Config.Action.Name, "text", "Module injection engine: picked explicit configuration version", "module", name, "version", pristineVersion)
			} else {
				// Dynamic Auto-Discovery returns the pristine version containing the '+' metadata delimiter
				discoveredVersion, lookupErr := run.discoverLatestRegistryVersion(name)
				if lookupErr == nil && discoveredVersion != "" {
					pristineVersion = discoveredVersion
					slog.Debug(run.Config.Action.Name, "text", "Module injection engine: successfully auto-discovered target tag from private registry", "module", name, "version", pristineVersion)
				} else if lookupErr != nil {
					slog.Debug(run.Config.Action.Name, "text", "Module injection engine: registry discovery lookup error occurred", "module", name, "error", lookupErr.Error())
				}
			}

			if pristineVersion == "" {
				slog.Warn(run.Config.Action.Name, "text", "Module declared in profile backend-modules but not found in active deployment tree or registry discovery endpoint, skipping injection", "module", name)
				continue
			}

			// Sanitize the metadata string boundary for gateway/routing engine compatibility (+ converted to -)
			gatewayVersion := strings.ReplaceAll(pristineVersion, "+", "-")
			syntheticID := fmt.Sprintf("%s-%s", name, gatewayVersion)

			// Fetch the pristine descriptor file directly over the wire from Okapi proxy registries map
			if baseRegistryURL != "" {
				descURL := fmt.Sprintf("%s/_/proxy/modules/%s-%s", strings.TrimSuffix(baseRegistryURL, "/"), name, pristineVersion)
				ctxClient := &http.Client{Timeout: 5 * time.Second}
				descResp, descErr := ctxClient.Get(descURL)
				if descErr == nil && descResp.StatusCode == http.StatusOK {
					var descriptorData map[string]any
					if err := json.NewDecoder(descResp.Body).Decode(&descriptorData); err == nil {
						// Mutate internal description ID reference to match gateway sanitation tags
						descriptorData["id"] = syntheticID
						prepopulatedDescriptors[syntheticID] = descriptorData
						slog.Debug(run.Config.Action.Name, "text", "Descriptor Pre-population Engine cached record successfully", "module", name, "id", syntheticID)
					}
					descResp.Body.Close()
				}
			}

			syntheticProxy := &models.ProxyModule{
				ID:     syntheticID,
				Action: "enable",
				Metadata: models.ProxyModuleMetadata{
					Name:        name,
					SidecarName: name + "-sc",
					Version:     &gatewayVersion,
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
		ModuleDescriptors: prepopulatedDescriptors, // Swap cache injection payload over
	})
}

func isStrictModuleID(id string, name string) bool {
	prefix := name + "-"
	if !strings.HasPrefix(id, prefix) {
		return false
	}
	remainder := strings.TrimPrefix(id, prefix)
	if len(remainder) == 0 {
		return false
	}
	return remainder[0] >= '0' && remainder[0] <= '9'
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

	repoPath := moduleName
	if rawModuleConfig, exists := run.Config.Action.ConfigBackendModules[moduleName]; exists {
		if moduleMap, ok := rawModuleConfig.(map[string]any); ok {
			if img, imgExists := moduleMap["image"]; imgExists {
				if imgStr, isStr := img.(string); isStr && strings.TrimSpace(imgStr) != "" {
					repoPath = strings.TrimSpace(imgStr)
				}
			}
		}
	}

	if idx := strings.Index(repoPath, "/"); idx != -1 {
		firstPart := repoPath[:idx]
		if strings.Contains(firstPart, ".") || strings.Contains(firstPart, ":") {
			repoPath = repoPath[idx+1:]
		}
	}

	var paths []string
	if strings.Contains(repoPath, "/") {
		paths = []string{repoPath}
	} else {
		paths = []string{repoPath}
		for _, ns := range constant.GetNamespaces() {
			paths = append(paths, ns+"/"+repoPath)
		}
	}

	registries := run.Config.Action.ConfigDockerRegistries
	if len(registries) == 0 {
		registries = []string{"docker.io"}
	}

	for i := len(items) - 1; i >= 0; i-- {
		item := items[i]
		if isStrictModuleID(item.ID, moduleName) {
			rawVersion := strings.TrimPrefix(item.ID, moduleName+"-")
			cleanTag := strings.Split(rawVersion, "+")[0]

			for _, path := range paths {
				for _, registry := range registries {
					var privateRegistryURL string
					if registry == "docker.io" || registry == "registry-1.docker.io" {
						if !strings.Contains(path, "/") {
							continue
						}
						privateRegistryURL = fmt.Sprintf("https://registry-1.docker.io/v2/%s/manifests/%s", path, cleanTag)
					} else {
						privateRegistryURL = fmt.Sprintf("https://%s/v2/%s/manifests/%s", registry, path, cleanTag)
					}

					req, err := http.NewRequest(http.MethodHead, privateRegistryURL, nil)
					if err != nil {
						continue
					}
					req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")

					hubResp, hubErr := ctxClient.Do(req)
					if hubErr == nil {
						statusCode := hubResp.StatusCode
						hubResp.Body.Close()

						if statusCode == http.StatusOK {
							slog.Debug(run.Config.Action.Name, "text", "Private Registry Carousel: tag verified live via manifest probe", "module", moduleName, "registry", registry, "path", path, "tag", cleanTag)
							return rawVersion, nil // Return full raw version containing '+' for accurate remote lookups
						}
					}
				}
			}
			slog.Debug(run.Config.Action.Name, "text", "Private Registry Carousel: tag missing across layout variants (fallback pass)", "module", moduleName, "tag", cleanTag)
		}
	}

	return "", fmt.Errorf("no live verified container tags found inside the registry carousel loop for %s", moduleName)
}

func init() {
	rootCmd.AddCommand(deployModulesCmd)
	deployModulesCmd.PersistentFlags().BoolVarP(&params.SkipRegistry, action.SkipRegistry.Long, action.SkipRegistry.Short, false, action.SkipRegistry.Description)
}