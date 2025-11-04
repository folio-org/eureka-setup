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

// ManagementProcessor defines the interface for management service operations
type ManagementProcessor interface {
	ManagementApplicationManager
	ManagementTenantManager
	ManagementTenantEntitlementManager
}

// ManagementApplicationManager defines the interface for application management operations
type ManagementApplicationManager interface {
	GetApplications() (models.ApplicationsResponse, error)
	CreateApplications(extract *models.RegistryModuleExtract) error
	RemoveApplication(applicationID string) error
	GetModuleDiscovery(name string) (models.ModuleDiscoveryResponse, error)
	UpdateModuleDiscovery(id string, sidecarURL string, restore bool) error
}

// ManagementSvc provides functionality for management operations including applications and tenants
type ManagementSvc struct {
	Action     *action.Action
	HTTPClient httpclient.HTTPClientRunner
	TenantSvc  tenantsvc.TenantProcessor
}

// New creates a new ManagementSvc instance
func New(action *action.Action, httpClient httpclient.HTTPClientRunner, tenantSvc tenantsvc.TenantProcessor) *ManagementSvc {
	return &ManagementSvc{Action: action, HTTPClient: httpClient, TenantSvc: tenantSvc}
}

func (ms *ManagementSvc) GetApplications() (models.ApplicationsResponse, error) {
	requestURL := ms.Action.GetRequestURL(constant.KongPort, "/applications")
	httpResponse, err := ms.HTTPClient.GetReturnResponse(requestURL, map[string]string{})
	if err != nil {
		return models.ApplicationsResponse{}, err
	}
	defer httpclient.CloseResponse(httpResponse)
	if httpResponse == nil {
		return models.ApplicationsResponse{}, nil
	}

	var applications models.ApplicationsResponse
	err = json.NewDecoder(httpResponse.Body).Decode(&applications)
	if err != nil && !errors.Is(err, io.EOF) {
		return models.ApplicationsResponse{}, err
	}

	return applications, nil
}

func (ms *ManagementSvc) CreateApplications(extract *models.RegistryModuleExtract) error {
	var (
		backendModules            []map[string]string
		frontendModules           []map[string]string
		discoveryModules          []map[string]string
		dependencies              map[string]any
		backendModuleDescriptors  []any
		frontendModuleDescriptors []any
	)
	if len(ms.Action.ConfigApplicationDependencies) > 0 {
		dependencies = ms.Action.ConfigApplicationDependencies
	}

	for registryName, rr := range extract.RegistryModules {
		if len(rr) > 0 {
			slog.Info(ms.Action.Name, "text", "Including modules to application", "registry", registryName, "application", ms.Action.ConfigApplicationID)
		}

		for _, module := range rr {
			if strings.Contains(module.Name, constant.ManagementModulePattern) {
				continue
			}

			backendModule, okBackend := extract.BackendModules[module.Name]
			frontendModule, okFrontend := extract.FrontendModules[module.Name]
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
			isLocalBackendModule := okBackend && backendModule.LocalDescriptorPath != ""
			isLocalFrontendModule := okFrontend && frontendModule.LocalDescriptorPath != ""
			isLocalModule := isLocalBackendModule || isLocalFrontendModule

			if ms.Action.ConfigApplicationFetchDescriptors || isLocalModule {
				if isLocalModule {
					slog.Info(ms.Action.Name, "text", "Fetching local module descriptor from file path", "module", module.ID)
					var descriptorPath string
					if isLocalBackendModule {
						descriptorPath = backendModule.LocalDescriptorPath
					} else {
						descriptorPath = frontendModule.LocalDescriptorPath
					}

					var descriptor map[string]any
					err := helpers.ReadJsonFromFile(ms.Action.Name, descriptorPath, &descriptor)
					if err != nil {
						return err
					}
					extract.ModuleDescriptors[module.ID] = descriptor
					slog.Info(ms.Action.Name, "text", "Successfully loaded descriptor from local file", "module", module.ID)
				} else {
					slog.Info(ms.Action.Name, "text", "Fetching module descriptor from URL", "module", module.ID, "url", moduleDescriptorURL)
					moduleDescriptorsMapResp, err := ms.HTTPClient.GetRetryDecodeReturnAny(moduleDescriptorURL, map[string]string{})
					if err != nil {
						return err
					}
					extract.ModuleDescriptors[module.ID] = moduleDescriptorsMapResp
					slog.Info(ms.Action.Name, "text", "Successfully loaded descriptor from the remote", "module", module.ID)
				}
			}

			if okBackend {
				serverPort := strconv.Itoa(backendModule.ModuleServerPort)
				backendModule := map[string]string{
					"id":      module.ID,
					"name":    module.Name,
					"version": *module.Version,
				}
				if ms.Action.ConfigApplicationFetchDescriptors || isLocalModule {
					backendModuleDescriptors = append(backendModuleDescriptors, extract.ModuleDescriptors[module.ID])
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
				if ms.Action.ConfigApplicationFetchDescriptors {
					frontendModuleDescriptors = append(frontendModuleDescriptors, extract.ModuleDescriptors[module.ID])
				} else {
					frontendModule["url"] = moduleDescriptorURL
				}
				frontendModules = append(frontendModules, frontendModule)
			}
			slog.Info(ms.Action.Name, "text", "Including module into application", "module", module.Name, "version", *module.Version, "application", ms.Action.ConfigApplicationID)
		}
	}

	payload1, err := json.Marshal(map[string]any{
		"id":                  ms.Action.ConfigApplicationID,
		"name":                ms.Action.ConfigApplicationName,
		"version":             ms.Action.ConfigApplicationVersion,
		"description":         "Default",
		"platform":            ms.Action.ConfigApplicationPlatform,
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
	slog.Info(ms.Action.Name, "text", "Created application", "application", ms.Action.ConfigApplicationID)

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

func (ms *ManagementSvc) GetModuleDiscovery(name string) (models.ModuleDiscoveryResponse, error) {
	requestURL := ms.Action.GetRequestURL(constant.KongPort, fmt.Sprintf("/modules/discovery?query=name==%s&limit=1", name))
	httpResponse, err := ms.HTTPClient.GetReturnResponse(requestURL, map[string]string{})
	if err != nil {
		return models.ModuleDiscoveryResponse{}, err
	}
	defer httpclient.CloseResponse(httpResponse)
	if httpResponse == nil {
		return models.ModuleDiscoveryResponse{}, nil
	}

	var moduleDiscovery models.ModuleDiscoveryResponse
	err = json.NewDecoder(httpResponse.Body).Decode(&moduleDiscovery)
	if err != nil && !errors.Is(err, io.EOF) {
		return models.ModuleDiscoveryResponse{}, err
	}

	return moduleDiscovery, nil
}

func (ms *ManagementSvc) UpdateModuleDiscovery(id string, sidecarURL string, restore bool) error {
	name := helpers.GetModuleNameFromID(id)
	if sidecarURL == "" || restore {
		if strings.HasPrefix(name, "edge") {
			sidecarURL = fmt.Sprintf("http://%s.eureka:%s", name, constant.PrivateServerPort)
		} else {
			sidecarURL = fmt.Sprintf("http://%s-sc.eureka:%s", name, constant.PrivateServerPort)
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
	slog.Info(ms.Action.Name, "text", "Updated application module discovery with new sidecar URL", "module", name, "sidecarUrl", sidecarURL)

	return nil
}
