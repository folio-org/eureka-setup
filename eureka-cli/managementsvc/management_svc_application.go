package managementsvc

import (
	"encoding/json"
	"fmt"
	"log/slog"
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
	CreateApplications(extract *models.RegistryExtract) error
	RemoveApplication(applicationID string) error
	GetModuleDiscovery(name string) (models.ModuleDiscoveryResponse, error)
	UpdateModuleDiscovery(id string, restore bool, privatePort int, sidecarURL string) error
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

	var applications models.ApplicationsResponse
	err := ms.HTTPClient.GetReturnStruct(requestURL, map[string]string{}, &applications)
	if err != nil {
		return models.ApplicationsResponse{}, err
	}

	return applications, nil
}

func (ms *ManagementSvc) CreateApplications(extract *models.RegistryExtract) error {
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

	allModules := [][]*models.ProxyModule{extract.Modules.FolioModules, extract.Modules.EurekaModules}
	for _, modules := range allModules {
		for _, module := range modules {
			if strings.Contains(module.Name, constant.ManagementModulePattern) {
				continue
			}

			backendModule, existsBackend := extract.BackendModules[module.Name]
			frontendModule, existsFrontend := extract.FrontendModules[module.Name]
			if (!existsBackend && !existsFrontend) || (existsBackend && !backendModule.DeployModule || existsFrontend && !frontendModule.DeployModule) {
				continue
			}
			if existsBackend && backendModule.ModuleVersion != nil || existsFrontend && frontendModule.ModuleVersion != nil {
				if backendModule.ModuleVersion != nil {
					module.Version = backendModule.ModuleVersion
				} else if frontendModule.ModuleVersion != nil {
					module.Version = frontendModule.ModuleVersion
				}
				module.ID = fmt.Sprintf("%s-%s", module.Name, *module.Version)
			}

			moduleDescriptorURL := fmt.Sprintf("%s/_/proxy/modules/%s", extract.RegistryURLs[constant.FolioRegistry], module.ID)
			isLocalBackendModule := existsBackend && backendModule.LocalDescriptorPath != ""
			isLocalFrontendModule := existsFrontend && frontendModule.LocalDescriptorPath != ""
			isLocalModule := isLocalBackendModule || isLocalFrontendModule
			if ms.Action.ConfigApplicationFetchDescriptors || isLocalModule {
				var descriptorPath string
				if isLocalBackendModule {
					descriptorPath = backendModule.LocalDescriptorPath
				} else if isLocalFrontendModule {
					descriptorPath = frontendModule.LocalDescriptorPath
				}

				err := ms.FetchModuleDescriptor(extract, module.ID, moduleDescriptorURL, descriptorPath, isLocalModule)
				if err != nil {
					return err
				}
			}

			if existsBackend {
				newBackendModule := map[string]string{
					"id":      module.ID,
					"name":    module.Name,
					"version": *module.Version,
				}
				if ms.Action.ConfigApplicationFetchDescriptors || isLocalModule {
					backendModuleDescriptors = append(backendModuleDescriptors, extract.ModuleDescriptors[module.ID])
				} else {
					newBackendModule["url"] = moduleDescriptorURL
				}
				backendModules = append(backendModules, newBackendModule)

				sidecarURL := fmt.Sprintf("http://%s.eureka:%d", module.SidecarName, backendModule.PrivatePort)
				discoveryModules = append(discoveryModules, map[string]string{
					"id":       module.ID,
					"name":     module.Name,
					"version":  *module.Version,
					"location": sidecarURL,
				})
			} else if existsFrontend {
				newFrontendModule := map[string]string{
					"id":      module.ID,
					"name":    module.Name,
					"version": *module.Version,
				}
				if ms.Action.ConfigApplicationFetchDescriptors {
					frontendModuleDescriptors = append(frontendModuleDescriptors, extract.ModuleDescriptors[module.ID])
				} else {
					newFrontendModule["url"] = moduleDescriptorURL
				}
				frontendModules = append(frontendModules, newFrontendModule)
			}
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

	var appResponse models.ApplicationDescriptor
	appRequestURL := ms.Action.GetRequestURL(constant.KongPort, "/applications?check=true")
	err = ms.HTTPClient.PostReturnStruct(appRequestURL, payload1, map[string]string{}, &appResponse)
	if err != nil {
		return err
	}
	slog.Info(ms.Action.Name, "text", "Created application", "id", appResponse.ID, "backendModules", len(backendModules), "frontendModules", len(frontendModules))

	if len(discoveryModules) > 0 {
		payload2, err := json.Marshal(map[string]any{
			"discovery": discoveryModules,
		})
		if err != nil {
			return err
		}

		var discoveryResponse models.ModuleDiscoveryResponse
		discoveryRequestURL := ms.Action.GetRequestURL(constant.KongPort, "/modules/discovery")
		err = ms.HTTPClient.PostReturnStruct(discoveryRequestURL, payload2, map[string]string{}, &discoveryResponse)
		if err != nil {
			return err
		}
		slog.Info(ms.Action.Name, "text", "Created module discovery", "count", len(discoveryModules), "totalRecords", discoveryResponse.TotalRecords)
	}

	return nil
}

func (ms *ManagementSvc) FetchModuleDescriptor(extract *models.RegistryExtract, moduleID, moduleDescriptorURL, localPath string, isLocalModule bool) error {
	if isLocalModule {
		slog.Info(ms.Action.Name, "text", "Fetching local module descriptor", "module", moduleID)

		var moduleDescriptorData map[string]any
		err := helpers.ReadJsonFromFile(ms.Action.Name, localPath, &moduleDescriptorData)
		if err != nil {
			return err
		}
		extract.ModuleDescriptors[moduleID] = moduleDescriptorData
		slog.Info(ms.Action.Name, "text", "Loaded module descriptor", "module", moduleID)

		return nil
	}
	slog.Info(ms.Action.Name, "text", "Fetching module descriptor", "module", moduleID, "url", moduleDescriptorURL)

	var decodedResponse any
	if err := ms.HTTPClient.GetRetryReturnStruct(moduleDescriptorURL, map[string]string{}, &decodedResponse); err != nil {
		return err
	}
	extract.ModuleDescriptors[moduleID] = decodedResponse
	slog.Info(ms.Action.Name, "text", "Loaded module descriptor", "module", moduleID, "url", moduleDescriptorURL)

	return nil
}

func (ms *ManagementSvc) RemoveApplication(applicationID string) error {
	requestURL := ms.Action.GetRequestURL(constant.KongPort, fmt.Sprintf("/applications/%s", applicationID))

	var appResponse models.ApplicationDescriptor
	err := ms.HTTPClient.DeleteReturnStruct(requestURL, map[string]string{}, &appResponse)
	if err != nil {
		return err
	}
	slog.Info(ms.Action.Name, "text", "Removed application", "id", appResponse.ID)

	return nil
}

func (ms *ManagementSvc) GetModuleDiscovery(name string) (models.ModuleDiscoveryResponse, error) {
	requestURL := ms.Action.GetRequestURL(constant.KongPort, fmt.Sprintf("/modules/discovery?query=name==%s&limit=1", name))

	var moduleDiscovery models.ModuleDiscoveryResponse
	err := ms.HTTPClient.GetReturnStruct(requestURL, map[string]string{}, &moduleDiscovery)
	if err != nil {
		return models.ModuleDiscoveryResponse{}, err
	}

	return moduleDiscovery, nil
}

func (ms *ManagementSvc) UpdateModuleDiscovery(id string, restore bool, privatePort int, sidecarURL string) error {
	name := helpers.GetModuleNameFromID(id)
	if sidecarURL == "" || restore {
		if strings.HasPrefix(name, "edge") {
			sidecarURL = fmt.Sprintf("http://%s.eureka:%d", name, privatePort)
		} else {
			sidecarURL = fmt.Sprintf("http://%s-sc.eureka:%d", name, privatePort)
		}
	}

	version := helpers.GetModuleVersionFromID(id)
	payload, err := json.Marshal(map[string]any{
		"id":       id,
		"name":     name,
		"version":  version,
		"location": sidecarURL,
	})
	if err != nil {
		return err
	}

	requestURL := ms.Action.GetRequestURL(constant.KongPort, fmt.Sprintf("/modules/%s/discovery", id))

	var discoveryResponse models.ModuleDiscovery
	err = ms.HTTPClient.PutReturnStruct(requestURL, payload, map[string]string{}, &discoveryResponse)
	if err != nil {
		return err
	}
	slog.Info(ms.Action.Name, "text", "Updated module discovery", "module", discoveryResponse.Name, "location", discoveryResponse.Location)

	return nil
}
