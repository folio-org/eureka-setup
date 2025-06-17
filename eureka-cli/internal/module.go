package internal

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"regexp"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/spf13/viper"
)

const (
	ColonDelimitedPattern string = ".*:"
	VaultContainerName    string = "vault"
	SidecarProjectName    string = "folio-module-sidecar"
)

func CreateDockerClient(commandName string) *client.Client {
	newClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		slog.Error(commandName, GetFuncName(), "client.NewClientWithOpts error")
		panic(err)
	}

	return newClient
}

func GetVaultRootToken(commandName string, client *client.Client) string {
	logStream, err := client.ContainerLogs(context.Background(), VaultContainerName, container.LogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		slog.Error(commandName, GetFuncName(), "client.ContainerLogs error")
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
			vaultRootToken := strings.TrimSpace(regexp.MustCompile(ColonDelimitedPattern).ReplaceAllString(parsedLogLine, `$1`))

			return vaultRootToken
		}
	}
}

func DeployModules(commandName string, client *client.Client, dto *DeployModulesDto, sidecarImage string, sidecarResources *container.Resources) map[string]int {
	deployedModules := make(map[string]int)
	networkConfig := NewModuleNetworkConfig()

	for registryName, registryModules := range dto.RegistryModules {
		if len(registryModules) > 0 {
			slog.Info(commandName, GetFuncName(), fmt.Sprintf("Deploying %s modules", registryName))
		}

		for _, registryModule := range registryModules {
			managementModule := strings.Contains(registryModule.Name, ManagementModulePattern)
			if (dto.ManagementOnly && !managementModule) || (!dto.ManagementOnly && managementModule) {
				continue
			}

			backendModule, ok := dto.BackendModulesMap[registryModule.Name]
			if !ok || !backendModule.DeployModule {
				continue
			}

			moduleVersion := GetModuleImageVersion(backendModule, registryModule)
			moduleImage := GetModuleImage(commandName, moduleVersion, registryModule)
			moduleEnvironment := GetModuleEnvironment(dto, registryModule, backendModule)
			moduleDeployDto := NewDeployModuleDto(registryModule.Name, moduleImage, moduleEnvironment, backendModule, networkConfig)

			DeployModule(commandName, client, moduleDeployDto)

			deployedModules[registryModule.Name] = backendModule.ModuleExposedServerPort

			if backendModule.DeploySidecar && sidecarImage != "" {
				go func() {
					sidecarEnvironment := GetSidecarEnvironment(dto, registryModule, backendModule, nil, nil)
					sidecarDeployDto := NewDeploySidecarDto(registryModule.SidecarName, sidecarImage, sidecarEnvironment, backendModule, networkConfig, sidecarResources)

					DeployModule(commandName, client, sidecarDeployDto)
				}()
			}
		}
	}

	return deployedModules
}

func GetBackendModule(commandName string, dto *DeployModulesDto, moduleName string) (*BackendModule, *RegistryModule) {
	for _, registryModules := range dto.RegistryModules {
		for _, registryModule := range registryModules {
			backendModule, ok := dto.BackendModulesMap[registryModule.Name]
			if !ok || !backendModule.DeployModule {
				continue
			}

			if registryModule.Name == moduleName {
				return &backendModule, registryModule
			}
		}
	}

	return nil, nil
}

func GetModuleImageVersion(backendModule BackendModule, registryModule *RegistryModule) string {
	if backendModule.ModuleVersion != nil {
		return *backendModule.ModuleVersion
	}

	return *registryModule.Version
}

func GetSidecarImage(commandName string, registryModules []*RegistryModule) (string, bool) {
	sidecarModule := viper.GetStringMap(SidecarModuleKey)
	sidecarImageVersion := getSidecarImageVersion(commandName, registryModules, sidecarModule[SidecarModuleVersionEntryKey])

	localImage := sidecarModule[SidecarModuleLocalImageEntryKey]
	if localImage != nil && localImage.(string) != "" {
		return fmt.Sprintf("%s:%s", localImage.(string), sidecarImageVersion), false
	}

	namespace := GetImageRegistryNamespace(commandName, sidecarImageVersion)
	image := sidecarModule[SidecarModuleImageEntryKey]
	return fmt.Sprintf("%s/%s", namespace, fmt.Sprintf("%s:%s", image.(string), sidecarImageVersion)), true
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

func GetModuleImage(commandName string, moduleVersion string, registryModule *RegistryModule) string {
	return fmt.Sprintf("%s/%s:%s", GetImageRegistryNamespace(commandName, moduleVersion), registryModule.Name, moduleVersion)
}

func GetModuleEnvironment(deployModulesDto *DeployModulesDto, module *RegistryModule, backendModule BackendModule) []string {
	var combinedEnvironment []string
	combinedEnvironment = append(combinedEnvironment, deployModulesDto.GlobalEnvironment...)

	if backendModule.UseVault {
		combinedEnvironment = AppendVaultEnvironment(combinedEnvironment, deployModulesDto.VaultRootToken)
	}
	if backendModule.UseOkapiUrl {
		combinedEnvironment = AppendOkapiEnvironment(combinedEnvironment, module.SidecarName, backendModule.ModuleServerPort)
	}
	if backendModule.DisableSystemUser {
		combinedEnvironment = AppendDisableSystemUserEnvironment(combinedEnvironment, module.Name)
	}
	combinedEnvironment = AppendModuleEnvironment(combinedEnvironment, backendModule.ModuleEnvironment)

	return combinedEnvironment
}

func GetSidecarEnvironment(deployModulesDto *DeployModulesDto, module *RegistryModule, backendModule BackendModule, moduleUrl *string, sidecarUrl *string) []string {
	var combinedEnvironment []string
	combinedEnvironment = append(combinedEnvironment, deployModulesDto.SidecarEnvironment...)
	combinedEnvironment = AppendVaultEnvironment(combinedEnvironment, deployModulesDto.VaultRootToken)
	combinedEnvironment = AppendKeycloakEnvironment(combinedEnvironment)
	combinedEnvironment = AppendSidecarEnvironment(combinedEnvironment, module, backendModule.ModuleServerPort, moduleUrl, sidecarUrl)

	return combinedEnvironment
}

func DeployModule(commandName string, client *client.Client, dto *DeployModuleDto) {
	containerName := getContainerName(dto)

	if dto.PullImage {
		PullModule(commandName, client, dto.Image)
	}

	cr, err := client.ContainerCreate(context.Background(), dto.Config, dto.HostConfig, dto.NetworkConfig, dto.Platform, containerName)
	if err != nil {
		slog.Error(commandName, GetFuncName(), "client.ContainerCreate error")
		panic(err)
	}

	if len(cr.Warnings) > 0 {
		slog.Warn(commandName, GetFuncName(), fmt.Sprintf("client.ContainerCreate warning - %s", cr.Warnings))
	}

	err = client.ContainerStart(context.Background(), cr.ID, container.StartOptions{})
	if err != nil {
		slog.Error(commandName, GetFuncName(), "client.ContainerStart error")
		panic(err)
	}

	slog.Info(commandName, GetFuncName(), fmt.Sprintf("Deployed module container, id: %s, name: %s", cr.ID, containerName))
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
		slog.Error(commandName, GetFuncName(), "client.ImagePull error")
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
		slog.Error(commandName, GetFuncName(), "client.ContainerList error")
		panic(err)
	}

	return deployedModules
}

func UndeployModuleByNamePattern(commandName string, client *client.Client, value string, removeAsync bool) {
	deployedModules := GetDeployedModules(commandName, client, filters.NewArgs(filters.KeyValuePair{Key: "name", Value: value}))
	for _, deployedModule := range deployedModules {
		undeployModule(commandName, client, deployedModule, removeAsync)
	}
}

func undeployModule(commandName string, client *client.Client, deployedModule types.Container, removeAsync bool) {
	err := client.NetworkDisconnect(context.Background(), DefaultNetworkId, deployedModule.ID, false)
	if err != nil {
		slog.Warn(commandName, GetFuncName(), fmt.Sprintf("client.NetworkDisconnect warning - %s", err.Error()))
	}

	err = client.ContainerStop(context.Background(), deployedModule.ID, container.StopOptions{Signal: "9"})
	if err != nil {
		slog.Error(commandName, GetFuncName(), "client.ContainerStop error")
		panic(err)
	}

	var removeContainerFunc = func() {
		err = client.ContainerRemove(context.Background(), deployedModule.ID, container.RemoveOptions{Force: true, RemoveVolumes: true})
		if err != nil {
			slog.Error(commandName, GetFuncName(), "client.ContainerRemove error")
			panic(err)
		}
	}

	if removeAsync {
		go removeContainerFunc()
	} else {
		removeContainerFunc()
	}

	containerName := strings.ReplaceAll(deployedModule.Names[0], "/", "")
	slog.Info(commandName, GetFuncName(), fmt.Sprintf("Undeployed module container, id: %s, name: %s, status: %s", deployedModule.ID, containerName, deployedModule.Status))
}
