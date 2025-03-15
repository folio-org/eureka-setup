package internal

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"regexp"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/spf13/viper"
)

const (
	VaultRootTokenPattern string = ".*:"
	SidecarProjectName    string = "folio-module-sidecar"
)

func CreateClient(commandName string) *client.Client {
	newClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		slog.Error(commandName, GetFuncName(), "client.NewClientWithOpts error")
		panic(err)
	}

	return newClient
}

func GetRootVaultToken(commandName string, client *client.Client) string {
	logStream, err := client.ContainerLogs(context.Background(), "vault", container.LogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		slog.Error(commandName, GetFuncName(), "cli.ContainerLogs error")
		panic(err)
	}
	defer logStream.Close()

	buffer := make([]byte, 8)
	for {
		_, err := logStream.Read(buffer)
		if err != nil {
			slog.Error(commandName, GetFuncName(), "logStream.Read(buffer) error")
			panic(err)
		}

		count := binary.BigEndian.Uint32(buffer[4:])
		rawLogLine := make([]byte, count)

		_, err = logStream.Read(rawLogLine)
		if err != nil {
			slog.Warn(commandName, GetFuncName(), "logStream.Read(rawLogLine) encoutered an EOF")
		}

		parsedLogLine := string(rawLogLine)

		if strings.Contains(parsedLogLine, "init.sh: Root VAULT TOKEN is:") {
			vaultRootToken := strings.TrimSpace(regexp.MustCompile(VaultRootTokenPattern).ReplaceAllString(parsedLogLine, `$1`))

			return vaultRootToken
		}
	}
}

func DeployModules(commandName string, client *client.Client, dto *DeployModulesDto, sidecarImage string, sidecarResources *container.Resources) map[string]int {
	deployedModules := make(map[string]int)
	networkConfig := NewModuleNetworkConfig()

	for registryName, registryModules := range dto.RegistryModules {
		slog.Info(commandName, GetFuncName(), fmt.Sprintf("Deploying %s modules", registryName))

		for _, registryModule := range registryModules {
			managementModule := strings.Contains(registryModule.Name, ManagementModulePattern)
			if (dto.ManagementOnly && !managementModule) || (!dto.ManagementOnly && managementModule) {
				continue
			}

			backendModule, ok := dto.BackendModulesMap[registryModule.Name]
			if !ok || !backendModule.DeployModule {
				continue
			}

			moduleVersion := getModuleImageVersion(backendModule, registryModule)
			moduleImage := getModuleImage(commandName, moduleVersion, registryModule)
			moduleEnvironment := getModuleEnvironment(dto, backendModule)
			DeployModule(commandName, client, NewDeployModuleDto(registryModule.Name, moduleImage, moduleEnvironment, backendModule, networkConfig))

			deployedModules[registryModule.Name] = backendModule.ModuleExposedServerPort

			if backendModule.DeploySidecar && sidecarImage != "" {
				go func() {
					sidecarEnvironment := getSidecarEnvironment(dto, registryModule, backendModule)
					DeployModule(commandName, client, NewDeploySidecarDto(registryModule.SidecarName, sidecarImage, sidecarEnvironment, backendModule, networkConfig, sidecarResources))
				}()
			}
		}
	}

	return deployedModules
}

func getModuleImageVersion(backendModule BackendModule, registryModule *RegistryModule) string {
	if backendModule.ModuleVersion != nil {
		return *backendModule.ModuleVersion
	}

	return *registryModule.Version
}

func GetSidecarImage(commandName string, registryModules []*RegistryModule) string {
	sidecarModule := viper.GetStringMap(SidecarModule)

	sidecarImageVersion := getSidecarImageVersion(commandName, registryModules, sidecarModule["version"])
	sidecarImage := fmt.Sprintf("%s/%s", GetImageRegistryNamespace(commandName, sidecarImageVersion),
		fmt.Sprintf("%s:%s", sidecarModule["image"].(string), sidecarImageVersion))

	slog.Info(commandName, GetFuncName(), fmt.Sprintf("Using sidecar image %s", sidecarImage))

	return sidecarImage
}

func getSidecarImageVersion(commandName string, registryModules []*RegistryModule, sidecarConfigVersion any) string {
	ok, sidecarRegistryVersion := findSidecarRegistryVersion(registryModules)
	if !ok && sidecarConfigVersion == nil {
		LogErrorPanic(commandName, "internal.GetImageVersion error - Sidecar version is not found in registry or in the current config")
		return ""
	}

	if sidecarConfigVersion != nil {
		return sidecarConfigVersion.(string)
	}

	return sidecarRegistryVersion
}

func findSidecarRegistryVersion(registryModules []*RegistryModule) (bool, string) {
	for _, registryModule := range registryModules {
		if registryModule.Name == SidecarProjectName {
			return true, *registryModule.Version
		}
	}

	return false, ""
}

func getModuleImage(commandName string, moduleVersion string, registryModule *RegistryModule) string {
	return fmt.Sprintf("%s/%s:%s", GetImageRegistryNamespace(commandName, moduleVersion), registryModule.Name, moduleVersion)
}

func getModuleEnvironment(deployModulesDto *DeployModulesDto, backendModule BackendModule) []string {
	var combinedEnvironment []string
	combinedEnvironment = append(combinedEnvironment, deployModulesDto.GlobalEnvironment...)
	combinedEnvironment = AppendModuleEnvironment(combinedEnvironment, backendModule.ModuleEnvironment)
	combinedEnvironment = AppendVaultEnvironment(combinedEnvironment, deployModulesDto.VaultRootToken)

	return combinedEnvironment
}

func getSidecarEnvironment(deployModulesDto *DeployModulesDto, module *RegistryModule, backendModule BackendModule) []string {
	var combinedEnvironment []string
	combinedEnvironment = append(combinedEnvironment, deployModulesDto.SidecarEnvironment...)
	combinedEnvironment = AppendKeycloakEnvironment(combinedEnvironment)
	combinedEnvironment = AppendVaultEnvironment(combinedEnvironment, deployModulesDto.VaultRootToken)
	combinedEnvironment = AppendManagementEnvironment(combinedEnvironment)
	combinedEnvironment = AppendSidecarEnvironment(combinedEnvironment, module, strconv.Itoa(backendModule.ModuleServerPort))

	return combinedEnvironment
}

func DeployModule(commandName string, client *client.Client, dto *DeployModuleDto) {
	containerName := getContainerName(dto)

	if dto.PullImage {
		PullModule(commandName, client, dto.Image)
	}

	cr, err := client.ContainerCreate(context.Background(), dto.Config, dto.HostConfig, dto.NetworkConfig, dto.Platform, containerName)
	if err != nil {
		slog.Error(commandName, GetFuncName(), "cli.ContainerCreate error")
		panic(err)
	}

	if len(cr.Warnings) > 0 {
		slog.Warn(commandName, GetFuncName(), fmt.Sprintf("cli.ContainerCreate warnings, '%s'", cr.Warnings))
	}

	err = client.ContainerStart(context.Background(), cr.ID, container.StartOptions{})
	if err != nil {
		slog.Error(commandName, GetFuncName(), "cli.ContainerStart error")
		panic(err)
	}

	slog.Info(commandName, GetFuncName(), fmt.Sprintf("Deployed module container %s %s", cr.ID, containerName))
}

func getContainerName(dto *DeployModuleDto) string {
	if strings.HasPrefix(dto.Name, ManagementModulePattern) {
		return fmt.Sprintf("eureka-%s", dto.Name)
	}

	return fmt.Sprintf("eureka-%s-%s", viper.GetString(ProfileNameKey), dto.Name)
}

func PullModule(commandName string, client *client.Client, image string) {
	registryAuthToken := GetRegistryAuthTokenIfPresent(commandName)

	reader, err := client.ImagePull(context.Background(), image, types.ImagePullOptions{RegistryAuth: registryAuthToken})
	if err != nil {
		slog.Error(commandName, GetFuncName(), "cli.ImagePull error")
		panic(err)
	}
	defer reader.Close()

	decoder := json.NewDecoder(reader)

	var event *Event
	for {
		if err := decoder.Decode(&event); err == io.EOF {
			break
		} else if err != nil {
			slog.Error(commandName, GetFuncName(), "decoder.Decode error")
			panic(err)
		}

		if event.Error != "" {
			slog.Error(commandName, GetFuncName(), fmt.Sprintf("Pulling module container image %+v", event.Error))
		}
	}
}

func GetDeployedModules(commandName string, client *client.Client, filters filters.Args) []types.Container {
	deployedModules, err := client.ContainerList(context.Background(), container.ListOptions{All: true, Filters: filters})
	if err != nil {
		slog.Error(commandName, GetFuncName(), "cli.ContainerList error")
		panic(err)
	}

	return deployedModules
}

func UndeployModule(commandName string, client *client.Client, deployedModule types.Container) {
	err := client.ContainerStop(context.Background(), deployedModule.ID, container.StopOptions{Signal: "9"})
	if err != nil {
		slog.Error(commandName, GetFuncName(), "cli.ContainerStop error")
		panic(err)
	}

	go func() {
		err = client.ContainerRemove(context.Background(), deployedModule.ID, container.RemoveOptions{Force: true, RemoveVolumes: true})
		if err != nil {
			slog.Error(commandName, GetFuncName(), "cli.ContainerRemove error")
			panic(err)
		}
	}()

	slog.Info(commandName, GetFuncName(), fmt.Sprintf("Undeployed module container %s %s %s", deployedModule.ID,
		strings.ReplaceAll(deployedModule.Names[0], "/", ""), deployedModule.Status))
}
