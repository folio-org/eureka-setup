package managementsvc

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"strings"

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/folio-org/eureka-cli/httpclient"
	"github.com/folio-org/eureka-cli/models"
	"github.com/folio-org/eureka-cli/tenantsvc"
)

type ManagementProcessor interface {
	ManagementApplicationManager
	ManagementTenantManager
	ManagementTenantEntitlementManager
}

type ManagementApplicationManager interface {
	GetApplications() (models.Applications, error)
	CreateApplications(extract *models.RegistryModuleExtract) error
	RemoveApplication(applicationID string) error
	UpdateModuleDiscovery(id string, sidecarURL string, restore bool, portServer string) error
}

type ManagementSvc struct {
	Action     *action.Action
	HTTPClient httpclient.HTTPClientRunner
	TenantSvc  tenantsvc.TenantProcessor
}

func New(action *action.Action, httpClient httpclient.HTTPClientRunner, tenantSvc tenantsvc.TenantProcessor) *ManagementSvc {
	return &ManagementSvc{Action: action, HTTPClient: httpClient, TenantSvc: tenantSvc}
}

func (ms *ManagementSvc) GetApplications() (models.Applications, error) {
	requestURL := ms.Action.GetRequestURL(constant.KongPort, "/applications")
	resp, err := ms.HTTPClient.GetReturnResponse(requestURL, map[string]string{})
	if err != nil {
		return models.Applications{}, err
	}
	defer httpclient.CloseResponse(resp)
	if resp == nil {
		return models.Applications{}, nil
	}

	var applications models.Applications
	err = json.NewDecoder(resp.Body).Decode(&applications)
	if err != nil && !errors.Is(err, io.EOF) {
		return models.Applications{}, err
	}

	return applications, nil
}

func (ms *ManagementSvc) CreateApplications(extract *models.RegistryModuleExtract) error {
	var (
		backendModules            []map[string]string
		frontendModules           []map[string]string
		backendModuleDescriptors  []any
		frontendModuleDescriptors []any
		dependencies              map[string]any
		discoveryModules          []map[string]string
	)
	applicationName := ms.Action.ConfigApplication["name"].(string)
	applicationVersion := ms.Action.ConfigApplication["version"].(string)
	applicationPlatform := ms.Action.ConfigApplication["platform"].(string)
	applicationFetchDescriptors := ms.Action.ConfigApplication["fetch-descriptors"].(bool)
	applicationID := fmt.Sprintf("%s-%s", applicationName, applicationVersion)
	if ms.Action.ConfigApplication["dependencies"] != nil {
		dependencies = ms.Action.ConfigApplication["dependencies"].(map[string]any)
	}

	for registryName, rr := range extract.RegistryModules {
		if len(rr) > 0 {
			slog.Info(ms.Action.Name, "text", "Including modules to application", "registry", registryName, "application", applicationID)
		}

		for _, module := range rr {
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
				module.ID = fmt.Sprintf("%s-%s", module.Name, *module.Version)
			}

			moduleDescriptorURL := fmt.Sprintf("%s/_/proxy/modules/%s", extract.RegistryURLs[constant.FolioRegistry], module.ID)
			isLocalModule := okBackend && backendModule.LocalDescriptorPath != ""
			if applicationFetchDescriptors || isLocalModule {
				if isLocalModule {
					slog.Info(ms.Action.Name, "text", "Fetching local module descriptor from file path", "module", module.ID)

					var descriptor map[string]any
					err := helpers.ReadJsonFromFile(ms.Action.Name, backendModule.LocalDescriptorPath, &descriptor)
					if err != nil {
						return err
					}
					extract.ModuleDescriptorsMap[module.ID] = descriptor
					slog.Info(ms.Action.Name, "text", "Successfully loaded descriptor from local file", "module", module.ID)
				} else {
					slog.Info(ms.Action.Name, "text", "Fetching module descriptor from URL", "module", module.ID, "url", moduleDescriptorURL)
					moduleDescriptorsMapResp, err := ms.HTTPClient.GetRetryDecodeReturnAny(moduleDescriptorURL, map[string]string{})
					if err != nil {
						return err
					}
					extract.ModuleDescriptorsMap[module.ID] = moduleDescriptorsMapResp
				}
			}

			if okBackend {
				serverPort := strconv.Itoa(backendModule.ModuleServerPort)
				backendModule := map[string]string{
					"id":      module.ID,
					"name":    module.Name,
					"version": *module.Version,
				}
				if applicationFetchDescriptors || isLocalModule {
					backendModuleDescriptors = append(backendModuleDescriptors, extract.ModuleDescriptorsMap[module.ID])
				} else {
					backendModule["url"] = moduleDescriptorURL
				}

				backendModules = append(backendModules, backendModule)
				sidecarUrl := fmt.Sprintf("http://%s.eureka:%s", module.SidecarName, serverPort)
				discoveryModules = append(discoveryModules, map[string]string{
					"id":       module.ID,
					"name":     module.Name,
					"version":  *module.Version,
					"location": sidecarUrl,
				})
			} else if okFrontend {
				frontendModule := map[string]string{
					"id":      module.ID,
					"name":    module.Name,
					"version": *module.Version,
				}
				if applicationFetchDescriptors {
					frontendModuleDescriptors = append(frontendModuleDescriptors, extract.ModuleDescriptorsMap[module.ID])
				} else {
					frontendModule["url"] = moduleDescriptorURL
				}
				frontendModules = append(frontendModules, frontendModule)
			}
			slog.Info(ms.Action.Name, "text", "Including module into application", "module", module.Name, "version", *module.Version, "application", applicationID)
		}
	}

	payload1, err := json.Marshal(map[string]any{
		"id":                  applicationID,
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

	err = ms.HTTPClient.PostReturnNoContent(ms.Action.GetRequestURL(constant.KongPort, "/applications?check=true"), payload1, map[string]string{})
	if err != nil {
		return err
	}
	slog.Info(ms.Action.Name, "text", "Created application", "application", applicationID)

	if len(discoveryModules) > 0 {
		payload2, err := json.Marshal(map[string]any{
			"discovery": discoveryModules,
		})
		if err != nil {
			return err
		}

		err = ms.HTTPClient.PostReturnNoContent(ms.Action.GetRequestURL(constant.KongPort, "/modules/discovery"), payload2, map[string]string{})
		if err != nil {
			return err
		}
	}
	slog.Info(ms.Action.Name, "text", "Created entries of application module discovery", "count", len(discoveryModules))

	return nil
}

func (ms *ManagementSvc) RemoveApplication(applicationID string) error {
	requestURL := ms.Action.GetRequestURL(constant.KongPort, fmt.Sprintf("/applications/%s", applicationID))
	err := ms.HTTPClient.Delete(requestURL, map[string]string{})
	if err != nil {
		return err
	}
	slog.Info(ms.Action.Name, "text", "Removed application", "application", applicationID)

	return nil
}

func (ms *ManagementSvc) UpdateModuleDiscovery(id string, sidecarURL string, restore bool, portServer string) error {
	id = strings.ReplaceAll(id, ":", "-")
	name := helpers.GetModuleNameFromID(id)
	if sidecarURL == "" || restore {
		if strings.HasPrefix(name, "edge") {
			sidecarURL = fmt.Sprintf("http://%s.eureka:%s", name, portServer)
		} else {
			sidecarURL = fmt.Sprintf("http://%s-sc.eureka:%s", name, portServer)
		}
	}

	payload, err := json.Marshal(map[string]any{
		"id":       id,
		"name":     name,
		"version":  helpers.GetModuleVersionFromID(id),
		"location": sidecarURL,
	})
	if err != nil {
		return err
	}

	requestURL := ms.Action.GetRequestURL(constant.KongPort, fmt.Sprintf("/modules/%s/discovery", id))
	err = ms.HTTPClient.PutReturnNoContent(requestURL, payload, map[string]string{})
	if err != nil {
		return err
	}
	slog.Info(ms.Action.Name, "text", "Updated application module discovery with sidecar URL", "module", name, "sidecarUrl", sidecarURL)

	return nil
}
