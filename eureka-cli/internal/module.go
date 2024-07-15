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
	RegistryUrls              map[string]string
	RegistryModules           map[string][]RegistryModule
	BackendModulesMap         map[string]BackendModule
	ModuleDescriptorsMap      map[string]interface{}
	CacheFileModuleEnvPointer *os.File
	EnableDebug               bool
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
	VaultToken              string
	RegistryHostname        map[string]string
	EurekaRegistryAuthToken string
	RegistryModules         map[string][]RegistryModule
	BackendModulesMap       map[string]BackendModule
	SharedEnv               []string
}

type BackendModule struct {
	DeployModule       bool
	ModuleName         string
	ModuleExposedPorts *nat.PortSet
	ModulePortBindings *nat.PortMap
	ModuleEnv          map[string]interface{}

	DeploySidecar       bool
	SidecarExposedPorts *nat.PortSet
	SidecarPortBindings *nat.PortMap
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

func AppendModuleEnv(envMap map[string]interface{}, moduleEnv []string) []string {
	if len(envMap) > 0 {
		for key, value := range envMap {
			moduleEnv = append(moduleEnv, fmt.Sprintf("%s=%s", strings.ToUpper(key), value))
		}
	}

	return moduleEnv
}

func AppendVaultEnv(moduleEnv []string, vaultToken string, vaultUrl string) []string {
	moduleEnv = append(moduleEnv, "SECRET_STORE_TYPE=VAULT")
	moduleEnv = append(moduleEnv, fmt.Sprintf("SECRET_STORE_VAULT_TOKEN=%s", vaultToken))
	moduleEnv = append(moduleEnv, fmt.Sprintf("SECRET_STORE_VAULT_ADDRESS=%s", vaultUrl))

	return moduleEnv
}

func NewBackendModuleWithSidecar(name string, port int, moduleEnv map[string]interface{}) *BackendModule {
	exposedPorts := CreateExposedPorts()
	modulePortBindings := CreatePortBindings(port, port+1000)
	sidecarPortBindings := CreatePortBindings(port+2000, port+3000)

	return &BackendModule{true, name, exposedPorts, modulePortBindings, moduleEnv, true, exposedPorts, sidecarPortBindings}
}

func NewBackendModule(name string, port int, moduleEnv map[string]interface{}) *BackendModule {
	exposedPorts := CreateExposedPorts()
	modulePortBindings := CreatePortBindings(port, port+1000)

	return &BackendModule{true, name, exposedPorts, modulePortBindings, moduleEnv, false, nil, nil}
}

func CreateExposedPorts() *nat.PortSet {
	moduleExposedPorts := make(map[nat.Port]struct{})
	moduleExposedPorts[nat.Port(ServerPort)] = struct{}{}
	moduleExposedPorts[nat.Port(DebugPort)] = struct{}{}

	portSet := nat.PortSet(moduleExposedPorts)

	return &portSet
}

func CreatePortBindings(hostServerPort int, hostServerDebugPort int) *nat.PortMap {
	var serverPortBinding []nat.PortBinding
	var serverDebugPortBinding []nat.PortBinding

	serverPortBinding = append(serverPortBinding, nat.PortBinding{HostIP: HostIp, HostPort: strconv.Itoa(hostServerPort)})
	serverDebugPortBinding = append(serverDebugPortBinding, nat.PortBinding{HostIP: HostIp, HostPort: strconv.Itoa(hostServerDebugPort)})

	portBindings := make(map[nat.Port][]nat.PortBinding)
	portBindings[nat.Port(ServerPort)] = serverPortBinding
	portBindings[nat.Port(DebugPort)] = serverDebugPortBinding

	portMap := nat.PortMap(portBindings)

	return &portMap
}

func CreateContainerdCli(commandName string) *client.Client {
	containerdCli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		slog.Error(commandName, SecondaryMessageKey, "client.NewClientWithOpts error")
		panic(err)
	}

	return containerdCli
}

func GetVaultToken(commandName string, containerdCli *client.Client) string {
	logStream, err := containerdCli.ContainerLogs(context.Background(), "vault", container.LogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		slog.Error(commandName, SecondaryMessageKey, "cli.ContainerLogs error")
		panic(err)
	}
	defer logStream.Close()

	buffer := make([]byte, 8)
	for {
		_, err := logStream.Read(buffer)
		if err != nil {
			slog.Error(commandName, SecondaryMessageKey, "logStream.Read(buffer) error")
			panic(err)
		}

		count := binary.BigEndian.Uint32(buffer[4:])
		rawLogLine := make([]byte, count)

		_, err = logStream.Read(rawLogLine)
		if err != nil {
			slog.Warn(commandName, SecondaryMessageKey, "logStream.Read(rawLogLine) encoutered an EOF")
		}

		parsedLogLine := string(rawLogLine)

		if strings.Contains(parsedLogLine, "init.sh: Root VAULT TOKEN is:") {
			vaultToken := strings.TrimSpace(VaultTokenRegexp.ReplaceAllString(parsedLogLine, `$1`))

			slog.Info(commandName, "Found Vault Token", vaultToken)

			return vaultToken
		}
	}
}

func GetDeployedModules(commandName string, containerdCli *client.Client, filters filters.Args) []types.Container {
	deployedModules, err := containerdCli.ContainerList(context.Background(), container.ListOptions{All: true, Filters: filters})
	if err != nil {
		slog.Error(commandName, SecondaryMessageKey, "cli.ContainerList error")
		panic(err)
	}

	if len(deployedModules) == 0 {
		slog.Info(commandName, SecondaryMessageKey, "No deployed module containers were found")
	}

	return deployedModules
}

func DeployModule(commandName string, containerdCli *client.Client, dto *DeployModuleDto) {
	containerName := fmt.Sprintf("eureka-%s", dto.Name)

	if dto.PullImage {
		reader, err := containerdCli.ImagePull(context.Background(), dto.Image, types.ImagePullOptions{RegistryAuth: dto.RegistryAuth})
		if err != nil {
			slog.Error(commandName, SecondaryMessageKey, "cli.ImagePull error")
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

			if event.Error == "" {
				slog.Info(commandName, "Pulling module container image", fmt.Sprintf("%+v", event.Status))
			} else {
				slog.Error(commandName, "Pulling module container image", fmt.Sprintf("%+v", event.Error))
			}
		}
	}

	cr, err := containerdCli.ContainerCreate(context.Background(), dto.Config, dto.HostConfig, dto.NetworkConfig, dto.Platform, containerName)
	if err != nil {
		slog.Error(commandName, SecondaryMessageKey, "cli.ContainerCreate error")
		panic(err)
	}

	if cr.Warnings != nil && len(cr.Warnings) > 0 {
		slog.Warn(commandName, SecondaryMessageKey, fmt.Sprintf("cli.ContainerCreate warnings=\"%s\"", cr.Warnings))
	}

	err = containerdCli.ContainerStart(context.Background(), cr.ID, container.StartOptions{})
	if err != nil {
		slog.Error(commandName, SecondaryMessageKey, "cli.ContainerStart error")
		panic(err)
	}

	slog.Info(commandName, "Deployed module container", fmt.Sprintf("%s %s", containerName, cr.ID))
}

func UndeployModule(commandName string, containerdCli *client.Client, deployedModule types.Container) {
	err := containerdCli.ContainerStop(context.Background(), deployedModule.ID, container.StopOptions{Signal: "9"})
	if err != nil {
		slog.Error(commandName, SecondaryMessageKey, "cli.ContainerStop error")
		panic(err)
	}

	err = containerdCli.ContainerRemove(context.Background(), deployedModule.ID, container.RemoveOptions{Force: true, RemoveVolumes: true})
	if err != nil {
		slog.Error(commandName, SecondaryMessageKey, "cli.ContainerRemove error")
		panic(err)
	}

	slog.Info(commandName, "Undeployed module container", fmt.Sprintf("%s %s %s", deployedModule.ID, deployedModule.Image, deployedModule.Status))
}

func DeployModules(commandName string, containerdCli *client.Client, dto *DeployModulesDto) {
	eurekaSidecarImage := fmt.Sprintf("%s/%s", viper.GetString(EurekaRegistryHostnameKey), viper.GetString(EurekaSidecarImageKey))

	resourceUrlKeycloak := viper.GetString(ResourceUrlKeycloakKey)
	resourceUrlVault := viper.GetString(ResourceUrlVaultKey)
	resourceUrlMgrTenants := viper.GetString(ResourceUrlMgrTenantsKey)
	resourceUrlMgrApplications := viper.GetString(ResourceUrlMgrApplicationsKey)
	resourceUrlMgrTenantEntitlements := viper.GetString(ResourceUrlMgrTenantEntitlements)

	networkConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{"fpm-net": {NetworkID: "eureka"}},
	}

	pullSidecarImage := true

	for registryName, registryModules := range dto.RegistryModules {
		slog.Info(commandName, PrimaryMessageKey, fmt.Sprintf("Deploying %s registry modules", registryName))

		for _, module := range registryModules {
			backendModule, ok := dto.BackendModulesMap[module.Name]
			if !ok {
				continue
			}

			moduleNameEnv := EnvNameRegexp.ReplaceAllString(strings.ToUpper(module.Name), `_`)

			var image string

			if registryName == "folio" {
				image = fmt.Sprintf("%s/%s:%s", dto.RegistryHostname["folio"], module.Name, module.Version)
			} else {
				image = fmt.Sprintf("%s/%s:%s", dto.RegistryHostname["eureka"], module.Name, module.Version)
			}

			var moduleEnv []string

			// Shared ENV
			moduleEnv = append(moduleEnv, dto.SharedEnv...)

			// Module ENV
			moduleEnv = AppendModuleEnv(backendModule.ModuleEnv, moduleEnv)

			// Vault ENV
			moduleEnv = AppendVaultEnv(moduleEnv, dto.VaultToken, resourceUrlVault)

			var moduleRegistryAuthToken string
			if registryName == "folio" {
				moduleRegistryAuthToken = ""
			} else {
				moduleRegistryAuthToken = dto.EurekaRegistryAuthToken
			}

			deployModuleDto := &DeployModuleDto{
				Name:         module.Name,
				Image:        image,
				RegistryAuth: moduleRegistryAuthToken,
				Config: &container.Config{
					Image:        image,
					Hostname:     strings.ToLower(moduleNameEnv),
					Env:          moduleEnv,
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

			// TODO Use mutex and make parallel
			DeployModule(commandName, containerdCli, deployModuleDto)

			if backendModule.DeploySidecar {
				var sidecarEnv []string

				// Keycloak ENV
				sidecarEnv = append(sidecarEnv, fmt.Sprintf("KC_URL=%s", resourceUrlKeycloak))
				sidecarEnv = append(sidecarEnv, "KC_ADMIN_CLIENT_ID=superAdmin")
				sidecarEnv = append(sidecarEnv, "KC_SERVICE_CLIENT_ID=m2m-client")
				sidecarEnv = append(sidecarEnv, "KC_LOGIN_CLIENT_SUFFIX=login-app")

				// Vault ENV
				sidecarEnv = AppendVaultEnv(sidecarEnv, dto.VaultToken, resourceUrlVault)

				// Management ENV
				sidecarEnv = append(sidecarEnv, fmt.Sprintf("TM_CLIENT_URL=%s", resourceUrlMgrTenants))
				sidecarEnv = append(sidecarEnv, fmt.Sprintf("AM_CLIENT_URL=%s", resourceUrlMgrApplications))
				sidecarEnv = append(sidecarEnv, fmt.Sprintf("TE_CLIENT_URL=%s", resourceUrlMgrTenantEntitlements))

				// Module ENV
				sidecarEnv = append(sidecarEnv, fmt.Sprintf("MODULE_NAME=%s", module.Name))
				sidecarEnv = append(sidecarEnv, fmt.Sprintf("MODULE_VERSION=%s", module.Version))
				sidecarEnv = append(sidecarEnv, fmt.Sprintf("MODULE_URL=http://%s.eureka:%s", module.Name, ServerPort))
				sidecarEnv = append(sidecarEnv, fmt.Sprintf("SIDECAR_URL=http://%s-sc.eureka:%s", module.Name, ServerPort))

				// Java ENV
				sidecarEnv = append(sidecarEnv, "JAVA_OPTIONS=-XX:MaxRAMPercentage=85.0 -Xms50m -Xmx128m -agentlib:jdwp=transport=dt_socket,server=y,suspend=n,address=*:5005")

				deploySidecarDto := &DeployModuleDto{
					Name:         fmt.Sprintf("%s-sc", module.Name),
					Image:        eurekaSidecarImage,
					RegistryAuth: dto.EurekaRegistryAuthToken,
					Config: &container.Config{
						Image:        eurekaSidecarImage,
						Hostname:     fmt.Sprintf("%s-sc", strings.ToLower(moduleNameEnv)),
						Env:          sidecarEnv,
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

				DeployModule(commandName, containerdCli, deploySidecarDto)

				pullSidecarImage = false
			}
		}
	}
}
