package registrysvc

import (
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/folio-org/eureka-setup/eureka-cli/action"
	"github.com/folio-org/eureka-setup/eureka-cli/awssvc"
	"github.com/folio-org/eureka-setup/eureka-cli/constant"
	appErrors "github.com/folio-org/eureka-setup/eureka-cli/errors"
	"github.com/folio-org/eureka-setup/eureka-cli/helpers"
	"github.com/folio-org/eureka-setup/eureka-cli/httpclient"
	"github.com/folio-org/eureka-setup/eureka-cli/models"
)

// RegistryProcessor defines the interface for registry-related operations
type RegistryProcessor interface {
	GetNamespace(version string) string
	GetModules(verbose bool) (*models.ProxyModulesByRegistry, error)
	ExtractModuleMetadata(modules *models.ProxyModulesByRegistry)
	GetAuthorizationToken() (string, error)
}

// RegistrySvc provides functionality for interacting with module registries
type RegistrySvc struct {
	Action     *action.Action
	HTTPClient httpclient.HTTPClientRunner
	AWSSvc     awssvc.AWSProcessor
}

// New creates a new RegistrySvc instance
func New(action *action.Action, httpClient httpclient.HTTPClientRunner, awsSvc awssvc.AWSProcessor) *RegistrySvc {
	return &RegistrySvc{Action: action, HTTPClient: httpClient, AWSSvc: awsSvc}
}

func (rs *RegistrySvc) GetAuthorizationToken() (string, error) {
	return rs.AWSSvc.GetAuthorizationToken()
}

func (rs *RegistrySvc) GetNamespace(version string) string {
	ecrNamespace := rs.AWSSvc.GetECRNamespace()
	if ecrNamespace != "" {
		return ecrNamespace
	}
	if strings.Contains(version, "SNAPSHOT") {
		return constant.SnapshotNamespace
	} else {
		return constant.ReleaseNamespace
	}
}

func (rs *RegistrySvc) ExtractModuleMetadata(modules *models.ProxyModulesByRegistry) {
	allModules := [][]*models.ProxyModule{modules.FolioModules, modules.EurekaModules}
	for _, moduleSet := range allModules {
		for i, module := range moduleSet {
			if module.ID == "okapi" {
				continue
			}
			module.Metadata.Name = helpers.GetModuleNameFromID(module.ID)
			module.Metadata.Version = helpers.GetOptionalModuleVersion(module.ID)
			module.Metadata.SidecarName = rs.getSidecarName(module)
			moduleSet[i] = module
		}
	}
}

func (rs *RegistrySvc) getSidecarName(module *models.ProxyModule) string {
	if strings.HasPrefix(module.Metadata.Name, "edge") {
		return module.Metadata.Name
	} else {
		return helpers.GetSidecarName(module.Metadata.Name)
	}
}

func (rs *RegistrySvc) GetModules(verbose bool) (*models.ProxyModulesByRegistry, error) {
	allModules, err := rs.getFlattenedModuleVersions()
	if err != nil {
		return nil, err
	}

	var folioModules, eurekaModules []*models.ProxyModule
	for _, m := range allModules {
		proxy := &models.ProxyModule{
			ID:     m.ID,
			Action: "enable",
		}
		if isEurekaModule(m.Name) {
			eurekaModules = append(eurekaModules, proxy)
		} else {
			folioModules = append(folioModules, proxy)
		}
	}

	if verbose {
		slog.Info(rs.Action.Name, "text", "LSP modules split", "folio", len(folioModules), "eureka", len(eurekaModules))
	}

	return &models.ProxyModulesByRegistry{
		FolioModules:  folioModules,
		EurekaModules: eurekaModules,
	}, nil
}

func (rs *RegistrySvc) getFlattenedModuleVersions() ([]models.ApplicationModule, error) {
	var descriptor models.PlatformDescriptor
	if err := rs.HTTPClient.GetRetryReturnStruct(rs.Action.ConfigLspURL, map[string]string{}, &descriptor); err != nil {
		return nil, err
	}

	slog.Info(rs.Action.Name, "text", "Fetched LSP platform descriptor", "name", descriptor.Name, "version", descriptor.Version)

	applications := append(descriptor.Applications.Required, descriptor.Applications.Optional...)
	applications = append(applications, descriptor.Applications.Experimental...)

	var modules []models.ApplicationModule
	for _, component := range descriptor.EurekaComponents {
		modules = append(modules, models.ApplicationModule{
			ID:      fmt.Sprintf("%s-%s", component.Name, component.Version),
			Name:    component.Name,
			Version: component.Version,
		})
	}

	type result struct {
		modules []models.ApplicationModule
		err     error
		appID   string
	}

	results := make([]result, len(applications))
	var wg sync.WaitGroup
	wg.Add(len(applications))
	for idx, app := range applications {
		// The decision of fetching what application for what module is irrelevant, fetch everything
		// and only then decide what to materialise into the realm of existence as a container
		go func(innerIdx int, innerApp models.PlatformApplication) {
			defer wg.Done()
			appID := fmt.Sprintf("%s-%s", innerApp.Name, innerApp.Version)
			farURL := fmt.Sprintf("%s/applications?query=id==%s", rs.Action.ConfigFarURL, appID)

			var response models.ApplicationsResponse
			if err := rs.HTTPClient.GetRetryReturnStruct(farURL, map[string]string{}, &response); err != nil {
				results[innerIdx] = result{appID: appID, err: err}
				return
			}
			slog.Info(rs.Action.Name, "text", "Fetched FAR application descriptor", "appId", appID)

			var appModules []models.ApplicationModule
			for _, appDescriptor := range response.ApplicationDescriptors {
				for _, key := range []string{"modules", "uiModules"} {
					for _, raw := range helpers.GetAnySlice(appDescriptor, key) {
						entry, ok := raw.(map[string]any)
						if !ok {
							continue
						}
						appModules = append(appModules, models.ApplicationModule{
							ID:      helpers.GetString(entry, "id"),
							Name:    helpers.GetString(entry, "name"),
							Version: helpers.GetString(entry, "version"),
						})
					}
				}
			}
			results[innerIdx] = result{appID: appID, modules: appModules}
		}(idx, app)
	}
	wg.Wait()

	for _, r := range results {
		if r.err != nil {
			return nil, appErrors.FARFetchFailed(r.appID, r.err)
		}
		modules = append(modules, r.modules...)
	}

	return modules, nil
}

func isEurekaModule(name string) bool {
	return strings.HasSuffix(name, "-keycloak") ||
		strings.HasPrefix(name, "mgr-") ||
		name == "folio-kong" ||
		name == "folio-module-sidecar" ||
		name == "mod-scheduler"
}
