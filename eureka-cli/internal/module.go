package internal

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/viper"
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

func NewBackendModuleAndSidecar(deployModule bool, name string, version *string, port int, portInternal int, deploySidecar bool, moduleEnvironment map[string]interface{}) *BackendModule {
	log.Println("name:", name, "port:", port, "portInternal:", portInternal)

	exposedPorts := CreateExposedPorts(portInternal)
	modulePortBindings := CreatePortBindings(port, port+1000, portInternal)
	sidecarPortBindings := CreatePortBindings(port+2000, port+3000, portInternal)

	log.Println("modulePortBindings:", modulePortBindings)

	return &BackendModule{
		DeployModule:            deployModule,
		ModuleName:              name,
		ModuleVersion:           version,
		ModuleExposedServerPort: port,
		ModuleExposedPorts:      exposedPorts,
		ModulePortBindings:      modulePortBindings,
		ModuleEnvironment:       moduleEnvironment,
		DeploySidecar:           deploySidecar,
		SidecarExposedPorts:     exposedPorts,
		SidecarPortBindings:     sidecarPortBindings,
	}
}

func NewBackendModule(name string, port int, portInternal int, moduleEnvironment map[string]interface{}) *BackendModule {
	exposedPorts := CreateExposedPorts(portInternal)
	modulePortBindings := CreatePortBindings(port, port+1000, portInternal)

	return &BackendModule{
		DeployModule:            true,
		ModuleName:              name,
		ModuleVersion:           nil,
		ModuleExposedServerPort: port,
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
	registryHostname map[string]string,
	registryModules map[string][]*RegistryModule,
	backendModulesMap map[string]BackendModule,
	globalEnvironment []string) *DeployModulesDto {
	return &DeployModulesDto{
		VaultRootToken:     vaultRootToken,
		RegistryHostname:   registryHostname,
		RegistryModules:    registryModules,
		BackendModulesMap:  backendModulesMap,
		GlobalEnvironment:  globalEnvironment,
		SidecarEnvironment: nil,
		ManagementOnly:     true,
	}
}

func NewDeployModulesDto(vaultRootToken string,
	registryHostname map[string]string,
	registryModules map[string][]*RegistryModule,
	backendModulesMap map[string]BackendModule,
	globalEnvironment []string,
	sidecarEnvironment []string) *DeployModulesDto {
	return &DeployModulesDto{
		VaultRootToken:     vaultRootToken,
		RegistryHostname:   registryHostname,
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

func CreateExposedPorts(internalPort int) *nat.PortSet {
	moduleExposedPorts := make(map[nat.Port]struct{})

	//moduleExposedPorts[nat.Port(ServerPort)] = struct{}{}
	moduleExposedPorts[nat.Port(strconv.Itoa(internalPort))] = struct{}{}
	moduleExposedPorts[nat.Port(DebugPort)] = struct{}{}

	portSet := nat.PortSet(moduleExposedPorts)

	return &portSet
}

func CreatePortBindings(hostServerPort int, hostServerDebugPort int, internalPort int) *nat.PortMap {
	var (
		serverPortBinding      []nat.PortBinding
		serverDebugPortBinding []nat.PortBinding
	)

	serverPortBinding = append(serverPortBinding, nat.PortBinding{HostIP: HostIp, HostPort: strconv.Itoa(hostServerPort)})
	serverDebugPortBinding = append(serverDebugPortBinding, nat.PortBinding{HostIP: HostIp, HostPort: strconv.Itoa(hostServerDebugPort)})

	portBindings := make(map[nat.Port][]nat.PortBinding)
	portBindings[nat.Port(strconv.Itoa(internalPort))] = serverPortBinding
	//portBindings[nat.Port(ServerPort)] = serverPortBinding
	//portBindings[nat.Port(strconv.Itoa(8082))] = serverPortBinding
	portBindings[nat.Port(DebugPort)] = serverDebugPortBinding

	portMap := nat.PortMap(portBindings)

	return &portMap
}

func CreateClient(commandName string) *client.Client {
	newClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		slog.Error(commandName, "client.NewClientWithOpts error", "")
		panic(err)
	}

	return newClient
}

func GetRootVaultToken(commandName string, client *client.Client) string {
	os.Setenv("DOCKER_HOST", "unix:///System/Volumes/Data/Users/sellis/.rd/docker.sock")

	logStream, err := client.ContainerLogs(context.Background(), "vault", container.LogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		slog.Error(commandName, "cli.ContainerLogs error", "")
		panic(err)
	}
	defer logStream.Close()

	buffer := make([]byte, 8)
	for {
		_, err := logStream.Read(buffer)
		if err != nil {
			slog.Error(commandName, "logStream.Read(buffer) error", "")
			panic(err)
		}

		count := binary.BigEndian.Uint32(buffer[4:])
		rawLogLine := make([]byte, count)

		_, err = logStream.Read(rawLogLine)
		if err != nil {
			slog.Warn(commandName, "logStream.Read(rawLogLine) encoutered an EOF", "")
		}

		parsedLogLine := string(rawLogLine)

		if strings.Contains(parsedLogLine, "init.sh: Root VAULT TOKEN is:") {
			vaultRootToken := strings.TrimSpace(VaultRootTokenRegexp.ReplaceAllString(parsedLogLine, `$1`))

			slog.Info(commandName, "Found vault root token", vaultRootToken)

			return vaultRootToken
		}
	}
}

func GetDeployedModules(commandName string, client *client.Client, filters filters.Args) []types.Container {
	deployedModules, err := client.ContainerList(context.Background(), container.ListOptions{All: true, Filters: filters})
	if err != nil {
		slog.Error(commandName, "cli.ContainerList error", "")
		panic(err)
	}

	return deployedModules
}

func UndeployModule(commandName string, client *client.Client, deployedModule types.Container) {
	err := client.ContainerStop(context.Background(), deployedModule.ID, container.StopOptions{Signal: "9"})
	if err != nil {
		slog.Error(commandName, "cli.ContainerStop error", "")
		panic(err)
	}

	err = client.ContainerRemove(context.Background(), deployedModule.ID, container.RemoveOptions{Force: true, RemoveVolumes: true})
	if err != nil {
		slog.Error(commandName, "cli.ContainerRemove error", "")
		panic(err)
	}

	slog.Info(commandName, "Undeployed module container", fmt.Sprintf("%s %s %s", deployedModule.ID, strings.ReplaceAll(deployedModule.Names[0], "/", ""), deployedModule.Status))
}

func DeployModule(commandName string, client *client.Client, dto *DeployModuleDto) {
	containerName := fmt.Sprintf("eureka-%s", dto.Name)

	if dto.PullImage {
		reader, err := client.ImagePull(context.Background(), dto.Image, types.ImagePullOptions{RegistryAuth: dto.RegistryAuth})
		if err != nil {
			slog.Error(commandName, "cli.ImagePull error", "")
			panic(err)
		}
		defer reader.Close()

		decoder := json.NewDecoder(reader)

		var event *Event
		for {
			if err := decoder.Decode(&event); err != nil {
				if err == io.EOF {
					break
				}

				panic(err)
			}

			if event.Error != "" {
				slog.Error(commandName, "Pulling module container image", fmt.Sprintf("%+v", event.Error))
			}
		}
	}

	cr, err := client.ContainerCreate(context.Background(), dto.Config, dto.HostConfig, dto.NetworkConfig, dto.Platform, containerName)
	if err != nil {
		slog.Error(commandName, "cli.ContainerCreate error", "")
		panic(err)
	}

	if cr.Warnings != nil && len(cr.Warnings) > 0 {
		slog.Warn(commandName, fmt.Sprintf("cli.ContainerCreate warnings, '%s'", cr.Warnings), "")
	}

	err = client.ContainerStart(context.Background(), cr.ID, container.StartOptions{})
	if err != nil {
		slog.Error(commandName, "cli.ContainerStart error", "")
		panic(err)
	}

	slog.Info(commandName, "Deployed module container", fmt.Sprintf("%s %s", cr.ID, containerName))
}

func DeployModules(commandName string, client *client.Client, dto *DeployModulesDto) map[string]int {
	deployedModules := make(map[string]int)

	sidecarModule := viper.GetStringMap(SidecarModule)

	// TODO Add automatic sidecar version determination
	var sidecarImage string
	if sidecarImage == "" {
		sidecarId := fmt.Sprintf("%s:%s", sidecarModule["image"], sidecarModule["version"])
		sidecarImage = fmt.Sprintf("%s/%s", DetermineImageRegistryNamespace(sidecarModule["version"].(string)), sidecarId)
	}

	resourceUrlVault := viper.GetString(ResourcesVaultKey)
	networkConfig := NewModuleNetworkConfig()
	pullSidecarImage := true

	for registryName, registryModules := range dto.RegistryModules {
		slog.Info(commandName, fmt.Sprintf("Deploying %s modules", registryName), "")

		for _, module := range registryModules {
			managementModule := strings.Contains(module.Name, ManagementModulePattern)
			if dto.ManagementOnly && !managementModule || !dto.ManagementOnly && managementModule {
				continue
			}

			backendModule, ok := dto.BackendModulesMap[module.Name]
			if !ok {
				continue
			}

			if !backendModule.DeployModule {
				slog.Info(commandName, fmt.Sprintf("Ignoring %s module deployment", module.Name), "")
				continue
			}

			image := fmt.Sprintf("%s/%s:%s", DetermineImageRegistryNamespace(*module.Version), module.Name, *module.Version)

			var combinedModuleEnvironment []string
			combinedModuleEnvironment = append(combinedModuleEnvironment, dto.GlobalEnvironment...)
			combinedModuleEnvironment = AppendModuleEnvironment(backendModule.ModuleEnvironment, combinedModuleEnvironment)
			combinedModuleEnvironment = AppendVaultEnvironment(combinedModuleEnvironment, dto.VaultRootToken, resourceUrlVault)

			// TODO Make calling this configurable such that if the env var isn't present pass an empty authToken string.
			authToken := GetEurekaRegistryAuthToken(commandName)

			slog.Info(commandName, fmt.Sprint("Deploying image: ", image, "Auth Token: ", len(authToken)), "")

			deployModuleDto := NewDeployModuleDto(module.Name, *module.Version, image, combinedModuleEnvironment, backendModule, networkConfig, authToken)

			DeployModule(commandName, client, deployModuleDto)

			deployedModules[module.Name] = backendModule.ModuleExposedServerPort

			if !backendModule.DeploySidecar {
				slog.Info(commandName, fmt.Sprintf("Ignoring %s sidecar deployment", module.Name), "")
				continue
			}

			var combinedSidecarEnvironment []string
			combinedSidecarEnvironment = append(combinedSidecarEnvironment, dto.SidecarEnvironment...)
			combinedSidecarEnvironment = AppendKeycloakEnvironment(commandName, combinedSidecarEnvironment)
			combinedSidecarEnvironment = AppendVaultEnvironment(combinedSidecarEnvironment, dto.VaultRootToken, resourceUrlVault)
			combinedSidecarEnvironment = AppendManagementEnvironment(combinedSidecarEnvironment)
			combinedSidecarEnvironment = AppendSidecarEnvironment(combinedSidecarEnvironment, module)

			// TODO Need to pass in auth token here too.
			deploySidecarDto := NewDeploySidecarDto(module.SidecarName, *module.Version, sidecarImage, combinedSidecarEnvironment, backendModule, networkConfig, pullSidecarImage, authToken)

			DeployModule(commandName, client, deploySidecarDto)

			pullSidecarImage = false
		}
	}

	return deployedModules
}

// TODO fix this
func DetermineImageRegistryNamespace(version string) string {
	var registryNamespace string
	//if strings.Contains(version, "SNAPSHOT") {
	//	registryNamespace = SnapshotRegistry
	//} else {
	//	registryNamespace = ReleaseRegistry
	//}

	registryNamespace = os.Getenv("AWS_ECR_FOLIO_REPO")

	return registryNamespace
}
