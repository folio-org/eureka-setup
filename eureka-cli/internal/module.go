package internal

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
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

const VaultRootTokenPattern string = ".*:"

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

			slog.Info(commandName, GetFuncName(), fmt.Sprintf("Found vault root token: %s", vaultRootToken))

			return vaultRootToken
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

	err = client.ContainerRemove(context.Background(), deployedModule.ID, container.RemoveOptions{Force: true, RemoveVolumes: true})
	if err != nil {
		slog.Error(commandName, GetFuncName(), "cli.ContainerRemove error")
		panic(err)
	}

	slog.Info(commandName, GetFuncName(), fmt.Sprintf("Undeployed module container %s %s %s", deployedModule.ID, strings.ReplaceAll(deployedModule.Names[0], "/", ""), deployedModule.Status))
}

func DeployModule(commandName string, client *client.Client, dto *DeployModuleDto) {
	var containerName string
	if strings.HasPrefix(dto.Name, ManagementModulePattern) {
		containerName = fmt.Sprintf("eureka-%s", dto.Name)
	} else {
		containerName = fmt.Sprintf("eureka-%s-%s", viper.GetString(ProfileNameKey), dto.Name)
	}

	if dto.PullImage {
		reader, err := client.ImagePull(context.Background(), dto.Image, types.ImagePullOptions{RegistryAuth: dto.RegistryAuth})
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

	cr, err := client.ContainerCreate(context.Background(), dto.Config, dto.HostConfig, dto.NetworkConfig, dto.Platform, containerName)
	if err != nil {
		slog.Error(commandName, GetFuncName(), "cli.ContainerCreate error")
		panic(err)
	}

	if cr.Warnings != nil && len(cr.Warnings) > 0 {
		slog.Warn(commandName, GetFuncName(), fmt.Sprintf("cli.ContainerCreate warnings, '%s'", cr.Warnings))
	}

	err = client.ContainerStart(context.Background(), cr.ID, container.StartOptions{})
	if err != nil {
		slog.Error(commandName, GetFuncName(), "cli.ContainerStart error")
		panic(err)
	}

	slog.Info(commandName, GetFuncName(), fmt.Sprintf("Deployed module container %s %s", cr.ID, containerName))
}

func DeployModules(commandName string, client *client.Client, dto *DeployModulesDto) map[string]int {
	deployedModules := make(map[string]int)

	sidecarImage, sidecarImageVersion := GetSidecarImage(commandName, dto.RegistryModules["eureka"])
	authToken := GetRegistryAuthTokenIfPresent(commandName)
	networkConfig := NewModuleNetworkConfig()
	pullSidecarImage := true

	for registryName, registryModules := range dto.RegistryModules {
		slog.Info(commandName, GetFuncName(), fmt.Sprintf("Deploying %s modules", registryName))

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
				continue
			}

			var moduleVersion string
			if backendModule.ModuleVersion != nil {
				moduleVersion = *backendModule.ModuleVersion
			} else {
				moduleVersion = *module.Version
			}

			moduleImage := fmt.Sprintf("%s/%s:%s", GetImageRegistryNamespace(commandName, moduleVersion), module.Name, moduleVersion)

			var combinedModuleEnvironment []string
			combinedModuleEnvironment = append(combinedModuleEnvironment, dto.GlobalEnvironment...)
			combinedModuleEnvironment = AppendModuleEnvironment(backendModule.ModuleEnvironment, combinedModuleEnvironment)
			combinedModuleEnvironment = AppendVaultEnvironment(combinedModuleEnvironment, dto.VaultRootToken)

			deployModuleDto := NewDeployModuleDto(module.Name, moduleVersion, moduleImage, combinedModuleEnvironment, backendModule, networkConfig, authToken)

			DeployModule(commandName, client, deployModuleDto)

			deployedModules[module.Name] = backendModule.ModuleExposedServerPort

			if !backendModule.DeploySidecar {
				continue
			}

			var combinedSidecarEnvironment []string
			combinedSidecarEnvironment = append(combinedSidecarEnvironment, dto.SidecarEnvironment...)
			combinedSidecarEnvironment = AppendKeycloakEnvironment(commandName, combinedSidecarEnvironment)
			combinedSidecarEnvironment = AppendVaultEnvironment(combinedSidecarEnvironment, dto.VaultRootToken)
			combinedSidecarEnvironment = AppendManagementEnvironment(combinedSidecarEnvironment)
			combinedSidecarEnvironment = AppendSidecarEnvironment(combinedSidecarEnvironment, module, strconv.Itoa(backendModule.ModuleServerPort))

			deploySidecarDto := NewDeploySidecarDto(module.SidecarName, sidecarImageVersion, sidecarImage, combinedSidecarEnvironment, backendModule, networkConfig, pullSidecarImage, authToken)

			DeployModule(commandName, client, deploySidecarDto)

			pullSidecarImage = false
		}
	}

	return deployedModules
}

func GetSidecarImage(commandName string, registryModules []*RegistryModule) (string, string) {
	sidecarModule := viper.GetStringMap(SidecarModule)
	sidecarImageName := sidecarModule["image"].(string)
	sidecarImageVersion := GetImageVersion(commandName, registryModules, sidecarModule["version"])
	sidecarImageNamespace := GetImageRegistryNamespace(commandName, sidecarImageVersion)
	sidecarImage := fmt.Sprintf("%s/%s", sidecarImageNamespace, fmt.Sprintf("%s:%s", sidecarImageName, sidecarImageVersion))

	slog.Info(commandName, GetFuncName(), fmt.Sprintf("Using sidecar image %s", sidecarImage))

	return sidecarImage, sidecarImageVersion
}

func GetImageVersion(commandName string, registryModules []*RegistryModule, imageVersion interface{}) string {
	if imageVersion != nil {
		return imageVersion.(string)
	}

	for _, module := range registryModules {
		if module.Name == "folio-module-sidecar" {
			return *module.Version
		}
	}

	errorMessage := "internal.GetImageVersion error - Sidecar module not found in registry"
	slog.Error(commandName, GetFuncName(), errorMessage)
	panic(errors.New(errorMessage))
}
