package modulesvc

import (
	"fmt"
	"time"

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/dockerclient"
	"github.com/folio-org/eureka-cli/errors"
	"github.com/folio-org/eureka-cli/field"
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
			backendModule, ok := containers.BackendModules[module.Name]
			if !ok || !backendModule.DeployModule {
				continue
			}
			if module.Name == moduleName {
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

	return *module.Version
}

func (ms *ModuleSvc) GetSidecarImage(modules []*models.ProxyModule) (string, bool, error) {
	sidecarImageVersion, err := ms.getSidecarImageVersion(modules, ms.Action.ConfigSidecarModule[field.SidecarModuleVersionEntry])
	if err != nil {
		return "", false, err
	}
	if ms.Action.ConfigSidecarModule[field.SidecarModuleLocalImageEntry] != nil && ms.Action.ConfigSidecarModule[field.SidecarModuleLocalImageEntry].(string) != "" {
		return fmt.Sprintf("%s:%s", ms.Action.ConfigSidecarModule[field.SidecarModuleLocalImageEntry].(string), sidecarImageVersion), false, nil
	}
	namespace := ms.RegistrySvc.GetNamespace(sidecarImageVersion)

	return fmt.Sprintf("%s/%s", namespace, fmt.Sprintf("%s:%s", ms.Action.ConfigSidecarModule[field.SidecarModuleImageEntry].(string), sidecarImageVersion)), true, nil
}

func (ms *ModuleSvc) getSidecarImageVersion(modules []*models.ProxyModule, sidecarConfigVersion any) (string, error) {
	ok, sidecarRegistryVersion := ms.findSidecarVersion(modules)
	if !ok && sidecarConfigVersion == nil {
		return "", errors.SidecarVersionNotFound(fmt.Sprintf("%v", sidecarConfigVersion))
	}
	if sidecarConfigVersion != nil {
		return sidecarConfigVersion.(string), nil
	}

	return sidecarRegistryVersion, nil
}

func (ms *ModuleSvc) findSidecarVersion(modules []*models.ProxyModule) (bool, string) {
	for _, module := range modules {
		if module.Name == constant.SidecarProjectName {
			return true, *module.Version
		}
	}

	return false, ""
}

func (ms *ModuleSvc) GetModuleImage(moduleVersion string, module *models.ProxyModule) string {
	return fmt.Sprintf("%s/%s:%s", ms.RegistrySvc.GetNamespace(moduleVersion), module.Name, moduleVersion)
}

func (ms *ModuleSvc) GetModuleEnv(container *models.Containers, module *models.ProxyModule, backendModule models.BackendModule) []string {
	var env []string
	env = append(env, container.GlobalEnv...)
	if backendModule.UseVault {
		env = ms.ModuleEnv.VaultEnv(env, container.VaultRootToken)
	}
	if backendModule.UseOkapiURL {
		env = ms.ModuleEnv.OkapiEnv(env, module.SidecarName, backendModule.PrivatePort)
	}
	if backendModule.DisableSystemUser {
		env = ms.ModuleEnv.DisabledSystemUserEnv(env, module.Name)
	}
	env = ms.ModuleEnv.ModuleEnv(env, backendModule.ModuleEnv)

	return env
}

func (ms *ModuleSvc) GetSidecarEnv(containers *models.Containers, module *models.ProxyModule, backendModule models.BackendModule, moduleURL, sidecarURL string) []string {
	var env []string
	env = append(env, containers.SidecarEnv...)
	env = ms.ModuleEnv.VaultEnv(env, containers.VaultRootToken)
	env = ms.ModuleEnv.KeycloakEnv(env)
	env = ms.ModuleEnv.SidecarEnv(env, module, backendModule.PrivatePort, moduleURL, sidecarURL)

	return env
}
