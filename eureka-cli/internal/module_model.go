package internal

import (
	"strconv"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

type DeployModuleDto struct {
	Name          string
	Version       string
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
	FileModuleEnv      string
}

type BackendModule struct {
	DeployModule            bool
	ModuleName              string
	ModuleVersion           *string
	ModuleExposedServerPort int
	ModuleExposedPorts      *nat.PortSet
	ModulePortBindings      *nat.PortMap
	ModuleEnvironment       map[string]interface{}
	DeploySidecar           bool
	SidecarExposedPorts     *nat.PortSet
	SidecarPortBindings     *nat.PortMap
	ModuleServerPort        int
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

func NewBackendModuleAndSidecar(deployModule bool, name string, version *string, port int, portServer int, deploySidecar bool, moduleEnvironment map[string]interface{}) *BackendModule {
	exposedPorts := createExposedPorts(portServer)
	modulePortBindings := createPortBindings(port, port+1000, portServer)
	sidecarPortBindings := createPortBindings(port+2000, port+3000, portServer)

	return &BackendModule{
		DeployModule:            deployModule,
		ModuleName:              name,
		ModuleVersion:           version,
		ModuleExposedServerPort: port,
		ModuleServerPort:        portServer,
		ModuleExposedPorts:      exposedPorts,
		ModulePortBindings:      modulePortBindings,
		ModuleEnvironment:       moduleEnvironment,
		DeploySidecar:           deploySidecar,
		SidecarExposedPorts:     exposedPorts,
		SidecarPortBindings:     sidecarPortBindings,
	}
}

func NewBackendModule(name string, version *string, port int, portServer int, moduleEnvironment map[string]interface{}) *BackendModule {
	exposedPorts := createExposedPorts(portServer)
	modulePortBindings := createPortBindings(port, port+1000, portServer)

	return &BackendModule{
		DeployModule:            true,
		ModuleName:              name,
		ModuleVersion:           version,
		ModuleExposedServerPort: port,
		ModuleServerPort:        portServer,
		ModuleExposedPorts:      exposedPorts,
		ModulePortBindings:      modulePortBindings,
		ModuleEnvironment:       moduleEnvironment,
		DeploySidecar:           false,
		SidecarExposedPorts:     nil,
		SidecarPortBindings:     nil,
	}
}

func NewFrontendModule(deployModule bool, name string, version *string) *FrontendModule {
	return &FrontendModule{
		DeployModule:  true,
		ModuleName:    name,
		ModuleVersion: version,
	}
}

func NewDeployManagementModulesDto(vaultRootToken string,
	registryHostnames map[string]string,
	registryModules map[string][]*RegistryModule,
	backendModulesMap map[string]BackendModule,
	globalEnvironment []string) *DeployModulesDto {
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

func NewDeployModulesDto(vaultRootToken string,
	registryHostnames map[string]string,
	registryModules map[string][]*RegistryModule,
	backendModulesMap map[string]BackendModule,
	globalEnvironment []string,
	sidecarEnvironment []string) *DeployModulesDto {
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

func NewDeployModuleDto(name string,
	version string,
	image string,
	env []string,
	backendModule BackendModule,
	networkConfig *network.NetworkingConfig,
	authToken string) *DeployModuleDto {
	return &DeployModuleDto{
		Name:         name,
		Version:      version,
		Image:        image,
		RegistryAuth: authToken,
		Config: &container.Config{
			Image:        image,
			Hostname:     name,
			Env:          env,
			ExposedPorts: *backendModule.ModuleExposedPorts,
		},
		HostConfig: &container.HostConfig{
			PortBindings:  *backendModule.ModulePortBindings,
			RestartPolicy: container.RestartPolicy{Name: container.RestartPolicyOnFailure, MaximumRetryCount: 3},
		},
		NetworkConfig: networkConfig,
		Platform:      &v1.Platform{},
		PullImage:     true,
	}
}

func NewDeploySidecarDto(name string,
	version string,
	image string,
	env []string,
	backendModule BackendModule,
	networkConfig *network.NetworkingConfig,
	pullSidecarImage bool,
	authToken string) *DeployModuleDto {
	return &DeployModuleDto{
		Name:         name,
		Version:      version,
		Image:        image,
		RegistryAuth: authToken,
		Config: &container.Config{
			Image:        image,
			Hostname:     name,
			Env:          env,
			ExposedPorts: *backendModule.SidecarExposedPorts,
		},
		HostConfig: &container.HostConfig{
			PortBindings:  *backendModule.SidecarPortBindings,
			RestartPolicy: container.RestartPolicy{Name: container.RestartPolicyOnFailure, MaximumRetryCount: 3},
		},
		NetworkConfig: networkConfig,
		Platform:      &v1.Platform{},
		PullImage:     pullSidecarImage,
	}
}

func NewModuleNetworkConfig() *network.NetworkingConfig {
	return &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{NetworkName: {NetworkID: NetworkId}},
	}
}

func createExposedPorts(serverPort int) *nat.PortSet {
	moduleExposedPorts := make(map[nat.Port]struct{})

	moduleExposedPorts[nat.Port(strconv.Itoa(serverPort))] = struct{}{}
	moduleExposedPorts[nat.Port(DefaultDebugPort)] = struct{}{}

	portSet := nat.PortSet(moduleExposedPorts)

	return &portSet
}

func createPortBindings(hostServerPort int, hostServerDebugPort int, serverPort int) *nat.PortMap {
	var (
		serverPortBinding      []nat.PortBinding
		serverDebugPortBinding []nat.PortBinding
	)

	serverPortBinding = append(serverPortBinding, nat.PortBinding{HostIP: HostIp, HostPort: strconv.Itoa(hostServerPort)})
	serverDebugPortBinding = append(serverDebugPortBinding, nat.PortBinding{HostIP: HostIp, HostPort: strconv.Itoa(hostServerDebugPort)})

	portBindings := make(map[nat.Port][]nat.PortBinding)
	portBindings[nat.Port(strconv.Itoa(serverPort))] = serverPortBinding
	portBindings[nat.Port(DefaultDebugPort)] = serverDebugPortBinding

	portMap := nat.PortMap(portBindings)

	return &portMap
}
