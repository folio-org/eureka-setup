package modulesvc

import (
	"fmt"
	"time"

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/dockerclient"
	"github.com/folio-org/eureka-cli/errors"
	"github.com/folio-org/eureka-cli/field"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/folio-org/eureka-cli/httpclient"
	"github.com/folio-org/eureka-cli/models"
	"github.com/folio-org/eureka-cli/moduleenv"
	"github.com/folio-org/eureka-cli/registrysvc"
)

// ModuleProcessor defines the composite interface for all module-related operations
type ModuleProcessor interface {
	ModuleVaultHandler
	ModuleReadinessChecker
	ModuleProvisioner
	ModuleManager
}

// ModuleProvisioner defines the interface for module provisioning operations
type ModuleProvisioner interface {
	GetBackendModule(containers *models.Containers, moduleName string) (*models.BackendModule, *models.ProxyModule)
	GetModuleImageVersion(backendModule models.BackendModule, module *models.ProxyModule) string
	GetSidecarImage(modules []*models.ProxyModule) (string, bool, error)
	GetModuleImage(moduleVersion string, module *models.ProxyModule) string
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
	allModules := [][]*models.ProxyModule{containers.Modules.FolioModules, containers.Modules.EurekaModules}
	for _, modules := range allModules {
		for _, module := range modules {
			backendModule, exists := containers.BackendModules[module.Metadata.Name]
			if !exists || !backendModule.DeployModule {
				continue
			}
			if module.Metadata.Name == moduleName {
				return &backendModule, module
			}
		}
	}

	return nil, nil
}

func (ms *ModuleSvc) GetModuleImageVersion(backendModule models.BackendModule, module *models.ProxyModule) string {
	if backendModule.ModuleVersion != nil {
		return *backendModule.ModuleVersion
	}

	return *module.Metadata.Version
}

func (ms *ModuleSvc) GetSidecarImage(modules []*models.ProxyModule) (string, bool, error) {
	configSidecarVersion := ms.Action.ConfigSidecarModule[field.SidecarModuleVersionEntry]
	sidecarImageVersion, err := ms.getSidecarImageVersion(modules, configSidecarVersion)
	if err != nil {
		return "", false, err
	}

	image := helpers.GetString(ms.Action.ConfigSidecarModule, field.SidecarModuleImageEntry)
	if image == "" {
		return "", false, errors.New("sidecar image is blank")
	}
	finalImage := fmt.Sprintf("%s:%s", image, sidecarImageVersion)

	customNamespace := helpers.GetBool(ms.Action.ConfigSidecarModule, field.SidecarModuleCustomNamespaceEntry)
	if customNamespace {
		return finalImage, true, nil
	}
	namespace := ms.RegistrySvc.GetNamespace(sidecarImageVersion)

	return fmt.Sprintf("%s/%s", namespace, finalImage), true, nil
}

func (ms *ModuleSvc) getSidecarImageVersion(modules []*models.ProxyModule, rawConfigSidecarVersion any) (string, error) {
	if rawConfigSidecarVersion != nil {
		configSidecarVersion, ok := rawConfigSidecarVersion.(string)
		if ok {
			return configSidecarVersion, nil
		}
	}

	registrySidecarVersion, exists := ms.findRegistrySidecarImageVersion(modules)
	if !exists || registrySidecarVersion == "" {
		return "", errors.SidecarVersionNotFound()
	}

	return registrySidecarVersion, nil
}

func (ms *ModuleSvc) findRegistrySidecarImageVersion(modules []*models.ProxyModule) (string, bool) {
	for _, module := range modules {
		if module.Metadata.Name == constant.SidecarProjectName {
			return *module.Metadata.Version, true
		}
	}

	return "", false
}

func (ms *ModuleSvc) GetModuleImage(moduleVersion string, module *models.ProxyModule) string {
	return fmt.Sprintf("%s/%s:%s", ms.RegistrySvc.GetNamespace(moduleVersion), module.Metadata.Name, moduleVersion)
}

func (ms *ModuleSvc) GetModuleEnv(container *models.Containers, module *models.ProxyModule, backendModule models.BackendModule) []string {
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
	env = ms.ModuleEnv.ModuleEnv(env, backendModule.ModuleEnv)

	return env
}

func (ms *ModuleSvc) GetSidecarEnv(containers *models.Containers, module *models.ProxyModule, backendModule models.BackendModule, moduleURL, sidecarURL string) []string {
	env := ms.Action.GetConfigEnvVars(field.SidecarModuleEnv)
	env = ms.ModuleEnv.VaultEnv(env, ms.Action.VaultRootToken)
	env = ms.ModuleEnv.KeycloakEnv(env)
	env = ms.ModuleEnv.SidecarEnv(env, module, backendModule.PrivatePort, moduleURL, sidecarURL)

	return env
}
