package models

import (
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/helpers"
)

// BackendModule represents configuration for a backend module and its optional sidecar
type BackendModule struct {
	DeployModule             bool
	UseVault                 bool
	UseOkapiURL              bool
	DisableSystemUser        bool
	LocalDescriptorPath      string
	ModuleName               string
	ModuleVersion            *string
	ModuleExposedServerPort  int
	ModuleExposedDebugPort   int
	ModuleExposedPorts       *nat.PortSet
	ModulePortBindings       *nat.PortMap
	ModuleEnv                map[string]any
	ModuleResources          container.Resources
	ModuleVolumes            []string
	DeploySidecar            bool
	SidecarExposedServerPort int
	SidecarExposedDebugPort  int
	SidecarExposedPorts      *nat.PortSet
	SidecarPortBindings      *nat.PortMap
	PrivatePort              int
}

// BackendModuleProperties contains the properties needed to construct a BackendModule
type BackendModuleProperties struct {
	DeployModule        bool
	DeploySidecar       *bool
	UseVault            bool
	UseOkapiURL         bool
	DisableSystemUser   bool
	LocalDescriptorPath string
	Name                string
	Version             *string
	Port                *int
	PrivatePort         *int
	Env                 map[string]any
	Resources           map[string]any
	Volumes             []string
}

// NewBackendModuleWithSidecar creates a new BackendModule instance with sidecar configuration
func NewBackendModuleWithSidecar(action *action.Action, properties BackendModuleProperties) (*BackendModule, error) {
	exposedPorts := helpers.CreateExposedPorts(*properties.PrivatePort)
	moduleServerPort := *properties.Port

	var (
		moduleDebugPort   = 0
		sidecarServerPort = 0
		sidecarDebugPort  = 0
	)
	if properties.DeployModule {
		ports, err := action.GetPreReserverPortSet([]func() (int, error){
			func() (int, error) { return action.GetPreReservedPort() },
			func() (int, error) { return action.GetPreReservedPort() },
			func() (int, error) { return action.GetPreReservedPort() },
		})
		if err != nil {
			return nil, err
		}

		moduleDebugPort = ports[0]
		sidecarServerPort = ports[1]
		sidecarDebugPort = ports[2]
	}

	return &BackendModule{
		DeployModule:             properties.DeployModule,
		UseVault:                 properties.UseVault,
		UseOkapiURL:              properties.UseOkapiURL,
		DisableSystemUser:        properties.DisableSystemUser,
		LocalDescriptorPath:      properties.LocalDescriptorPath,
		ModuleName:               properties.Name,
		ModuleVersion:            properties.Version,
		ModuleExposedServerPort:  moduleServerPort,
		ModuleExposedDebugPort:   moduleDebugPort,
		PrivatePort:              *properties.PrivatePort,
		ModuleExposedPorts:       exposedPorts,
		ModulePortBindings:       helpers.CreatePortBindings(moduleServerPort, moduleDebugPort, *properties.PrivatePort),
		ModuleEnv:                properties.Env,
		ModuleResources:          *helpers.CreateResources(true, properties.Resources),
		ModuleVolumes:            properties.Volumes,
		DeploySidecar:            *properties.DeploySidecar,
		SidecarExposedServerPort: sidecarServerPort,
		SidecarExposedDebugPort:  sidecarDebugPort,
		SidecarExposedPorts:      exposedPorts,
		SidecarPortBindings:      helpers.CreatePortBindings(sidecarServerPort, sidecarDebugPort, *properties.PrivatePort),
	}, nil
}

// NewBackendModule creates a new BackendModule instance without sidecar configuration
func NewBackendModule(action *action.Action, properties BackendModuleProperties) (*BackendModule, error) {
	moduleServerPort := *properties.Port
	moduleDebugPort, err := action.GetPreReservedPort()
	if err != nil {
		return nil, err
	}

	return &BackendModule{
		DeployModule:            properties.DeployModule,
		UseVault:                properties.UseVault,
		UseOkapiURL:             properties.UseOkapiURL,
		DisableSystemUser:       properties.DisableSystemUser,
		LocalDescriptorPath:     properties.LocalDescriptorPath,
		ModuleName:              properties.Name,
		ModuleVersion:           properties.Version,
		ModuleExposedServerPort: moduleServerPort,
		ModuleExposedDebugPort:  moduleDebugPort,
		PrivatePort:             *properties.PrivatePort,
		ModuleExposedPorts:      helpers.CreateExposedPorts(*properties.PrivatePort),
		ModulePortBindings:      helpers.CreatePortBindings(moduleServerPort, moduleDebugPort, *properties.PrivatePort),
		ModuleEnv:               properties.Env,
		ModuleResources:         *helpers.CreateResources(true, properties.Resources),
		ModuleVolumes:           properties.Volumes,
		DeploySidecar:           false,
		SidecarExposedPorts:     nil,
		SidecarPortBindings:     nil,
	}, nil
}
