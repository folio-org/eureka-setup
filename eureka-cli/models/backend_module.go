package models

import (
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/helpers"
)

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
	ModuleServerPort         int
}

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
	PortServer          *int
	Env                 map[string]any
	Resources           map[string]any
	Volumes             []string
}

func NewBackendModuleWithSidecar(action *action.Action, properties BackendModuleProperties) (*BackendModule, error) {
	exposedPorts := helpers.CreateExposedPorts(*properties.PortServer)

	moduleServerPort := *properties.Port

	var moduleDebugPort, sidecarServerPort, sidecarDebugPort = 0, 0, 0
	var err error
	if properties.DeployModule {
		moduleDebugPort, err = helpers.SetFreePortFromRange(action)
		if err != nil {
			return nil, err
		}

		sidecarServerPort, err = helpers.SetFreePortFromRange(action)
		if err != nil {
			return nil, err
		}

		sidecarDebugPort, err = helpers.SetFreePortFromRange(action)
		if err != nil {
			return nil, err
		}
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
		ModuleServerPort:         *properties.PortServer,
		ModuleExposedPorts:       exposedPorts,
		ModulePortBindings:       helpers.CreatePortBindings(moduleServerPort, moduleDebugPort, *properties.PortServer),
		ModuleEnv:                properties.Env,
		ModuleResources:          *helpers.CreateResources(true, properties.Resources),
		ModuleVolumes:            properties.Volumes,
		DeploySidecar:            *properties.DeploySidecar,
		SidecarExposedServerPort: sidecarServerPort,
		SidecarExposedDebugPort:  sidecarDebugPort,
		SidecarExposedPorts:      exposedPorts,
		SidecarPortBindings:      helpers.CreatePortBindings(sidecarServerPort, sidecarDebugPort, *properties.PortServer),
	}, nil
}

func NewBackendModule(action *action.Action, properties BackendModuleProperties) (*BackendModule, error) {
	moduleServerPort := *properties.Port
	moduleDebugPort, err := helpers.SetFreePortFromRange(action)
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
		ModuleServerPort:        *properties.PortServer,
		ModuleExposedPorts:      helpers.CreateExposedPorts(*properties.PortServer),
		ModulePortBindings:      helpers.CreatePortBindings(moduleServerPort, moduleDebugPort, *properties.PortServer),
		ModuleEnv:               properties.Env,
		ModuleResources:         *helpers.CreateResources(true, properties.Resources),
		ModuleVolumes:           properties.Volumes,
		DeploySidecar:           false,
		SidecarExposedPorts:     nil,
		SidecarPortBindings:     nil,
	}, nil
}
