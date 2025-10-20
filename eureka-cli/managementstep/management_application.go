package managementstep

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/field"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/folio-org/eureka-cli/models"
	"github.com/spf13/viper"
)

func (ms *ManagementStep) GetApplications() (models.Applications, error) {
	requestURL := fmt.Sprintf(ms.Action.GatewayURL, constant.KongPort, "/applications")

	response, err := ms.HTTPClient.DoGetReturnResponse(requestURL, map[string]string{})
	if err != nil {
		return models.Applications{}, err
	}

	if response == nil {
		return models.Applications{}, nil
	}
	defer func() {
		_ = response.Body.Close()
	}()

	var applications models.Applications

	err = json.NewDecoder(response.Body).Decode(&applications)
	if err != nil {
		return models.Applications{}, err
	}

	return applications, nil
}

func (ms *ManagementStep) CreateApplications(extract *models.RegistryModuleExtract) error {
	var (
		backendModules            []map[string]string
		frontendModules           []map[string]string
		backendModuleDescriptors  []any
		frontendModuleDescriptors []any
		dependencies              map[string]any
		discoveryModules          []map[string]string
	)

	applicationMap := viper.GetStringMap(field.Application)
	applicationName := applicationMap["name"].(string)
	applicationVersion := applicationMap["version"].(string)
	applicationPlatform := applicationMap["platform"].(string)
	applicationFetchDescriptors := applicationMap["fetch-descriptors"].(bool)

	applicationId := fmt.Sprintf("%s-%s", applicationName, applicationVersion)

	if applicationMap["dependencies"] != nil {
		dependencies = applicationMap["dependencies"].(map[string]any)
	}

	for registryName, registryModules := range extract.RegistryModules {
		if len(registryModules) > 0 {
			slog.Info(ms.Action.Name, "text", fmt.Sprintf("Adding %s modules to %s application", registryName, applicationId))
		}

		for _, module := range registryModules {
			if strings.Contains(module.Name, constant.ManagementModulePattern) {
				continue
			}

			backendModule, okBackend := extract.BackendModulesMap[module.Name]
			frontendModule, okFrontend := extract.FrontendModulesMap[module.Name]
			if (!okBackend && !okFrontend) || (okBackend && !backendModule.DeployModule || okFrontend && !frontendModule.DeployModule) {
				continue
			}

			if okBackend && backendModule.ModuleVersion != nil || okFrontend && frontendModule.ModuleVersion != nil {
				if backendModule.ModuleVersion != nil {
					module.Version = backendModule.ModuleVersion
				} else if frontendModule.ModuleVersion != nil {
					module.Version = frontendModule.ModuleVersion
				}
				module.Id = fmt.Sprintf("%s-%s", module.Name, *module.Version)
			}

			moduleDescriptorUrl := fmt.Sprintf("%s/_/proxy/modules/%s", extract.RegistryURLs[constant.FolioRegistry], module.Id)

			isLocalModule := okBackend && backendModule.LocalDescriptorPath != ""

			if applicationFetchDescriptors || isLocalModule {
				if isLocalModule {
					slog.Info(ms.Action.Name, "text", fmt.Sprintf("Fetching local module descriptor for %s from file path", module.Id))

					var descriptor map[string]any
					helpers.ReadJsonFromFile(ms.Action, backendModule.LocalDescriptorPath, &descriptor)
					extract.ModuleDescriptorsMap[module.Id] = descriptor

					slog.Info(ms.Action.Name, "text", fmt.Sprintf("Successfully loaded descriptor for %s from local file", module.Id))
				} else {
					slog.Info(ms.Action.Name, "text", fmt.Sprintf("Fetching module descriptor for %s from %s", module.Id, moduleDescriptorUrl))
					moduleDescriptorsMapResp, err := ms.HTTPClient.DoGetDecodeReturnAny(moduleDescriptorUrl, map[string]string{})
					if err != nil {
						return err
					}

					extract.ModuleDescriptorsMap[module.Id] = moduleDescriptorsMapResp
				}
			}

			if okBackend {
				serverPort := strconv.Itoa(backendModule.ModuleServerPort)
				backendModule := map[string]string{"id": module.Id, "name": module.Name, "version": *module.Version}

				if applicationFetchDescriptors || isLocalModule {
					backendModuleDescriptors = append(backendModuleDescriptors, extract.ModuleDescriptorsMap[module.Id])
				} else {
					backendModule["url"] = moduleDescriptorUrl
				}

				backendModules = append(backendModules, backendModule)

				sidecarUrl := fmt.Sprintf("http://%s.eureka:%s", module.SidecarName, serverPort)

				discoveryModules = append(discoveryModules, map[string]string{"id": module.Id, "name": module.Name, "version": *module.Version, "location": sidecarUrl})
			} else if okFrontend {
				frontendModule := map[string]string{"id": module.Id, "name": module.Name, "version": *module.Version}
				if applicationFetchDescriptors {
					frontendModuleDescriptors = append(frontendModuleDescriptors, extract.ModuleDescriptorsMap[module.Id])
				} else {
					frontendModule["url"] = moduleDescriptorUrl
				}

				frontendModules = append(frontendModules, frontendModule)
			}

			slog.Info(ms.Action.Name, "text", fmt.Sprintf("Adding module to application, module: %s, version: %s", module.Name, *module.Version))
		}
	}

	applicationBytes, err := json.Marshal(map[string]any{
		"id":                  applicationId,
		"name":                applicationName,
		"version":             applicationVersion,
		"description":         "Default",
		"platform":            applicationPlatform,
		"dependencies":        dependencies,
		"modules":             backendModules,
		"uiModules":           frontendModules,
		"moduleDescriptors":   backendModuleDescriptors,
		"uiModuleDescriptors": frontendModuleDescriptors,
	})
	if err != nil {
		return err
	}

	err = ms.HTTPClient.DoPostReturnNoContent(fmt.Sprintf(ms.Action.GatewayURL, constant.KongPort, "/applications?check=true"), applicationBytes, map[string]string{})
	if err != nil {
		return err
	}

	slog.Info(ms.Action.Name, "text", fmt.Sprintf("Created %s application", applicationId))

	if len(discoveryModules) > 0 {
		applicationDiscoveryBytes, err := json.Marshal(map[string]any{"discovery": discoveryModules})
		if err != nil {
			return err
		}

		err = ms.HTTPClient.DoPostReturnNoContent(fmt.Sprintf(ms.Action.GatewayURL, constant.KongPort, "/modules/discovery"), applicationDiscoveryBytes, map[string]string{})
		if err != nil {
			return err
		}
	}

	slog.Info(ms.Action.Name, "text", fmt.Sprintf("Created %d entries of application module discovery", len(discoveryModules)))

	return nil
}

func (ms *ManagementStep) RemoveApplication(applicationId string) {
	requestURL := fmt.Sprintf(ms.Action.GatewayURL, constant.KongPort, fmt.Sprintf("/applications/%s", applicationId))

	_ = ms.HTTPClient.DoDelete(requestURL, map[string]string{})

	slog.Info(ms.Action.Name, "text", fmt.Sprintf("Removed %s application", applicationId))
}

func (ms *ManagementStep) UpdateModuleDiscovery(id string, sidecarUrl string, restore bool, portServer string) error {
	id = strings.ReplaceAll(id, ":", "-")
	name := helpers.GetModuleNameFromID(id)
	if sidecarUrl == "" || restore {
		if strings.HasPrefix(name, "edge") {
			sidecarUrl = fmt.Sprintf("http://%s.eureka:%s", name, portServer)
		} else {
			sidecarUrl = fmt.Sprintf("http://%s-sc.eureka:%s", name, portServer)
		}
	}

	applicationDiscoveryBytes, err := json.Marshal(map[string]any{
		"id":       id,
		"name":     name,
		"version":  helpers.GetModuleVersionFromID(id),
		"location": sidecarUrl,
	})
	if err != nil {
		return err
	}

	requestURL := fmt.Sprintf(ms.Action.GatewayURL, constant.KongPort, fmt.Sprintf("/modules/%s/discovery", id))

	ms.HTTPClient.DoPutReturnNoContent(requestURL, applicationDiscoveryBytes, map[string]string{})

	slog.Info(ms.Action.Name, "text", fmt.Sprintf("Updated application module discovery for %s module with %s sidecar URL", name, sidecarUrl))

	return nil
}
