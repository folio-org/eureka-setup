package modulesvc

import (
	"fmt"

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

// ModuleSvc provides comprehensive functionality for managing backend modules
type ModuleSvc struct {
	Action       *action.Action
	HTTPClient   httpclient.HTTPClientRunner
	DockerClient dockerclient.DockerClientRunner
	RegistrySvc  registrysvc.RegistryProcessor
	ModuleEnv    moduleenv.ModuleEnvProcessor
}

// ModuleProvisioner defines the interface for module provisioning operations
type ModuleProvisioner interface {
	GetBackendModule(containers *models.Containers, moduleName string) (*models.BackendModule, *models.RegistryModule)
	GetModuleImageVersion(backendModule models.BackendModule, registryModule *models.RegistryModule) string
	GetSidecarImage(registryModules []*models.RegistryModule) (string, bool, error)
	GetModuleImage(moduleVersion string, registryModule *models.RegistryModule) string
	GetModuleEnv(container *models.Containers, module *models.RegistryModule, backendModule models.BackendModule) []string
	GetSidecarEnv(containers *models.Containers, module *models.RegistryModule, backendModule models.BackendModule, moduleURL, sidecarURL string) []string
}

func New(action *action.Action,
	httpClient httpclient.HTTPClientRunner,
	dockerClient dockerclient.DockerClientRunner,
	registrySvc registrysvc.RegistryProcessor,
	moduleEnv moduleenv.ModuleEnvProcessor) *ModuleSvc {
	return &ModuleSvc{Action: action, HTTPClient: httpClient, DockerClient: dockerClient, RegistrySvc: registrySvc, ModuleEnv: moduleEnv}
}

func (ms *ModuleSvc) GetBackendModule(containers *models.Containers, moduleName string) (*models.BackendModule, *models.RegistryModule) {
	for _, registryModules := range containers.RegistryModules {
		for _, registryModule := range registryModules {
			backendModule, ok := containers.BackendModules[registryModule.Name]
			if !ok || !backendModule.DeployModule {
				continue
			}

			if registryModule.Name == moduleName {
				return &backendModule, registryModule
			}
		}
	}

	return nil, nil
}

func (ms *ModuleSvc) GetModuleImageVersion(backendModule models.BackendModule, registryModule *models.RegistryModule) string {
	if backendModule.ModuleVersion != nil {
		return *backendModule.ModuleVersion
	}

	return *registryModule.Version
}

func (ms *ModuleSvc) GetSidecarImage(registryModules []*models.RegistryModule) (string, bool, error) {
	sidecarImageVersion, err := ms.getSidecarImageVersion(registryModules, ms.Action.ConfigSidecarModule[field.SidecarModuleVersionEntry])
	if err != nil {
		return "", false, err
	}

	if ms.Action.ConfigSidecarModule[field.SidecarModuleLocalImageEntry] != nil && ms.Action.ConfigSidecarModule[field.SidecarModuleLocalImageEntry].(string) != "" {
		return fmt.Sprintf("%s:%s", ms.Action.ConfigSidecarModule[field.SidecarModuleLocalImageEntry].(string), sidecarImageVersion), false, nil
	}
	namespace := ms.RegistrySvc.GetNamespace(sidecarImageVersion)

	return fmt.Sprintf("%s/%s", namespace, fmt.Sprintf("%s:%s", ms.Action.ConfigSidecarModule[field.SidecarModuleImageEntry].(string), sidecarImageVersion)), true, nil
}

func (ms *ModuleSvc) getSidecarImageVersion(registryModules []*models.RegistryModule, sidecarConfigVersion any) (string, error) {
	ok, sidecarRegistryVersion := ms.findSidecarRegistryVersion(registryModules)
	if !ok && sidecarConfigVersion == nil {
		return "", errors.SidecarVersionNotFound(fmt.Sprintf("%v", sidecarConfigVersion))
	}
	if sidecarConfigVersion != nil {
		return sidecarConfigVersion.(string), nil
	}

	return sidecarRegistryVersion, nil
}

func (ms *ModuleSvc) findSidecarRegistryVersion(registryModules []*models.RegistryModule) (bool, string) {
	for _, registryModule := range registryModules {
		if registryModule.Name == constant.SidecarProjectName {
			return true, *registryModule.Version
		}
	}

	return false, ""
}

func (ms *ModuleSvc) GetModuleImage(moduleVersion string, registryModule *models.RegistryModule) string {
	return fmt.Sprintf("%s/%s:%s", ms.RegistrySvc.GetNamespace(moduleVersion), registryModule.Name, moduleVersion)
}

func (ms *ModuleSvc) GetModuleEnv(container *models.Containers, module *models.RegistryModule, backendModule models.BackendModule) []string {
	var env []string
	env = append(env, container.GlobalEnv...)
	if backendModule.UseVault {
		env = ms.ModuleEnv.VaultEnv(env, container.VaultRootToken)
	}
	if backendModule.UseOkapiURL {
		env = ms.ModuleEnv.OkapiEnv(env, module.SidecarName, backendModule.ModuleServerPort)
	}
	if backendModule.DisableSystemUser {
		env = ms.ModuleEnv.DisabledSystemUserEnv(env, module.Name)
	}
	env = ms.ModuleEnv.ModuleEnv(env, backendModule.ModuleEnv)

	return env
}

func (ms *ModuleSvc) GetSidecarEnv(containers *models.Containers, module *models.RegistryModule, backendModule models.BackendModule, moduleURL, sidecarURL string) []string {
	var env []string
	env = append(env, containers.SidecarEnv...)
	env = ms.ModuleEnv.VaultEnv(env, containers.VaultRootToken)
	env = ms.ModuleEnv.KeycloakEnv(env)
	env = ms.ModuleEnv.SidecarEnv(env, module, backendModule.ModuleServerPort, moduleURL, sidecarURL)

	return env
}
