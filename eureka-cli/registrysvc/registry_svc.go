package registrysvc

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

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
	InjectProfileModules(modules *models.ProxyModulesByRegistry, backendModules map[string]models.BackendModule) (map[string]any, error)
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

func (rs *RegistrySvc) GetModules(verbose bool, forceRefresh bool) (*models.ProxyModulesByRegistry, error) {
	moduleVersions, err := rs.getFlattenedModuleVersions(forceRefresh)
	if err != nil {
		return nil, err
	}

	var folioModules, eurekaModules []*models.ProxyModule
	for _, m := range moduleVersions {
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

func (rs *RegistrySvc) InjectProfileModules(modules *models.ProxyModulesByRegistry, backendModules map[string]models.BackendModule) (map[string]any, error) {
	prepopulatedDescriptors := make(map[string]any)

	// 1. Process Backend Modules
	backendNames := make([]string, 0, len(backendModules))
	for name := range backendModules {
		backendNames = append(backendNames, name)
	}
	rs.injectModules(modules, backendNames, prepopulatedDescriptors, func(name string) string {
		if bMod, ok := backendModules[name]; ok && bMod.ModuleVersion != nil {
			return *bMod.ModuleVersion
		}
		return ""
	}, false)

	// 2. Process Frontend/UI Modules
	frontendNames := make([]string, 0, len(rs.Action.ConfigFrontendModules))
	for name := range rs.Action.ConfigFrontendModules {
		frontendNames = append(frontendNames, name)
	}
	rs.injectModules(modules, frontendNames, prepopulatedDescriptors, func(name string) string {
		if rawMap, ok := rs.Action.ConfigFrontendModules[name].(map[string]any); ok {
			if v, ok := rawMap["version"].(string); ok {
				return v
			}
		}
		return ""
	}, true)

	return prepopulatedDescriptors, nil
}

// Private helper to isolate the shared registry orchestration and injection layer
func (rs *RegistrySvc) injectModules(
	modules *models.ProxyModulesByRegistry,
	names []string,
	prepopulatedDescriptors map[string]any,
	versionExtractor func(string) string,
	isUI bool,
) {
	baseRegistryURL := strings.TrimSpace(rs.Action.ConfigRegistryURL)
	label := "Backend"
	profileKey := "backend-modules"
	if isUI {
		label = "UI"
		profileKey = "frontend-modules"
	}

	for _, name := range names {
		// Check if the module is already tracked in active deployment trees
		found := false
		for _, fm := range modules.FolioModules {
			if fm.Metadata.Name == name || helpers.IsStrictModuleID(fm.ID, name) {
				found = true
				break
			}
		}
		if !found {
			for _, em := range modules.EurekaModules {
				if em.Metadata.Name == name || helpers.IsStrictModuleID(em.ID, name) {
					found = true
					break
				}
			}
		}

		if found {
			continue
		}

		var pristineVersion string
		if explicitVersion := versionExtractor(name); explicitVersion != "" {
			pristineVersion = explicitVersion
			slog.Debug(rs.Action.Name, "text", fmt.Sprintf("%s module injection engine: picked explicit configuration version", label), "module", name, "version", pristineVersion)
		} else {
			// Fallback to active registry discovery if unpinned
			discoveredVersion, lookupErr := rs.discoverLatestRegistryVersion(name)
			if lookupErr == nil && discoveredVersion != "" {
				pristineVersion = discoveredVersion
				slog.Debug(rs.Action.Name, "text", fmt.Sprintf("%s module injection engine: auto-discovered version from registry", label), "module", name, "version", pristineVersion)
			} else if lookupErr != nil {
				slog.Debug(rs.Action.Name, "text", "Module injection engine: registry discovery lookup error occurred", "module", name, "error", lookupErr.Error())
			}
		}

		if pristineVersion == "" {
			slog.Warn(rs.Action.Name, "text", fmt.Sprintf("%s Module declared in profile %s but not found anywhere, skipping injection", label, profileKey), "module", name)
			continue
		}

        gatewayVersion := helpers.ToGatewayVersion(pristineVersion)
		syntheticID := helpers.ToSyntheticID(name, pristineVersion)

		// Descriptor Pre-population Engine
		if baseRegistryURL != "" {
            moduleDescriptorURL := rs.Action.GetModuleURL(syntheticID)
            var descriptorData map[string]any

            descErr := rs.HTTPClient.GetRetryReturnStruct(moduleDescriptorURL, map[string]string{}, &descriptorData)
            if descErr == nil && descriptorData != nil {
                descriptorData["id"] = syntheticID
                prepopulatedDescriptors[syntheticID] = descriptorData
                slog.Debug(rs.Action.Name, "text", fmt.Sprintf("%s Descriptor Pre-population Engine cached record successfully", label), "module", name, "id", syntheticID)
            } else if descErr != nil {
                slog.Debug(rs.Action.Name, "text", fmt.Sprintf("%s Descriptor Pre-population Engine skipped remote URL fallback", label), "module", name, "error", descErr.Error())
            }
        }

		syntheticProxy := &models.ProxyModule{
			ID:      syntheticID,
			Action:  "enable",
			Metadata: models.ProxyModuleMetadata{
				Name:        name,
				SidecarName: name + "-sc",
				Version:     &gatewayVersion,
			},
		}
		modules.FolioModules = append(modules.FolioModules, syntheticProxy)
		slog.Info(rs.Action.Name, "text", fmt.Sprintf("Injected custom %s proxy module into registry orchestration layer", label), "module", name, "id", syntheticID)
	}
}

func (rs *RegistrySvc) discoverLatestRegistryVersion(moduleName string) (string, error) {
	type RegistryItem struct {
		ID string `json:"id"`
	}

	baseRegistryURL := strings.TrimSpace(rs.Action.ConfigRegistryURL)
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

    override := helpers.ExtractImageOverride(rs.Action.ConfigBackendModules, moduleName)
    repoPath := helpers.ResolveRepoPath(moduleName, override.Image)
    moduleSpecificRegistry := override.Registry

	for i := len(items) - 1; i >= 0; i-- {
		item := items[i]
		if helpers.IsStrictModuleID(item.ID, moduleName) {
			rawVersion := strings.TrimPrefix(item.ID, moduleName+"-")
			cleanTag := strings.Split(rawVersion, "+")[0]

			if moduleSpecificRegistry != "" {
				privateRegistryURL := fmt.Sprintf("https://%s/v2/%s/manifests/%s", moduleSpecificRegistry, repoPath, cleanTag)
				req, err := http.NewRequest(http.MethodHead, privateRegistryURL, nil)
				if err != nil {
					continue
				}
				req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")

				hubResp, hubErr := ctxClient.Do(req)
				if hubErr == nil {
					hubResp.Body.Close()
					if hubResp.StatusCode == http.StatusOK {
						slog.Debug(rs.Action.Name, "text", "Module Override Discovery Match", "module", moduleName, "registry", moduleSpecificRegistry, "path", repoPath, "tag", cleanTag)
						return rawVersion, nil
					}
				}
				continue
			}

			return rawVersion, nil
		}
	}

	return "", fmt.Errorf("no live verified container tags found for module %s", moduleName)
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
			ID:      helpers.ToSyntheticID(component.Name, component.Version),
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

	if modules == nil {
		modules = make([]models.ApplicationModule, 0)
	}

	if err := helpers.WriteJSONToFile(filePath, modules); err != nil {
		return nil, err
	}
	slog.Info(rs.Action.Name, "text", "Persisted module versions to a local file", "file", constant.ModulesFile)

	return modules, nil
}

func isEurekaModule(name string) bool {
	return strings.HasSuffix(name, "-keycloak") ||
		strings.HasPrefix(name, "mgr-") ||
		name == "folio-kong" ||
		name == "folio-module-sidecar" ||
		name == "mod-scheduler"
}
