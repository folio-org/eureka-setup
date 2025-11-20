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
	headers, err := helpers.SecureApplicationJSONHeaders(ms.Action.KeycloakMasterAccessToken)
	if err != nil {
		return models.ApplicationsResponse{}, err
	}

	var decodedResponse models.ApplicationsResponse
	if err := ms.HTTPClient.GetReturnStruct(requestURL, headers, &decodedResponse); err != nil {
		return models.ApplicationsResponse{}, err
	}

	return decodedResponse, nil
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

	headers, err := helpers.SecureApplicationJSONHeaders(ms.Action.KeycloakMasterAccessToken)
	if err != nil {
		return err
	}

	allModules := [][]*models.ProxyModule{extract.Modules.FolioModules, extract.Modules.EurekaModules}
	for _, modules := range allModules {
		for _, module := range modules {
			if strings.Contains(module.Metadata.Name, constant.ManagementModulePattern) {
				continue
			}

			backendModule, existsBackend := extract.BackendModules[module.Metadata.Name]
			frontendModule, existsFrontend := extract.FrontendModules[module.Metadata.Name]
			if (!existsBackend && !existsFrontend) || (existsBackend && !backendModule.DeployModule || existsFrontend && !frontendModule.DeployModule) {
				continue
			}
			if existsBackend && backendModule.ModuleVersion != nil || existsFrontend && frontendModule.ModuleVersion != nil {
				if backendModule.ModuleVersion != nil {
					module.Metadata.Version = backendModule.ModuleVersion
				} else if frontendModule.ModuleVersion != nil {
					module.Metadata.Version = frontendModule.ModuleVersion
				}
				module.ID = fmt.Sprintf("%s-%s", module.Metadata.Name, *module.Metadata.Version)
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
				if err := ms.FetchModuleDescriptor(extract, module.ID, moduleDescriptorURL, descriptorPath, isLocalModule); err != nil {
					return err
				}
			}

			if existsBackend {
				newBackendModule := map[string]string{
					"id":      module.ID,
					"name":    module.Metadata.Name,
					"version": *module.Metadata.Version,
				}
				if ms.Action.ConfigApplicationFetchDescriptors || isLocalModule {
					backendModuleDescriptors = append(backendModuleDescriptors, extract.ModuleDescriptors[module.ID])
				} else {
					newBackendModule["url"] = moduleDescriptorURL
				}
				backendModules = append(backendModules, newBackendModule)

				sidecarURL := fmt.Sprintf("http://%s.eureka:%d", module.Metadata.SidecarName, backendModule.PrivatePort)
				discoveryModules = append(discoveryModules, map[string]string{
					"id":       module.ID,
					"name":     module.Metadata.Name,
					"version":  *module.Metadata.Version,
					"location": sidecarURL,
				})
			} else if existsFrontend {
				newFrontendModule := map[string]string{
					"id":      module.ID,
					"name":    module.Metadata.Name,
					"version": *module.Metadata.Version,
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
	if err := ms.HTTPClient.PostReturnStruct(appRequestURL, payload1, headers, &appResponse); err != nil {
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
		discoveryRequestURL := ms.Action.GetRequestURL(constant.KongPort, "/modules/discovery")

		var discoveryResponse models.ModuleDiscoveryResponse
		if err := ms.HTTPClient.PostReturnStruct(discoveryRequestURL, payload2, headers, &discoveryResponse); err != nil {
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
		if err := helpers.ReadJSONFromFile(localPath, &moduleDescriptorData); err != nil {
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
	headers, err := helpers.SecureApplicationJSONHeaders(ms.Action.KeycloakMasterAccessToken)
	if err != nil {
		return err
	}

	var decodedResponse models.ApplicationDescriptor
	if err := ms.HTTPClient.DeleteReturnStruct(requestURL, headers, &decodedResponse); err != nil {
		return err
	}
	slog.Info(ms.Action.Name, "text", "Removed application", "id", decodedResponse.ID)

	return nil
}

func (ms *ManagementSvc) GetModuleDiscovery(name string) (models.ModuleDiscoveryResponse, error) {
	requestURL := ms.Action.GetRequestURL(constant.KongPort, fmt.Sprintf("/modules/discovery?query=name==%s&limit=1", name))
	headers, err := helpers.SecureApplicationJSONHeaders(ms.Action.KeycloakMasterAccessToken)
	if err != nil {
		return models.ModuleDiscoveryResponse{}, err
	}

	var decodedResponse models.ModuleDiscoveryResponse
	if err := ms.HTTPClient.GetReturnStruct(requestURL, headers, &decodedResponse); err != nil {
		return models.ModuleDiscoveryResponse{}, err
	}

	return decodedResponse, nil
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
	headers, err := helpers.SecureApplicationJSONHeaders(ms.Action.KeycloakMasterAccessToken)
	if err != nil {
		return err
	}

	var moduleDiscovery models.ModuleDiscovery
	if err := ms.HTTPClient.PutReturnStruct(requestURL, payload, headers, &moduleDiscovery); err != nil {
		return err
	}
	slog.Info(ms.Action.Name, "text", "Updated module discovery", "module", moduleDiscovery.Name, "location", moduleDiscovery.Location)

	return nil
}
