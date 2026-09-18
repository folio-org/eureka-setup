package registrysvc

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
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
	GetModules(verbose bool, forceRefresh bool) (*models.ProxyModulesByRegistry, error)
	ResolveModuleMetadata(modules *models.ProxyModulesByRegistry)
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

func (rs *RegistrySvc) ResolveModuleMetadata(modules *models.ProxyModulesByRegistry) {
	moduleSets := [][]*models.ProxyModule{modules.FolioModules, modules.EurekaModules}
	for _, moduleSet := range moduleSets {
		for i, module := range moduleSet {
			if module.ID == "okapi" {
				continue
			}
			// Name and version stated by the application descriptor win over what the ID parses to,
			// since the ID pattern does not cover every version the descriptor schema allows
			if module.Metadata.Name == "" {
				module.Metadata.Name = helpers.GetModuleNameFromID(module.ID)
			}
			if module.Metadata.Version == nil {
				module.Metadata.Version = helpers.GetOptionalModuleVersion(module.ID)
			}
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

func (rs *RegistrySvc) GetModules(verbose bool, forceRefresh bool) (*models.ProxyModulesByRegistry, error) {
	var (
		moduleVersions    []models.ApplicationModule
		moduleDescriptors map[string]any
		err               error
	)
	if rs.Action.ConfigApplicationDescriptor != "" {
		moduleVersions, moduleDescriptors, err = rs.readDescriptorInventory()
	} else {
		moduleVersions, err = rs.getFlattenedModuleVersions(forceRefresh)
	}
	if err != nil {
		return nil, err
	}

	var folioModules, eurekaModules []*models.ProxyModule
	moduleDescriptorURLs := make(map[string]string)
	for _, m := range moduleVersions {
		proxy := &models.ProxyModule{
			ID:       m.ID,
			Action:   "enable",
			Metadata: models.ProxyModuleMetadata{Name: m.Name},
		}
		if m.Version != "" {
			version := m.Version
			proxy.Metadata.Version = &version
		}
		if m.URL != "" {
			moduleDescriptorURLs[m.ID] = m.URL
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
		FolioModules:         folioModules,
		EurekaModules:        eurekaModules,
		ModuleDescriptors:    moduleDescriptors,
		ModuleDescriptorURLs: moduleDescriptorURLs,
	}, nil
}

// readDescriptorInventory builds the inventory from application.descriptor plus, when lsp.url is set, the
// eureka-components of the platform descriptor, so that the sidecar version resolves as for any other profile
func (rs *RegistrySvc) readDescriptorInventory() ([]models.ApplicationModule, map[string]any, error) {
	if rs.Action.Param.SkipRegistry {
		slog.Info(rs.Action.Name, "text", "Skip registry flag has no effect, application.descriptor is read on every run")
	}
	modules, moduleDescriptors, err := rs.readApplicationDescriptor(rs.Action.ConfigApplicationDescriptor)
	if err != nil {
		return nil, nil, err
	}
	if rs.Action.ConfigLspURL == "" {
		slog.Info(rs.Action.Name, "text", "No lsp.url set, sidecar version must be set via sidecar-module.version")
		return modules, moduleDescriptors, nil
	}

	descriptor, err := rs.fetchPlatformDescriptor()
	if err != nil {
		return nil, nil, err
	}

	return append(modules, platformComponents(descriptor)...), moduleDescriptors, nil
}

// readApplicationDescriptor uses an application descriptor (local file or URL) as the module inventory instead of
// the LSP platform descriptor and FAR. The descriptor is the same JSON that is registered in mgr-applications in a
// real deployment, so a profile backed by one mirrors that application. Embedded module descriptors are returned
// keyed by module ID so they can be inlined into the application registration.
func (rs *RegistrySvc) readApplicationDescriptor(source string) ([]models.ApplicationModule, map[string]any, error) {
	var descriptor models.ApplicationDescriptorSource
	if helpers.IsURL(source) {
		if err := rs.HTTPClient.GetRetryReturnStruct(source, map[string]string{}, &descriptor); err != nil {
			return nil, nil, appErrors.ApplicationDescriptorReadFailed(source, err)
		}
	} else {
		path, err := helpers.ExpandHomeDir(source)
		if err != nil {
			return nil, nil, err
		}
		if err := helpers.ReadJSONFromFile(path, &descriptor); err != nil {
			return nil, nil, appErrors.ApplicationDescriptorReadFailed(source, err)
		}
	}
	slog.Info(rs.Action.Name, "text", "Read application descriptor", "source", source, "id", descriptor.ID)
	if descriptor.Name != "" && descriptor.Name != rs.Action.ConfigApplicationName {
		slog.Warn(rs.Action.Name, "text", "Application descriptor name differs from application.name in the profile", "descriptor", descriptor.Name, "profile", rs.Action.ConfigApplicationName)
	}

	var modules []models.ApplicationModule
	for _, module := range append(descriptor.Modules, descriptor.UIModules...) {
		if module.ID == "" {
			module.ID = fmt.Sprintf("%s-%s", module.Name, module.Version)
		}
		modules = append(modules, module)
	}
	if len(modules) == 0 {
		return nil, nil, appErrors.ApplicationDescriptorNoModules(source)
	}

	moduleDescriptors := make(map[string]any)
	for _, moduleDescriptor := range append(descriptor.ModuleDescriptors, descriptor.UIModuleDescriptors...) {
		if id := helpers.GetString(moduleDescriptor, "id"); id != "" {
			moduleDescriptors[id] = moduleDescriptor
		}
	}
	slog.Info(rs.Action.Name, "text", "Using application descriptor as module inventory", "modules", len(modules), "embeddedDescriptors", len(moduleDescriptors))

	return modules, moduleDescriptors, nil
}

func (rs *RegistrySvc) getFlattenedModuleVersions(forceRefresh bool) ([]models.ApplicationModule, error) {
	homeDir, err := helpers.GetHomeDirPath()
	if err != nil {
		return nil, err
	}
	filePath := filepath.Join(homeDir, constant.ModulesFile)

	if rs.Action.Param.SkipRegistry {
		return rs.readModulesLocalFile(filePath)
	}
	if !forceRefresh {
		if info, statErr := os.Stat(filePath); statErr == nil && info.Mode().IsRegular() {
			return rs.readModulesLocalFile(filePath)
		}
	}

	return rs.fetchAndPersistModuleVersions(filePath)
}

func (rs *RegistrySvc) readModulesLocalFile(path string) ([]models.ApplicationModule, error) {
	if err := helpers.IsRegularFile(path); err != nil {
		return nil, appErrors.LocalInstallFileNotFound(err)
	}

	var modules []models.ApplicationModule
	if err := helpers.ReadJSONFromFile(path, &modules); err != nil {
		return nil, err
	}
	if modules == nil {
		modules = make([]models.ApplicationModule, 0)
	}
	slog.Info(rs.Action.Name, "text", "Read module versions from a local file", "file", constant.ModulesFile)

	return modules, nil
}

func (rs *RegistrySvc) fetchAndPersistModuleVersions(filePath string) ([]models.ApplicationModule, error) {
	descriptor, err := rs.fetchPlatformDescriptor()
	if err != nil {
		return nil, err
	}

	applications := append(descriptor.Applications.Required, descriptor.Applications.Optional...)
	applications = append(applications, descriptor.Applications.Experimental...)

	modules := platformComponents(descriptor)

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
							URL:     helpers.GetString(entry, "url"),
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

	if modules == nil {
		modules = make([]models.ApplicationModule, 0)
	}

	if err := helpers.WriteJSONToFile(filePath, modules); err != nil {
		return nil, err
	}
	slog.Info(rs.Action.Name, "text", "Persisted module versions to a local file", "file", constant.ModulesFile)

	return modules, nil
}

func (rs *RegistrySvc) fetchPlatformDescriptor() (*models.PlatformDescriptor, error) {
	var descriptor models.PlatformDescriptor
	if err := rs.HTTPClient.GetRetryReturnStruct(rs.Action.ConfigLspURL, map[string]string{}, &descriptor); err != nil {
		return nil, err
	}
	slog.Info(rs.Action.Name, "text", "Fetched LSP platform descriptor", "name", descriptor.Name, "version", descriptor.Version)

	return &descriptor, nil
}

// platformComponents returns the eureka-components of a platform descriptor (management modules, sidecar, Kong, Keycloak)
func platformComponents(descriptor *models.PlatformDescriptor) []models.ApplicationModule {
	var modules []models.ApplicationModule
	for _, component := range descriptor.EurekaComponents {
		modules = append(modules, models.ApplicationModule{
			ID:      fmt.Sprintf("%s-%s", component.Name, component.Version),
			Name:    component.Name,
			Version: component.Version,
		})
	}

	return modules
}

func isEurekaModule(name string) bool {
	return strings.HasSuffix(name, "-keycloak") ||
		strings.HasPrefix(name, "mgr-") ||
		name == "folio-kong" ||
		name == "folio-module-sidecar" ||
		name == "mod-scheduler"
}
