package internal

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
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

type RegisterModuleDto struct {
	RegistryUrls         map[string]string
	RegistryModules      map[string][]RegistryModule
	BackendModulesMap    map[string]BackendModule
	FrontendModulesMap   map[string]FrontendModule
	ModuleDescriptorsMap map[string]interface{}
	FileModuleEnvPointer *os.File
	EnableDebug          bool
}

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
	VaultToken        string
	RegistryHostname  map[string]string
	RegistryModules   map[string][]RegistryModule
	BackendModulesMap map[string]BackendModule
	GlobalEnvironment []string
	ManagementOnly    bool
}

type BackendModule struct {
	DeployModule       bool
	ModuleName         string
	ModuleExposedPorts *nat.PortSet
	ModulePortBindings *nat.PortMap
	ModuleEnvironment  map[string]interface{}

	DeploySidecar       bool
	SidecarExposedPorts *nat.PortSet
	SidecarPortBindings *nat.PortMap
}

type FrontendModule struct {
	DeployModule bool
	ModuleName   string
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

func AppendModuleEnvironment(environmentMap map[string]interface{}, moduleEnvironment []string) []string {
	if len(environmentMap) > 0 {
		for key, value := range environmentMap {
			moduleEnvironment = append(moduleEnvironment, fmt.Sprintf("%s=%s", strings.ToUpper(key), value))
		}
	}

	return moduleEnvironment
}

func AppendVaultEnvironment(moduleEnvironment []string, vaultToken string, vaultUrl string) []string {
	moduleEnvironment = append(moduleEnvironment, "SECRET_STORE_TYPE=VAULT")
	moduleEnvironment = append(moduleEnvironment, fmt.Sprintf("SECRET_STORE_VAULT_TOKEN=%s", vaultToken))
	moduleEnvironment = append(moduleEnvironment, fmt.Sprintf("SECRET_STORE_VAULT_ADDRESS=%s", vaultUrl))

	return moduleEnvironment
}

func NewRegisterModuleDto(registryUrls map[string]string,
	registryModules map[string][]RegistryModule,
	backendModulesMap map[string]BackendModule,
	frontendModulesMap map[string]FrontendModule,
	moduleDescriptorsMap map[string]interface{},
	fileModuleEnvPointer *os.File,
	enableDebug bool) *RegisterModuleDto {
	return &RegisterModuleDto{
		RegistryUrls:         registryUrls,
		RegistryModules:      registryModules,
		BackendModulesMap:    backendModulesMap,
		FrontendModulesMap:   frontendModulesMap,
		ModuleDescriptorsMap: moduleDescriptorsMap,
		FileModuleEnvPointer: fileModuleEnvPointer,
		EnableDebug:          enableDebug,
	}
}

func NewBackendModuleAndSidecar(name string, port int, deploySidecar bool, moduleEnvironment map[string]interface{}) *BackendModule {
	exposedPorts := CreateExposedPorts()
	modulePortBindings := CreatePortBindings(port, port+1000)
	sidecarPortBindings := CreatePortBindings(port+2000, port+3000)

	return &BackendModule{true, name, exposedPorts, modulePortBindings, moduleEnvironment, deploySidecar, exposedPorts, sidecarPortBindings}
}

func NewBackendModule(name string, port int, moduleEnvironment map[string]interface{}) *BackendModule {
	exposedPorts := CreateExposedPorts()
	modulePortBindings := CreatePortBindings(port, port+1000)

	return &BackendModule{true, name, exposedPorts, modulePortBindings, moduleEnvironment, false, nil, nil}
}

func NewFrontendModule(name string) *FrontendModule {
	return &FrontendModule{true, name}
}

func NewDeployManagementModulesDto(vaultToken string,
	registryHostname map[string]string,
	registryModules map[string][]RegistryModule,
	backendModulesMap map[string]BackendModule,
	globalEnvironment []string) *DeployModulesDto {
	return &DeployModulesDto{
		VaultToken:        vaultToken,
		RegistryHostname:  registryHostname,
		RegistryModules:   registryModules,
		BackendModulesMap: backendModulesMap,
		GlobalEnvironment: globalEnvironment,
		ManagementOnly:    true,
	}
}

func NewDeployModulesDto(vaultToken string,
	registryHostname map[string]string,
	registryModules map[string][]RegistryModule,
	backendModulesMap map[string]BackendModule,
	globalEnvironment []string) *DeployModulesDto {
	return &DeployModulesDto{
		VaultToken:        vaultToken,
		RegistryHostname:  registryHostname,
		RegistryModules:   registryModules,
		BackendModulesMap: backendModulesMap,
		GlobalEnvironment: globalEnvironment,
		ManagementOnly:    false,
	}
}

func NewDeployModuleDto(name string,
	image string,
	env []string,
	backendModule BackendModule,
	networkConfig *network.NetworkingConfig) *DeployModuleDto {
	return &DeployModuleDto{
		Name:         name,
		Image:        image,
		RegistryAuth: "",
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
	image string,
	env []string,
	backendModule BackendModule,
	networkConfig *network.NetworkingConfig,
	pullSidecarImage bool) *DeployModuleDto {
	return &DeployModuleDto{
		Name:         name,
		Image:        image,
		RegistryAuth: "",
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

func CreateExposedPorts() *nat.PortSet {
	moduleExposedPorts := make(map[nat.Port]struct{})

	moduleExposedPorts[nat.Port(ServerPort)] = struct{}{}
	moduleExposedPorts[nat.Port(DebugPort)] = struct{}{}

	portSet := nat.PortSet(moduleExposedPorts)

	return &portSet
}

func CreatePortBindings(hostServerPort int, hostServerDebugPort int) *nat.PortMap {
	var (
		serverPortBinding      []nat.PortBinding
		serverDebugPortBinding []nat.PortBinding
	)

	serverPortBinding = append(serverPortBinding, nat.PortBinding{HostIP: HostIp, HostPort: strconv.Itoa(hostServerPort)})
	serverDebugPortBinding = append(serverDebugPortBinding, nat.PortBinding{HostIP: HostIp, HostPort: strconv.Itoa(hostServerDebugPort)})

	portBindings := make(map[nat.Port][]nat.PortBinding)
	portBindings[nat.Port(ServerPort)] = serverPortBinding
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

func GetVaultToken(commandName string, client *client.Client) string {
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
			vaultToken := strings.TrimSpace(VaultTokenRegexp.ReplaceAllString(parsedLogLine, `$1`))

			slog.Info(commandName, "Found Vault Token", vaultToken)

			return vaultToken
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
		slog.Warn(commandName, fmt.Sprintf("cli.ContainerCreate warnings=\"%s\"", cr.Warnings), "")
	}

	err = client.ContainerStart(context.Background(), cr.ID, container.StartOptions{})
	if err != nil {
		slog.Error(commandName, "cli.ContainerStart error", "")
		panic(err)
	}

	slog.Info(commandName, "Deployed module container", fmt.Sprintf("%s %s", containerName, cr.ID))
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

	slog.Info(commandName, "Undeployed module container", fmt.Sprintf("%s %s %s", deployedModule.ID, deployedModule.Image, deployedModule.Status))
}

func DeployModules(commandName string, client *client.Client, dto *DeployModulesDto) {
	sidecarImage := viper.GetString(RegistrySidecarImageKey)
	resourceUrlKeycloak := viper.GetString(ResourcesKeycloakKey)
	resourceUrlVault := viper.GetString(ResourcesVaultKey)
	resourceUrlMgrTenants := viper.GetString(ResourcesMgrTenantsKey)
	resourceUrlMgrApplications := viper.GetString(ResourcesMgrApplicationsKey)
	resourceUrlMgrTenantEntitlements := viper.GetString(ResourcesMgrTenantEntitlements)

	networkConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{NetworkName: {NetworkID: NetworkId}},
	}

	pullSidecarImage := true

	for registryName, registryModules := range dto.RegistryModules {
		slog.Info(commandName, fmt.Sprintf("Deploying %s registry modules", registryName), "")

		for _, module := range registryModules {
			managementModule := strings.Contains(module.Name, ManagementModulePattern)
			if dto.ManagementOnly && !managementModule || !dto.ManagementOnly && managementModule {
				continue
			}

			backendModule, ok := dto.BackendModulesMap[module.Name]
			if !ok {
				continue
			}

			var image string
			if strings.Contains(module.Version, "SNAPSHOT") {
				image = fmt.Sprintf("folioci/%s:%s", module.Name, module.Version)
			} else {
				image = fmt.Sprintf("folioorg/%s:%s", module.Name, module.Version)
			}

			var combinedModuleEnvironment []string

			// Global environment
			combinedModuleEnvironment = append(combinedModuleEnvironment, dto.GlobalEnvironment...)

			// Module environment
			combinedModuleEnvironment = AppendModuleEnvironment(backendModule.ModuleEnvironment, combinedModuleEnvironment)

			// Vault environment
			combinedModuleEnvironment = AppendVaultEnvironment(combinedModuleEnvironment, dto.VaultToken, resourceUrlVault)

			deployModuleDto := NewDeployModuleDto(module.Name, image, combinedModuleEnvironment, backendModule, networkConfig)

			// TODO Use mutex and make parallel
			DeployModule(commandName, client, deployModuleDto)

			if backendModule.DeploySidecar {
				var combinedSidecarEnvironment []string

				// Keycloak environment
				combinedSidecarEnvironment = append(combinedSidecarEnvironment, fmt.Sprintf("KC_URL=%s", resourceUrlKeycloak))
				// TODO Fetch from environment map
				combinedSidecarEnvironment = append(combinedSidecarEnvironment, "KC_ADMIN_CLIENT_ID=superAdmin")
				combinedSidecarEnvironment = append(combinedSidecarEnvironment, "KC_SERVICE_CLIENT_ID=m2m-client")
				combinedSidecarEnvironment = append(combinedSidecarEnvironment, "KC_LOGIN_CLIENT_SUFFIX=-login-app")

				// Vault environment
				combinedSidecarEnvironment = AppendVaultEnvironment(combinedSidecarEnvironment, dto.VaultToken, resourceUrlVault)

				// Management environment
				combinedSidecarEnvironment = append(combinedSidecarEnvironment, fmt.Sprintf("TM_CLIENT_URL=%s", resourceUrlMgrTenants))
				combinedSidecarEnvironment = append(combinedSidecarEnvironment, fmt.Sprintf("AM_CLIENT_URL=%s", resourceUrlMgrApplications))
				combinedSidecarEnvironment = append(combinedSidecarEnvironment, fmt.Sprintf("TE_CLIENT_URL=%s", resourceUrlMgrTenantEntitlements))

				// Module environment
				combinedSidecarEnvironment = append(combinedSidecarEnvironment, fmt.Sprintf("MODULE_NAME=%s", module.Name))
				combinedSidecarEnvironment = append(combinedSidecarEnvironment, fmt.Sprintf("MODULE_VERSION=%s", module.Version))
				combinedSidecarEnvironment = append(combinedSidecarEnvironment, fmt.Sprintf("MODULE_URL=http://%s.eureka:%s", module.Name, ServerPort))
				combinedSidecarEnvironment = append(combinedSidecarEnvironment, fmt.Sprintf("SIDECAR_URL=http://%s-sc.eureka:%s", module.Name, ServerPort))

				// Java environment
				// TODO Move to config
				combinedSidecarEnvironment = append(combinedSidecarEnvironment, "JAVA_OPTIONS=-XX:MaxRAMPercentage=85.0 -Xms50m -Xmx128m -agentlib:jdwp=transport=dt_socket,server=y,suspend=n,address=*:5005")

				sidecarName := fmt.Sprintf("%s-sc", module.Name)

				deploySidecarDto := NewDeploySidecarDto(sidecarName, sidecarImage, combinedSidecarEnvironment, backendModule, networkConfig, pullSidecarImage)

				DeployModule(commandName, client, deploySidecarDto)

				pullSidecarImage = false
			}
		}
	}
}
