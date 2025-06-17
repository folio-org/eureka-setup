package internal

import (
	"strconv"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

const (
	DefaultModuleCpus              int64 = 1
	DefaultModuleMemoryReservation int64 = 128
	DefaultModuleMemory            int64 = 750
	DefaultModuleSwap              int64 = -1

	DefaultSidecarCpus              int64 = 1
	DefaultSidecarMemoryReservation int64 = 64
	DefaultSidecarMemory            int64 = 450
	DefaultSidecarSwap              int64 = -1
)

type DeployModuleDto struct {
	Name          string
	Image         string
	RegistryAuth  string
	Config        *container.Config
	HostConfig    *container.HostConfig
	NetworkConfig *network.NetworkingConfig
	Platform      *v1.Platform
	PullImage     bool
}

type DeployModulesDto struct {
	VaultRootToken     string
	RegistryHostname   map[string]string
	RegistryModules    map[string][]*RegistryModule
	BackendModulesMap  map[string]BackendModule
	GlobalEnvironment  []string
	SidecarEnvironment []string
	ManagementOnly     bool
}

type BackendModuleDto struct {
	deployModule      bool
	deploySidecar     *bool
	useVault          bool
	useOkapiUrl       bool
	disableSystemUser bool
	name              string
	version           *string
	port              *int
	portServer        *int
	environment       map[string]any
	resources         map[string]any
	volumes           []string
}

type BackendModule struct {
	DeployModule             bool
	UseVault                 bool
	UseOkapiUrl              bool
	DisableSystemUser        bool
	ModuleName               string
	ModuleVersion            *string
	ModuleExposedServerPort  int
	ModuleExposedDebugPort   int
	ModuleExposedPorts       *nat.PortSet
	ModulePortBindings       *nat.PortMap
	ModuleEnvironment        map[string]any
	ModuleResources          container.Resources
	ModuleVolumes            []string
	DeploySidecar            bool
	SidecarExposedServerPort int
	SidecarExposedDebugPort  int
	SidecarExposedPorts      *nat.PortSet
	SidecarPortBindings      *nat.PortMap
	ModuleServerPort         int
}

type FrontendModule struct {
	DeployModule  bool
	ModuleVersion *string
	ModuleName    string
}

type Event struct {
	Status         string `json:"status"`
	Error          string `json:"error"`
	Progress       string `json:"progress"`
	ProgressDetail struct {
		Current int `json:"current"`
		Total   int `json:"total"`
	} `json:"progressDetail"`
}

func NewBackendModuleWithSidecar(commandName string, dto BackendModuleDto) *BackendModule {
	exposedPorts := createExposedPorts(*dto.portServer)

	moduleServerPort := *dto.port

	var moduleDebugPort, sidecarServerPort, sidecarDebugPort int = 0, 0, 0
	if dto.deployModule {
		moduleDebugPort = GetAndSetFreePortFromRange(commandName, PortStartIndex, PortEndIndex, &ReservedPorts)
		sidecarServerPort = GetAndSetFreePortFromRange(commandName, PortStartIndex, PortEndIndex, &ReservedPorts)
		sidecarDebugPort = GetAndSetFreePortFromRange(commandName, PortStartIndex, PortEndIndex, &ReservedPorts)
	}

	return &BackendModule{
		DeployModule:             dto.deployModule,
		UseVault:                 dto.useVault,
		UseOkapiUrl:              dto.useOkapiUrl,
		DisableSystemUser:        dto.disableSystemUser,
		ModuleName:               dto.name,
		ModuleVersion:            dto.version,
		ModuleExposedServerPort:  moduleServerPort,
		ModuleExposedDebugPort:   moduleDebugPort,
		ModuleServerPort:         *dto.portServer,
		ModuleExposedPorts:       exposedPorts,
		ModulePortBindings:       CreatePortBindings(moduleServerPort, moduleDebugPort, *dto.portServer),
		ModuleEnvironment:        dto.environment,
		ModuleResources:          *CreateResources(true, dto.resources),
		ModuleVolumes:            dto.volumes,
		DeploySidecar:            *dto.deploySidecar,
		SidecarExposedServerPort: sidecarServerPort,
		SidecarExposedDebugPort:  sidecarDebugPort,
		SidecarExposedPorts:      exposedPorts,
		SidecarPortBindings:      CreatePortBindings(sidecarServerPort, sidecarDebugPort, *dto.portServer),
	}
}

func NewBackendModule(commandName string, dto BackendModuleDto) *BackendModule {
	moduleServerPort := *dto.port
	moduleDebugPort := GetAndSetFreePortFromRange(commandName, PortStartIndex, PortEndIndex, &ReservedPorts)

	return &BackendModule{
		DeployModule:            dto.deployModule,
		UseVault:                dto.useVault,
		UseOkapiUrl:             dto.useOkapiUrl,
		DisableSystemUser:       dto.disableSystemUser,
		ModuleName:              dto.name,
		ModuleVersion:           dto.version,
		ModuleExposedServerPort: moduleServerPort,
		ModuleExposedDebugPort:  moduleDebugPort,
		ModuleServerPort:        *dto.portServer,
		ModuleExposedPorts:      createExposedPorts(*dto.portServer),
		ModulePortBindings:      CreatePortBindings(moduleServerPort, moduleDebugPort, *dto.portServer),
		ModuleEnvironment:       dto.environment,
		ModuleResources:         *CreateResources(true, dto.resources),
		ModuleVolumes:           dto.volumes,
		DeploySidecar:           false,
		SidecarExposedPorts:     nil,
		SidecarPortBindings:     nil,
	}
}

func createExposedPorts(serverPort int) *nat.PortSet {
	moduleExposedPorts := make(map[nat.Port]struct{})

	moduleExposedPorts[nat.Port(strconv.Itoa(serverPort))] = struct{}{}
	moduleExposedPorts[nat.Port(DefaultDebugPort)] = struct{}{}

	portSet := nat.PortSet(moduleExposedPorts)

	return &portSet
}

func CreatePortBindings(hostServerPort int, hostServerDebugPort int, serverPort int) *nat.PortMap {
	var (
		serverPortBinding      []nat.PortBinding
		serverDebugPortBinding []nat.PortBinding
	)

	serverPortBinding = append(serverPortBinding, nat.PortBinding{HostIP: DefaultHostIp, HostPort: strconv.Itoa(hostServerPort)})
	serverDebugPortBinding = append(serverDebugPortBinding, nat.PortBinding{HostIP: DefaultHostIp, HostPort: strconv.Itoa(hostServerDebugPort)})

	portBindings := make(map[nat.Port][]nat.PortBinding)
	portBindings[nat.Port(strconv.Itoa(serverPort))] = serverPortBinding
	portBindings[nat.Port(DefaultDebugPort)] = serverDebugPortBinding

	portMap := nat.PortMap(portBindings)

	return &portMap
}

func CreateResources(isModule bool, resources map[string]any) *container.Resources {
	if len(resources) == 0 {
		return createDefaultResources(isModule)
	}

	oomKillDisable := getBoolValueOrDefault(ModuleResourceOomKillDisableEntryKey, resources, false)

	return &container.Resources{
		CPUCount:          getIntValueOrDefault(ModuleResourceCpuCountEntryKey, resources, DefaultModuleCpus),
		MemoryReservation: convertMiBToBytes(getIntValueOrDefault(ModuleResourceMemoryReservationEntryKey, resources, DefaultModuleMemoryReservation)),
		Memory:            convertMiBToBytes(getIntValueOrDefault(ModuleResourceMemoryEntryKey, resources, DefaultModuleMemory)),
		MemorySwap:        convertMiBToBytes(getIntValueOrDefault(ModuleResourceMemorySwapEntryKey, resources, DefaultModuleSwap)),
		OomKillDisable:    &oomKillDisable,
	}
}

func createDefaultResources(isModule bool) *container.Resources {
	oomKillDisable := false

	if isModule {
		return &container.Resources{
			CPUCount:          DefaultModuleCpus,
			MemoryReservation: convertMiBToBytes(DefaultModuleMemoryReservation),
			Memory:            convertMiBToBytes(DefaultModuleMemory),
			MemorySwap:        convertMiBToBytes(DefaultModuleSwap),
			OomKillDisable:    &oomKillDisable,
		}
	}

	return &container.Resources{
		CPUCount:          DefaultSidecarCpus,
		MemoryReservation: convertMiBToBytes(DefaultSidecarMemoryReservation),
		Memory:            convertMiBToBytes(DefaultSidecarMemory),
		MemorySwap:        convertMiBToBytes(DefaultSidecarSwap),
		OomKillDisable:    &oomKillDisable,
	}
}

func convertMiBToBytes(mib int64) int64 {
	if mib < 0 {
		return mib
	}

	return mib * 1024 * 1024
}

func getIntValueOrDefault(key string, resources map[string]any, defaultValue int64) int64 {
	value, ok := resources[key].(int)
	if !ok || resources[key] == nil {
		return int64(defaultValue)
	}

	return int64(value)
}

func getBoolValueOrDefault(key string, resources map[string]any, defaultValue bool) bool {
	value, ok := resources[key].(bool)
	if !ok || resources[key] == nil {
		return defaultValue
	}

	return value
}

func NewFrontendModule(deployModule bool, name string, version *string) *FrontendModule {
	return &FrontendModule{
		DeployModule:  true,
		ModuleName:    name,
		ModuleVersion: version,
	}
}

func NewDeployManagementModulesDto(vaultRootToken string, registryHostnames map[string]string, registryModules map[string][]*RegistryModule,
	backendModulesMap map[string]BackendModule, globalEnvironment []string) *DeployModulesDto {
	return &DeployModulesDto{
		VaultRootToken:     vaultRootToken,
		RegistryHostname:   registryHostnames,
		RegistryModules:    registryModules,
		BackendModulesMap:  backendModulesMap,
		GlobalEnvironment:  globalEnvironment,
		SidecarEnvironment: nil,
		ManagementOnly:     true,
	}
}

func NewDeployModulesDto(vaultRootToken string, registryHostnames map[string]string, registryModules map[string][]*RegistryModule,
	backendModulesMap map[string]BackendModule, globalEnvironment []string, sidecarEnvironment []string) *DeployModulesDto {
	return &DeployModulesDto{
		VaultRootToken:     vaultRootToken,
		RegistryHostname:   registryHostnames,
		RegistryModules:    registryModules,
		BackendModulesMap:  backendModulesMap,
		GlobalEnvironment:  globalEnvironment,
		SidecarEnvironment: sidecarEnvironment,
		ManagementOnly:     false,
	}
}

func NewDeployModuleDto(name string, image string, env []string, backendModule BackendModule,
	networkConfig *network.NetworkingConfig) *DeployModuleDto {
	return &DeployModuleDto{
		Name:  name,
		Image: image,
		Config: &container.Config{
			Image:        image,
			Hostname:     name,
			Env:          env,
			ExposedPorts: *backendModule.ModuleExposedPorts,
		},
		HostConfig: &container.HostConfig{
			PortBindings:  *backendModule.ModulePortBindings,
			RestartPolicy: container.RestartPolicy{Name: container.RestartPolicyAlways},
			Resources:     backendModule.ModuleResources,
			Binds:         backendModule.ModuleVolumes,
		},
		NetworkConfig: networkConfig,
		Platform:      &v1.Platform{},
		PullImage:     true,
	}
}

func NewDeploySidecarDto(name string, image string, env []string, backendModule BackendModule,
	networkConfig *network.NetworkingConfig, resources *container.Resources) *DeployModuleDto {
	return &DeployModuleDto{
		Name:   name,
		Image:  image,
		Config: &container.Config{Image: image, Hostname: name, Env: env, ExposedPorts: *backendModule.SidecarExposedPorts},
		HostConfig: &container.HostConfig{
			PortBindings:  *backendModule.SidecarPortBindings,
			RestartPolicy: container.RestartPolicy{Name: container.RestartPolicyAlways},
			Resources:     *resources,
		},
		NetworkConfig: networkConfig,
		Platform:      &v1.Platform{},
		PullImage:     false,
	}
}

func NewModuleNetworkConfig() *network.NetworkingConfig {
	return &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			DefaultNetworkId: {
				NetworkID: DefaultNetworkId,
				Aliases:   []string{DefaultNetworkAlias},
			},
		},
	}
}
