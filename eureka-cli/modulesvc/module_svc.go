package modulesvc

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/folio-org/eureka-setup/eureka-cli/action"
	"github.com/folio-org/eureka-setup/eureka-cli/constant"
	"github.com/folio-org/eureka-setup/eureka-cli/dockerclient"
	"github.com/folio-org/eureka-setup/eureka-cli/errors"
	"github.com/folio-org/eureka-setup/eureka-cli/field"
	"github.com/folio-org/eureka-setup/eureka-cli/helpers"
	"github.com/folio-org/eureka-setup/eureka-cli/httpclient"
	"github.com/folio-org/eureka-setup/eureka-cli/models"
	"github.com/folio-org/eureka-setup/eureka-cli/moduleenv"
	"github.com/folio-org/eureka-setup/eureka-cli/registrysvc"
)

// ModuleProcessor defines the composite interface for all module-related operations
type ModuleProcessor interface {
	ModuleVaultHandler
	ModuleReadinessChecker
	ModuleProvisioner
	ModuleManager
	ModuleCustomizer
}

// ModuleProvisioner defines the interface for module provisioning operations
type ModuleProvisioner interface {
	GetBackendModule(containers *models.Containers, moduleName string) (*models.BackendModule, *models.ProxyModule)
	GetModuleImageVersion(backendModule models.BackendModule, module *models.ProxyModule) string
	GetSidecarImage(modules []*models.ProxyModule) (string, bool, error)
	GetModuleImage(module *models.ProxyModule) string
	GetLocalModuleImage(namespace, moduleName, moduleVersion string) string
	GetModuleEnv(container *models.Containers, module *models.ProxyModule, backendModule models.BackendModule) []string
	GetSidecarEnv(containers *models.Containers, module *models.ProxyModule, backendModule models.BackendModule, moduleURL, sidecarURL string) []string
}

// ModuleSvc provides comprehensive functionality for managing backend modules
type ModuleSvc struct {
	Action              *action.Action
	HTTPClient          httpclient.HTTPClientRunner
	DockerClient        dockerclient.DockerClientRunner
	RegistrySvc         registrysvc.RegistryProcessor
	ModuleEnv           moduleenv.ModuleEnvProcessor
	ReadinessMaxRetries int
	ReadinessWait       time.Duration
}

func New(action *action.Action,
	httpClient httpclient.HTTPClientRunner,
	dockerClient dockerclient.DockerClientRunner,
	registrySvc registrysvc.RegistryProcessor,
	moduleEnv moduleenv.ModuleEnvProcessor) *ModuleSvc {
	return &ModuleSvc{Action: action, HTTPClient: httpClient, DockerClient: dockerClient, RegistrySvc: registrySvc, ModuleEnv: moduleEnv}
}

func (ms *ModuleSvc) GetBackendModule(containers *models.Containers, moduleName string) (*models.BackendModule, *models.ProxyModule) {
	slog.Debug(ms.Action.Name, "text", "Resolving backend module from application profile tracking context", "moduleName", moduleName)

	// 1. Instantly resolve the backend module profile directly using the CLI requested moduleName
	backendModule, exists := containers.BackendModules[moduleName]
	// Exit early if the module doesn't exist OR is explicitly disabled via configuration rules
	if !exists || !backendModule.DeployModule {
		slog.Debug(ms.Action.Name, "text", "Requested module name missing or explicitly disabled via deployModule profile flag", "moduleName", moduleName)
		return nil, nil
	}

	// 2. Scan the tracking lists using the robust IsStrictModuleID prefix helper
	allModules := [][]*models.ProxyModule{containers.Modules.FolioModules, containers.Modules.EurekaModules}
	for _, modules := range allModules {
		for _, module := range modules {
			if module != nil && helpers.IsStrictModuleID(module.ID, moduleName) {
				slog.Debug(ms.Action.Name, "text", "Matched proxy module via strict ID prefix matching engine", "id", module.ID, "moduleName", moduleName)
				return &backendModule, module
			}
		}
	}

	// 3. Resilient fallback: Check metadata name property just in case it was populated in-memory
	for _, modules := range allModules {
		for _, module := range modules {
			if module != nil && module.Metadata.Name == moduleName {
				slog.Debug(ms.Action.Name, "text", "Matched proxy module via structural metadata name string fallback", "id", module.ID, "moduleName", moduleName)
				return &backendModule, module
			}
		}
	}

	// 4. Centralised out-of-tree fallback
	// If it's an external app, it won't be in the registry tracking arrays.
	// We synthesize the tracking metadata on the fly
	slog.Debug(ms.Action.Name, "text", "Module missing from tracking arrays. Synthesizing proxy record fallback context", "moduleName", moduleName)

	fallbackID := moduleName
	var fallbackVersion *string

	if ms.Action.Param != nil && ms.Action.Param.ID != "" {
		fallbackID = ms.Action.Param.ID
		vStr := helpers.GetModuleVersionFromID(ms.Action.Param.ID)
		fallbackVersion = &vStr
	}

	synthesizedModule := models.NewSynthesizedProxyModule(fallbackID, moduleName, fallbackVersion)

	return &backendModule, synthesizedModule
}

func (ms *ModuleSvc) GetModuleImageVersion(backendModule models.BackendModule, module *models.ProxyModule) string {
	if backendModule.ModuleVersion != nil {
		slog.Debug(ms.Action.Name, "text", "Image resolution engine prioritizing explicit profile version override", "version", *backendModule.ModuleVersion)
		return *backendModule.ModuleVersion
	}

	fallbackVersion := ""
	if module != nil && module.Metadata.Version != nil {
		fallbackVersion = *module.Metadata.Version
	}
	slog.Debug(ms.Action.Name, "text", "Image resolution engine defaulting to catalog proxy version index", "version", fallbackVersion)
	return fallbackVersion
}

func (ms *ModuleSvc) GetSidecarImage(modules []*models.ProxyModule) (string, bool, error) {
	configSidecarVersion := ms.Action.ConfigSidecarModule[field.SidecarModuleVersionEntry]
	sidecarImageVersion, err := ms.getSidecarImageVersion(modules, configSidecarVersion)
	if err != nil {
		return "", false, err
	}

	image := helpers.GetString(ms.Action.ConfigSidecarModule, field.SidecarModuleImageEntry)
	if image == "" {
		return "", false, errors.SidecarImageBlank()
	}
	finalImage := fmt.Sprintf("%s:%s", image, sidecarImageVersion)

	customNamespace := helpers.GetBool(ms.Action.ConfigSidecarModule, field.SidecarModuleCustomNamespaceEntry)
	if customNamespace {
		slog.Debug(ms.Action.Name, "text", "Utilizing custom local fallback namespace allocation mapping rule", "image", finalImage)
		return finalImage, true, nil
	}
	namespace := ms.RegistrySvc.GetNamespace(sidecarImageVersion)

	computedTarget := fmt.Sprintf("%s/%s", namespace, finalImage)
	slog.Debug(ms.Action.Name, "text", "Resolved final full sidecar tracking location path reference", "target", computedTarget)
	return computedTarget, true, nil
}

func (ms *ModuleSvc) getSidecarImageVersion(modules []*models.ProxyModule, rawConfigSidecarVersion any) (string, error) {
	if rawConfigSidecarVersion != nil {
		configSidecarVersion, ok := rawConfigSidecarVersion.(string)
		if ok {
			slog.Debug(ms.Action.Name, "text", "Sidecar version discovery selected explicit profile configuration tag", "version", configSidecarVersion)
			return configSidecarVersion, nil
		}
	}

	registrySidecarVersion, exists := ms.findRegistrySidecarImageVersion(modules)
	if !exists || registrySidecarVersion == "" {
		return "", errors.SidecarVersionNotFound()
	}

	slog.Debug(ms.Action.Name, "text", "Sidecar version discovery auto-resolved tag directly from registry manifest records", "version", registrySidecarVersion)
	return registrySidecarVersion, nil
}

func (ms *ModuleSvc) findRegistrySidecarImageVersion(modules []*models.ProxyModule) (string, bool) {
	for _, module := range modules {
		if module.Metadata.Name == constant.SidecarProjectName && module.Metadata.Version != nil {
			return *module.Metadata.Version, true
		}
	}
	return "", false
}

func (ms *ModuleSvc) GetModuleImage(module *models.ProxyModule) string {
	moduleVersion := ""
	if module.Metadata.Version != nil {
		moduleVersion = *module.Metadata.Version
	}
	namespace := ms.RegistrySvc.GetNamespace(moduleVersion)
	computedImage := fmt.Sprintf("%s/%s:%s", namespace, module.Metadata.Name, moduleVersion)

	slog.Debug(ms.Action.Name, "text", "Constructed standard default framework repository mapping descriptor", "image", computedImage)
	return computedImage
}

func (ms *ModuleSvc) GetLocalModuleImage(namespace, moduleName, moduleVersion string) string {
	return fmt.Sprintf("%s/%s:%s", namespace, moduleName, moduleVersion)
}

func (ms *ModuleSvc) GetModuleEnv(container *models.Containers, module *models.ProxyModule, backendModule models.BackendModule) []string {
	slog.Debug(ms.Action.Name, "text", "Compiling isolated environment injection vectors for backend task container",
		"module", module.Metadata.Name,
		"useVault", backendModule.UseVault,
		"useOkapiURL", backendModule.UseOkapiURL,
		"disableSystemUser", backendModule.DisableSystemUser,
	)

	env := ms.Action.GetConfigEnvVars(field.Env)
	if backendModule.UseVault {
		env = ms.ModuleEnv.VaultEnv(env, ms.Action.VaultRootToken)
	}
	if backendModule.UseOkapiURL {
		env = ms.ModuleEnv.OkapiEnv(env, module.Metadata.SidecarName, backendModule.PrivatePort)
	}
	if backendModule.DisableSystemUser {
		env = ms.ModuleEnv.DisabledSystemUserEnv(env, module.Metadata.Name)
	}
	env = append(env, ms.Action.GetTemplateEnvVars(field.TemplateEnv, module.Metadata.Name)...)
	env = ms.ModuleEnv.ModuleEnv(env, backendModule.ModuleEnv)

	slog.Debug(ms.Action.Name, "text", "Backend task container environment assembly complete", "module", module.Metadata.Name, "totalEnvCount", len(env))
	return env
}

func (ms *ModuleSvc) GetSidecarEnv(containers *models.Containers, module *models.ProxyModule, backendModule models.BackendModule, moduleURL, sidecarURL string) []string {
	slog.Debug(ms.Action.Name, "text", "Compiling routing environment parameters for sidecar proxy boundary container",
		"sidecar", module.Metadata.SidecarName,
		"moduleUrl", moduleURL,
		"sidecarUrl", sidecarURL,
	)

	env := ms.Action.GetConfigEnvVars(field.SidecarModuleEnv)
	env = ms.ModuleEnv.VaultEnv(env, ms.Action.VaultRootToken)
	env = ms.ModuleEnv.KeycloakEnv(env)
	env = ms.ModuleEnv.SidecarEnv(env, module, backendModule.PrivatePort, moduleURL, sidecarURL)
	env = ms.ModuleEnv.ModuleEnv(env, backendModule.SidecarEnv)

	slog.Debug(ms.Action.Name, "text", "Sidecar proxy container environment assembly complete", "sidecar", module.Metadata.SidecarName, "totalEnvCount", len(env))
	return env
}
