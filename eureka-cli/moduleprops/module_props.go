package moduleprops

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

// ModulePropsProcessor defines the interface for reading module parameters from configuration
type ModulePropsProcessor interface {
	ReadBackendModules(isManagement bool, verbose bool) (map[string]models.BackendModule, error)
	ReadFrontendModules(verbose bool) (map[string]models.FrontendModule, error)
}

// ModuleProps provides functionality for parsing and processing module configuration parameters
type ModuleProps struct {
	Action *action.Action
}

// New creates a new ModuleProps instance
func New(action *action.Action) *ModuleProps {
	return &ModuleProps{Action: action}
}

func (mp *ModuleProps) ReadBackendModules(isManagement bool, verbose bool) (map[string]models.BackendModule, error) {
	modules := make(map[string]models.BackendModule)
	if len(mp.Action.ConfigBackendModules) == 0 {
		slog.Info(mp.Action.Name, "text", "No backend modules were read")
		return modules, nil
	}

	for name, value := range mp.Action.ConfigBackendModules {
		if isManagement && !mp.isManagementModule(name) || !isManagement && mp.isManagementModule(name) {
			continue
		}

		p, err := mp.createBackendProperties(name, value)
		if err != nil {
			return nil, err
		}

		backendModule, err := mp.createBackendModule(p)
		if err != nil {
			return nil, err
		}
		modules[name] = *backendModule
		if verbose {
			if p.Version == nil {
				slog.Info(mp.Action.Name, "text", "Read backend module", "module", name,
					"port1", modules[name].ModuleExposedServerPort,
					"port2", modules[name].ModuleExposedDebugPort,
					"port3", modules[name].SidecarExposedServerPort,
					"port4", modules[name].SidecarExposedDebugPort)
			} else {
				slog.Info(mp.Action.Name, "text", "Read backend module", "module", name, "version", *p.Version,
					"port1", modules[name].ModuleExposedServerPort,
					"port2", modules[name].ModuleExposedDebugPort,
					"port3", modules[name].SidecarExposedServerPort,
					"port4", modules[name].SidecarExposedDebugPort)
			}
		}
	}

	return modules, nil
}

func (mp *ModuleProps) createBackendProperties(name string, value any) (models.BackendModuleProperties, error) {
	if value == nil {
		return mp.createDefaultBackendProperties(name)
	} else {
		return mp.createConfigurableBackendProperties(value, name)
	}
}

func (mp *ModuleProps) createBackendModule(properties models.BackendModuleProperties) (*models.BackendModule, error) {
	if properties.DeploySidecar != nil && *properties.DeploySidecar {
		return models.NewBackendModuleWithSidecar(mp.Action, properties)
	}

	return models.NewBackendModule(mp.Action, properties)
}

func (mp *ModuleProps) createDefaultBackendProperties(name string) (p models.BackendModuleProperties, err error) {
	p.DeployModule = true
	if !mp.isManagementModule(name) && !mp.isEdgeModule(name) {
		p.DeploySidecar = helpers.BoolP(true)
	}

	p.Port, err = mp.getDefaultPort()
	if err != nil {
		return models.BackendModuleProperties{}, err
	}
	p.PrivatePort = mp.getDefaultPrivatePort()
	p.Env = make(map[string]any)
	p.Resources = make(map[string]any)
	p.Volumes = []string{}

	return p, nil
}

func (mp *ModuleProps) isManagementModule(name string) bool {
	return strings.HasPrefix(name, constant.ManagementModulePattern)
}

func (mp *ModuleProps) isEdgeModule(name string) bool {
	return strings.HasPrefix(name, constant.EdgeModulePattern)
}

func (mp *ModuleProps) createConfigurableBackendProperties(value any, name string) (p models.BackendModuleProperties, err error) {
	entry := value.(map[string]any)
	p.DeployModule = helpers.GetAnyOrDefault(entry, field.ModuleDeployModuleEntry, true).(bool)
	if !strings.HasPrefix(name, constant.ManagementModulePattern) && !strings.HasPrefix(name, constant.EdgeModulePattern) {
		p.DeploySidecar = mp.getDeploySidecar(entry)
	}

	p.UseVault = helpers.GetAnyOrDefault(entry, field.ModuleUseVaultEntry, false).(bool)
	p.DisableSystemUser = helpers.GetAnyOrDefault(entry, field.ModuleDisableSystemUserEntry, false).(bool)
	p.UseOkapiURL = helpers.GetAnyOrDefault(entry, field.ModuleUseOkapiURLEntry, false).(bool)
	p.LocalDescriptorPath = helpers.GetAnyOrDefault(entry, field.ModuleLocalDescriptorPathEntry, "").(string)
	if p.LocalDescriptorPath != "" {
		if _, err := os.Stat(p.LocalDescriptorPath); os.IsNotExist(err) {
			return models.BackendModuleProperties{}, errors.LocalDescriptorNotFound(p.LocalDescriptorPath, name)
		}
	}

	p.Version = mp.getVersion(entry)
	p.Port, err = mp.getPort(entry, p.DeployModule)
	if err != nil {
		return models.BackendModuleProperties{}, err
	}

	p.PrivatePort = mp.getPrivatePort(entry)
	p.Env = helpers.GetAnyOrDefault(entry, field.ModuleEnvEntry, make(map[string]any)).(map[string]any)
	p.Resources = helpers.GetAnyOrDefault(entry, field.ModuleResourceEntry, make(map[string]any)).(map[string]any)
	p.Volumes, err = mp.getVolumes(entry)
	if err != nil {
		return models.BackendModuleProperties{}, err
	}

	return p, nil
}

func (mp *ModuleProps) getDeploySidecar(entry map[string]any) *bool {
	if entry[field.ModuleDeploySidecarEntry] == nil {
		return helpers.BoolP(true)
	}

	return helpers.BoolP(entry[field.ModuleDeploySidecarEntry].(bool))
}

func (mp *ModuleProps) getVersion(entry map[string]any) *string {
	rawVersion, exists := entry[field.ModuleVersionEntry]
	if !exists {
		return nil
	}

	floatValue, ok := rawVersion.(float64)
	if ok {
		value := strconv.FormatFloat(floatValue, 'f', -1, 64)
		return helpers.StringP(value)
	}

	value, ok := rawVersion.(string)
	if ok {
		return helpers.StringP(value)
	}

	return nil
}

func (mp *ModuleProps) getPort(entry map[string]any, deployModule bool) (*int, error) {
	if !deployModule {
		return helpers.IntP(0), nil
	}
	if entry[field.ModulePortEntry] == nil {
		return mp.getDefaultPort()
	}

	return helpers.IntP(entry[field.ModulePortEntry].(int)), nil
}

func (mp *ModuleProps) getDefaultPort() (*int, error) {
	port, err := mp.Action.GetPreReservedPort()
	if err != nil {
		return nil, err
	}

	return helpers.IntP(port), nil
}

func (mp *ModuleProps) getPrivatePort(entry map[string]any) *int {
	if entry[field.ModulePrivatePortEntry] == nil {
		return mp.getDefaultPrivatePort()
	}

	return helpers.IntP(entry[field.ModulePrivatePortEntry].(int))
}

func (mp *ModuleProps) getDefaultPrivatePort() *int {
	defaultServerPort, _ := strconv.Atoi(constant.PrivateServerPort)

	return helpers.IntP(defaultServerPort)
}

func (mp *ModuleProps) getVolumes(entry map[string]any) ([]string, error) {
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

func (mp *ModuleProps) ReadFrontendModules(verbose bool) (map[string]models.FrontendModule, error) {
	modules := make(map[string]models.FrontendModule)
	combinedConfigModules := []map[string]any{mp.Action.ConfigFrontendModules, mp.Action.ConfigCustomFrontendModules}
	if len(combinedConfigModules) == 0 {
		slog.Info(mp.Action.Name, "text", "No frontend modules were read")
		return modules, nil
	}

	for _, configModules := range combinedConfigModules {
		for name, value := range configModules {
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

			modules[name] = models.FrontendModule{
				DeployModule:        deployModule,
				ModuleName:          name,
				ModuleVersion:       version,
				LocalDescriptorPath: localDescriptorPath,
			}
			if verbose {
				if version == nil {
					slog.Info(mp.Action.Name, "text", "Read frontend module", "module", name)
				} else {
					slog.Info(mp.Action.Name, "text", "Read frontend module", "module", name, "version", *version)
				}
			}
		}
	}

	return modules, nil
}
