package moduleparams

import (
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
	ReadBackendModulesFromConfig(managementOnly bool, verbose bool) (map[string]models.BackendModule, error)
	ReadFrontendModulesFromConfig(verbose bool) (map[string]models.FrontendModule, error)
}

// ModuleParams provides functionality for parsing and processing module configuration parameters
type ModuleParams struct {
	Action *action.Action
}

// New creates a new ModuleParams instance
func New(action *action.Action) *ModuleParams {
	return &ModuleParams{Action: action}
}

func (mp *ModuleParams) ReadBackendModulesFromConfig(managementOnly bool, verbose bool) (map[string]models.BackendModule, error) {
	if len(mp.Action.ConfigBackendModules) == 0 {
		slog.Info(mp.Action.Name, "text", "No backend modules were read")
		return make(map[string]models.BackendModule), nil
	}

	backendModules := make(map[string]models.BackendModule)
	for name, value := range mp.Action.ConfigBackendModules {
		if managementOnly && !mp.isManagementModule(name) || !managementOnly && mp.isManagementModule(name) {
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
		moduleServerPort := backendModules[name].ModuleExposedServerPort
		moduleDebugPort := backendModules[name].ModuleExposedDebugPort
		sidecarServerPort := backendModules[name].SidecarExposedServerPort
		sidecarDebugPort := backendModules[name].SidecarExposedDebugPort
		if verbose {
			if properties.Version == nil {
				slog.Info(mp.Action.Name, "text", "Read backend module", "module", name, "port1", moduleServerPort, "port2", moduleDebugPort, "port3", sidecarServerPort, "port4", sidecarDebugPort)
			} else {
				slog.Info(mp.Action.Name, "text", "Read backend module", "module", name, "version", *properties.Version, "port1", moduleServerPort, "port2", moduleDebugPort, "port3", sidecarServerPort, "port4", sidecarDebugPort)
			}
		}
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

func (mp *ModuleParams) createDefaultBackendProperties(name string) (properties models.BackendModuleProperties, err error) {
	properties.DeployModule = true
	if !mp.isManagementModule(name) && !mp.isEdgeModule(name) {
		properties.DeploySidecar = helpers.BoolP(true)
	}

	properties.Port, err = mp.getDefaultPort()
	if err != nil {
		return models.BackendModuleProperties{}, err
	}
	properties.PrivatePort = mp.getDefaultPrivatePort()
	properties.Env = make(map[string]any)
	properties.Resources = make(map[string]any)
	properties.Volumes = []string{}

	return properties, nil
}

func (mp *ModuleParams) isManagementModule(name string) bool {
	return strings.HasPrefix(name, constant.ManagementModulePattern)
}

func (mp *ModuleParams) isEdgeModule(name string) bool {
	return strings.HasPrefix(name, constant.EdgeModulePattern)
}

func (mp *ModuleParams) createConfigurableBackendProperties(value any, name string) (properties models.BackendModuleProperties, err error) {
	entry := value.(map[string]any)
	properties.DeployModule = helpers.GetAnyOrDefault(entry, field.ModuleDeployModuleEntry, true).(bool)
	if !strings.HasPrefix(name, constant.ManagementModulePattern) && !strings.HasPrefix(name, constant.EdgeModulePattern) {
		properties.DeploySidecar = mp.getDeploySidecar(entry)
	}

	properties.UseVault = helpers.GetAnyOrDefault(entry, field.ModuleUseVaultEntry, false).(bool)
	properties.DisableSystemUser = helpers.GetAnyOrDefault(entry, field.ModuleDisableSystemUserEntry, false).(bool)
	properties.UseOkapiURL = helpers.GetAnyOrDefault(entry, field.ModuleUseOkapiURLEntry, false).(bool)
	properties.LocalDescriptorPath = helpers.GetAnyOrDefault(entry, field.ModuleLocalDescriptorPathEntry, "").(string)
	if properties.LocalDescriptorPath != "" {
		if _, err := os.Stat(properties.LocalDescriptorPath); os.IsNotExist(err) {
			return models.BackendModuleProperties{}, errors.LocalDescriptorNotFound(properties.LocalDescriptorPath, name)
		}
	}

	properties.Version = mp.getVersion(entry)
	properties.Port, err = mp.getPort(entry, properties.DeployModule)
	if err != nil {
		return models.BackendModuleProperties{}, err
	}

	properties.PrivatePort = mp.getPrivatePort(entry)
	properties.Env = helpers.GetAnyOrDefault(entry, field.ModuleEnvEntry, make(map[string]any)).(map[string]any)
	properties.Resources = helpers.GetAnyOrDefault(entry, field.ModuleResourceEntry, make(map[string]any)).(map[string]any)
	properties.Volumes, err = mp.getVolumes(entry)
	if err != nil {
		return models.BackendModuleProperties{}, err
	}

	return properties, nil
}

func (mp *ModuleParams) getDeploySidecar(entry map[string]any) *bool {
	if entry[field.ModuleDeploySidecarEntry] == nil {
		return helpers.BoolP(true)
	}

	return helpers.BoolP(entry[field.ModuleDeploySidecarEntry].(bool))
}

func (mp *ModuleParams) getVersion(entry map[string]any) *string {
	if entry[field.ModuleVersionEntry] == nil {
		return nil
	}

	_, ok := entry[field.ModuleVersionEntry].(float64)
	if ok {
		return helpers.StringP(strconv.FormatFloat(entry[field.ModuleVersionEntry].(float64), 'f', -1, 64))
	}

	return helpers.StringP(entry[field.ModuleVersionEntry].(string))
}

func (mp *ModuleParams) getPort(entry map[string]any, deployModule bool) (*int, error) {
	if !deployModule {
		return helpers.IntP(0), nil
	}

	if entry[field.ModulePortEntry] == nil {
		return mp.getDefaultPort()
	}

	return helpers.IntP(entry[field.ModulePortEntry].(int)), nil
}

func (mp *ModuleParams) getDefaultPort() (*int, error) {
	port, err := mp.Action.GetPreReservedPort()
	if err != nil {
		return nil, err
	}

	return helpers.IntP(port), nil
}

func (mp *ModuleParams) getPrivatePort(entry map[string]any) *int {
	if entry[field.ModulePrivatePortEntry] == nil {
		return mp.getDefaultPrivatePort()
	}

	return helpers.IntP(entry[field.ModulePrivatePortEntry].(int))
}

func (mp *ModuleParams) getDefaultPrivatePort() *int {
	defaultServerPort, _ := strconv.Atoi(constant.PrivateServerPort)

	return helpers.IntP(defaultServerPort)
}

func (mp *ModuleParams) getVolumes(entry map[string]any) ([]string, error) {
	if entry[field.ModuleVolumesEntry] == nil {
		return []string{}, nil
	}

	var volumes []string
	for _, value := range entry[field.ModuleVolumesEntry].([]any) {
		var volume = value.(string)
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

func (mp *ModuleParams) ReadFrontendModulesFromConfig(verbose bool) (map[string]models.FrontendModule, error) {
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
				entry := value.(map[string]any)
				if entry[field.ModuleDeployModuleEntry] != nil {
					deployModule = entry[field.ModuleDeployModuleEntry].(bool)
				}
				if entry[field.ModuleVersionEntry] != nil {
					version = helpers.StringP(entry[field.ModuleVersionEntry].(string))
				}

				localDescriptorPath = helpers.GetAnyOrDefault(entry, field.ModuleLocalDescriptorPathEntry, "").(string)
				if localDescriptorPath != "" {
					if _, err := os.Stat(localDescriptorPath); os.IsNotExist(err) {
						return nil, errors.LocalDescriptorNotFound(localDescriptorPath, name)
					}
				}

			}

			frontendModules[name] = *models.NewFrontendModule(deployModule, name, version, localDescriptorPath)
			if verbose {
				if version == nil {
					slog.Info(mp.Action.Name, "text", "Read frontend module", "module", name)
				} else {
					slog.Info(mp.Action.Name, "text", "Read frontend module", "module", name, "version", *version)
				}
			}
		}
	}

	return frontendModules, nil
}
