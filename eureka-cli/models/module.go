package models

import (
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/j011195/eureka-setup/eureka-cli/action"
	"github.com/j011195/eureka-setup/eureka-cli/helpers"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

// ==================== Proxy Module ====================

// ProxyModulesResponse represents the response containing a list of proxy modules from the registry
type ProxyModulesResponse []ProxyModule

// ProxyModule represents a proxy module with ID and metadata
type ProxyModule struct {
	ID       string              `json:"id"`
	Action   string              `json:"action,omitempty"`
	Metadata ProxyModuleMetadata `json:"-"`
}

// ProxyModuleMetadata represents proxy module metadata
type ProxyModuleMetadata struct {
	Name        string
	SidecarName string
	Version     *string
}

// ProxyModulesByRegistry organizes proxy modules by their registry source
type ProxyModulesByRegistry struct {
	FolioModules  []*ProxyModule
	EurekaModules []*ProxyModule
}

// ==================== Backend Module ====================

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

// SidecarRequest contains all the information needed to deploy a sidecar container
type SidecarRequest struct {
	Client           *client.Client
	Containers       *Containers
	Module           *ProxyModule
	BackendModule    BackendModule
	SidecarImage     string
	SidecarResources *container.Resources
}

// NewBackendModuleWithSidecar creates a new BackendModule instance with sidecar configuration
func NewBackendModuleWithSidecar(action *action.Action, p BackendModuleProperties) (*BackendModule, error) {
	exposedPorts := helpers.CreateExposedPorts(*p.PrivatePort)
	moduleServerPort := *p.Port

	var moduleDebugPort, sidecarServerPort, sidecarDebugPort = 0, 0, 0
	if p.DeployModule {
		ports, err := action.GetPreReservedPortSet(3)
		if err != nil {
			return nil, err
		}

		moduleDebugPort = ports[0]
		sidecarServerPort = ports[1]
		sidecarDebugPort = ports[2]
	}

	return &BackendModule{
		DeployModule:             p.DeployModule,
		UseVault:                 p.UseVault,
		UseOkapiURL:              p.UseOkapiURL,
		DisableSystemUser:        p.DisableSystemUser,
		LocalDescriptorPath:      p.LocalDescriptorPath,
		ModuleName:               p.Name,
		ModuleVersion:            p.Version,
		ModuleExposedServerPort:  moduleServerPort,
		ModuleExposedDebugPort:   moduleDebugPort,
		PrivatePort:              *p.PrivatePort,
		ModuleExposedPorts:       exposedPorts,
		ModulePortBindings:       helpers.CreatePortBindings(moduleServerPort, moduleDebugPort, *p.PrivatePort),
		ModuleEnv:                p.Env,
		ModuleResources:          *helpers.CreateResources(true, p.Resources),
		ModuleVolumes:            p.Volumes,
		DeploySidecar:            *p.DeploySidecar,
		SidecarExposedServerPort: sidecarServerPort,
		SidecarExposedDebugPort:  sidecarDebugPort,
		SidecarExposedPorts:      exposedPorts,
		SidecarPortBindings:      helpers.CreatePortBindings(sidecarServerPort, sidecarDebugPort, *p.PrivatePort),
	}, nil
}

// NewBackendModule creates a new BackendModule instance without sidecar configuration
func NewBackendModule(action *action.Action, p BackendModuleProperties) (*BackendModule, error) {
	serverPort := *p.Port
	debugPort, err := action.GetPreReservedPort()
	if err != nil {
		return nil, err
	}

	return &BackendModule{
		DeployModule:            p.DeployModule,
		UseVault:                p.UseVault,
		UseOkapiURL:             p.UseOkapiURL,
		DisableSystemUser:       p.DisableSystemUser,
		LocalDescriptorPath:     p.LocalDescriptorPath,
		ModuleName:              p.Name,
		ModuleVersion:           p.Version,
		ModuleExposedServerPort: serverPort,
		ModuleExposedDebugPort:  debugPort,
		PrivatePort:             *p.PrivatePort,
		ModuleExposedPorts:      helpers.CreateExposedPorts(*p.PrivatePort),
		ModulePortBindings:      helpers.CreatePortBindings(serverPort, debugPort, *p.PrivatePort),
		ModuleEnv:               p.Env,
		ModuleResources:         *helpers.CreateResources(true, p.Resources),
		ModuleVolumes:           p.Volumes,
		DeploySidecar:           false,
		SidecarExposedPorts:     nil,
		SidecarPortBindings:     nil,
	}, nil
}

// ==================== Frontend Module ====================

// FrontendModule represents configuration for a frontend module
type FrontendModule struct {
	DeployModule        bool
	ModuleVersion       *string
	ModuleName          string
	LocalDescriptorPath string
}

// ==================== Container ====================

// Container represents a Docker container configuration with all necessary settings
type Container struct {
	Name          string
	RegistryAuth  string
	Config        *container.Config
	HostConfig    *container.HostConfig
	NetworkConfig *network.NetworkingConfig
	Platform      *v1.Platform
	PullImage     bool
}

// Containers represents a collection of container configurations and their associated metadata
type Containers struct {
	Modules        *ProxyModulesByRegistry
	BackendModules map[string]BackendModule
	IsManagement   bool
}

// ==================== Event ====================

// Event represents a Docker container event with status, error, and progress information
type Event struct {
	Status         string `json:"status"`
	Error          string `json:"error"`
	Progress       string `json:"progress"`
	ProgressDetail struct {
		Current int `json:"current"`
		Total   int `json:"total"`
	} `json:"progressDetail"`
}

// ==================== Registry Extract ====================

// RegistryExtract contains extracted information about modules from registries
type RegistryExtract struct {
	Modules           *ProxyModulesByRegistry
	BackendModules    map[string]BackendModule
	FrontendModules   map[string]FrontendModule
	ModuleDescriptors map[string]any
}

type RegistryRequest struct {
	RegistryName   string
	InstallJsonURL string
	HomeDir        string
	UseRemote      bool
	Metadata       struct {
		FileName string
		Path     string
	}
}
