package moduleparams

import (
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/errors"
	"github.com/folio-org/eureka-cli/field"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/folio-org/eureka-cli/models"
)

// ModuleParamsProcessor defines the interface for reading module parameters from configuration
type ModuleParamsProcessor interface {
	ReadBackendModulesFromConfig(managementOnly bool) (map[string]models.BackendModule, error)
	ReadFrontendModulesFromConfig() (map[string]models.FrontendModule, error)
}

// ModuleParams provides functionality for parsing and processing module configuration parameters
type ModuleParams struct {
	Action *action.Action
}

// New creates a new ModuleParams instance
func New(action *action.Action) *ModuleParams {
	return &ModuleParams{Action: action}
}

func (mp *ModuleParams) ReadBackendModulesFromConfig(managementOnly bool) (map[string]models.BackendModule, error) {
	if len(mp.Action.ConfigBackendModules) == 0 {
		slog.Info(mp.Action.Name, "text", "No backend modules were read")
		return make(map[string]models.BackendModule), nil
	}

	backendModules := make(map[string]models.BackendModule)
	for name, value := range mp.Action.ConfigBackendModules {
		if managementOnly && !mp.IsManagementModule(name) || !managementOnly && mp.IsManagementModule(name) {
			continue
		}

		properties, err := mp.createBackendProperties(name, value)
		if err != nil {
			return nil, err
		}

		backendModule, err := mp.createBackendModule(properties)
		if err != nil {
			return nil, err
		}

		backendModules[name] = *backendModule
		moduleInfo := name
		if properties.Version != nil {
			moduleInfo = fmt.Sprintf("%s with fixed version %s", name, *properties.Version)
		}

		moduleServerPort := backendModules[name].ModuleExposedServerPort
		moduleDebugPort := backendModules[name].ModuleExposedDebugPort
		sidecarServerPort := backendModules[name].SidecarExposedServerPort
		sidecarDebugPort := backendModules[name].SidecarExposedDebugPort
		slog.Info(mp.Action.Name, "text", "Read backend module with port reserve", "module", moduleInfo, "port1", moduleServerPort, "port2", moduleDebugPort, "port3", sidecarServerPort, "port4", sidecarDebugPort)
	}

	return backendModules, nil
}

func (mp *ModuleParams) createBackendProperties(name string, value any) (models.BackendModuleProperties, error) {
	if value == nil {
		return mp.createDefaultBackendProperties(name)
	}

	return mp.createConfigurableBackendProperties(value, name)
}

func (mp *ModuleParams) createBackendModule(properties models.BackendModuleProperties) (*models.BackendModule, error) {
	if properties.DeploySidecar != nil && *properties.DeploySidecar {
		return models.NewBackendModuleWithSidecar(mp.Action, properties)
	}

	return models.NewBackendModule(mp.Action, properties)
}

func (mp *ModuleParams) IsManagementModule(name string) bool {
	return strings.HasPrefix(name, constant.ManagementModulePattern)
}

func (mp *ModuleParams) IsEdgeModule(name string) bool {
	return strings.HasPrefix(name, constant.EdgeModulePattern)
}

func (mp *ModuleParams) createDefaultBackendProperties(name string) (properties models.BackendModuleProperties, err error) {
	properties.DeployModule = true
	if !mp.IsManagementModule(name) && !mp.IsEdgeModule(name) {
		properties.DeploySidecar = helpers.BoolP(true)
	}

	properties.Port, err = mp.getDefaultPort()
	if err != nil {
		return models.BackendModuleProperties{}, err
	}
	properties.PortServer = mp.getDefaultPortServer()
	properties.Env = make(map[string]any)
	properties.Resources = make(map[string]any)
	properties.Volumes = []string{}

	return properties, nil
}

func (mp *ModuleParams) createConfigurableBackendProperties(value any, name string) (properties models.BackendModuleProperties, err error) {
	mapEntry := value.(map[string]any)
	properties.DeployModule = helpers.GetAnyOrDefault(mapEntry, field.ModuleDeployModuleEntry, true).(bool)
	if !strings.HasPrefix(name, constant.ManagementModulePattern) && !strings.HasPrefix(name, constant.EdgeModulePattern) {
		properties.DeploySidecar = mp.getDeploySidecar(mapEntry)
	}

	properties.UseVault = helpers.GetAnyOrDefault(mapEntry, field.ModuleUseVaultEntry, false).(bool)
	properties.DisableSystemUser = helpers.GetAnyOrDefault(mapEntry, field.ModuleDisableSystemUserEntry, false).(bool)
	properties.UseOkapiURL = helpers.GetAnyOrDefault(mapEntry, field.ModuleUseOkapiURLEntry, false).(bool)
	properties.LocalDescriptorPath = helpers.GetAnyOrDefault(mapEntry, field.ModuleLocalDescriptorPathEntry, "").(string)
	if properties.LocalDescriptorPath != "" {
		if _, err := os.Stat(properties.LocalDescriptorPath); os.IsNotExist(err) {
			return models.BackendModuleProperties{}, errors.LocalDescriptorNotFound(properties.LocalDescriptorPath, name)
		}
	}

	properties.Version = mp.getVersion(mapEntry)
	properties.Port, err = mp.getPort(properties.DeployModule, mapEntry)
	if err != nil {
		return models.BackendModuleProperties{}, err
	}

	properties.PortServer = mp.getPortServer(mapEntry)
	properties.Env = helpers.GetAnyOrDefault(mapEntry, field.ModuleEnvEntry, make(map[string]any)).(map[string]any)
	properties.Resources = helpers.GetAnyOrDefault(mapEntry, field.ModuleResourceEntry, make(map[string]any)).(map[string]any)
	properties.Volumes, err = mp.getVolumes(mapEntry)
	if err != nil {
		return models.BackendModuleProperties{}, err
	}

	return properties, nil
}

func (mp *ModuleParams) getDeploySidecar(mapEntry map[string]any) *bool {
	if mapEntry[field.ModuleDeploySidecarEntry] == nil {
		return helpers.BoolP(true)
	}

	return helpers.BoolP(mapEntry[field.ModuleDeploySidecarEntry].(bool))
}

func (mp *ModuleParams) getVersion(mapEntry map[string]any) *string {
	if mapEntry[field.ModuleVersionEntry] == nil {
		return nil
	}

	_, ok := mapEntry[field.ModuleVersionEntry].(float64)
	if ok {
		return helpers.StringP(strconv.FormatFloat(mapEntry[field.ModuleVersionEntry].(float64), 'f', -1, 64))
	}

	return helpers.StringP(mapEntry[field.ModuleVersionEntry].(string))
}

func (mp *ModuleParams) getPort(deployModule bool, mapEntry map[string]any) (*int, error) {
	if !deployModule {
		return helpers.IntP(0), nil
	}

	if mapEntry[field.ModulePortEntry] == nil {
		return mp.getDefaultPort()
	}

	return helpers.IntP(mapEntry[field.ModulePortEntry].(int)), nil
}

func (mp *ModuleParams) getDefaultPort() (*int, error) {
	port, err := mp.Action.GetPreReservedPort()
	if err != nil {
		return nil, err
	}

	return helpers.IntP(port), nil
}

func (mp *ModuleParams) getPortServer(mapEntry map[string]any) *int {
	if mapEntry[field.ModulePortServerEntry] == nil {
		return mp.getDefaultPortServer()
	}

	return helpers.IntP(mapEntry[field.ModulePortServerEntry].(int))
}

func (mp *ModuleParams) getDefaultPortServer() *int {
	defaultServerPort, _ := strconv.Atoi(constant.ServerPort)

	return helpers.IntP(defaultServerPort)
}

func (mp *ModuleParams) getVolumes(mapEntry map[string]any) ([]string, error) {
	if mapEntry[field.ModuleVolumesEntry] == nil {
		return []string{}, nil
	}

	var volumes []string
	for _, value := range mapEntry[field.ModuleVolumesEntry].([]any) {
		var volume = value.(string)
		// Windows paths are hard to set, this is why we have this to automatically resolve path location
		if runtime.GOOS == "windows" && strings.Contains(volume, "$EUREKA") {
			homeConfigDir, err := helpers.GetHomeDirPath()
			if err != nil {
				return nil, err
			}

			volume = strings.ReplaceAll(volume, "$EUREKA", homeConfigDir)
		}

		if _, err := os.Stat(volume); os.IsNotExist(err) {
			if err != nil {
				return nil, err
			}
		}
		volumes = append(volumes, volume)
	}

	return volumes, nil
}

func (mp *ModuleParams) ReadFrontendModulesFromConfig() (map[string]models.FrontendModule, error) {
	frontendModules := make(map[string]models.FrontendModule)
	configModules := []map[string]any{mp.Action.ConfigFrontendModules, mp.Action.ConfigCustomFrontendModules}
	if len(configModules) == 0 {
		slog.Info(mp.Action.Name, "text", "No frontend modules were read")
		return frontendModules, nil
	}

	for _, modules := range configModules {
		for name, value := range modules {
			var (
				deployModule        = true
				version             *string
				localDescriptorPath = ""
			)

			if value != nil {
				mapEntry := value.(map[string]any)
				if mapEntry[field.ModuleDeployModuleEntry] != nil {
					deployModule = mapEntry[field.ModuleDeployModuleEntry].(bool)
				}
				if mapEntry[field.ModuleVersionEntry] != nil {
					version = helpers.StringP(mapEntry[field.ModuleVersionEntry].(string))
				}

				localDescriptorPath = helpers.GetAnyOrDefault(mapEntry, field.ModuleLocalDescriptorPathEntry, "").(string)
				if localDescriptorPath != "" {
					if _, err := os.Stat(localDescriptorPath); os.IsNotExist(err) {
						return nil, errors.LocalDescriptorNotFound(localDescriptorPath, name)
					}
				}

			}

			frontendModules[name] = *models.NewFrontendModule(deployModule, name, version, localDescriptorPath)
			moduleInfo := name
			if version != nil {
				moduleInfo = fmt.Sprintf("name %s with version %s", name, *version)
			}

			slog.Info(mp.Action.Name, "text", "Read frontend module", "module", moduleInfo)
		}
	}

	return frontendModules, nil
}
