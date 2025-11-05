package modulesvc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/folio-org/eureka-cli/constant"
	appErrors "github.com/folio-org/eureka-cli/errors"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/folio-org/eureka-cli/models"
)

// ModuleManager defines the interface for managing module deployment and lifecycle
type ModuleManager interface {
	GetDeployedModules(client *client.Client, filters filters.Args) ([]container.Summary, error)
	PullModule(client *client.Client, imageName string) error
	DeployModules(client *client.Client, containers *models.Containers, sidecarImage string, sidecarResources *container.Resources) (map[string]int, error)
	DeployModule(client *client.Client, container *models.Container) error
	UndeployModuleByNamePattern(client *client.Client, pattern string) error
}

// SidecarRequest contains all the information needed to deploy a sidecar container
type SidecarRequest struct {
	Client           *client.Client
	Containers       *models.Containers
	RegistryModule   *models.RegistryModule
	BackendModule    models.BackendModule
	SidecarImage     string
	NetworkConfig    *network.NetworkingConfig
	SidecarResources *container.Resources
}

func (ms *ModuleSvc) GetDeployedModules(client *client.Client, filters filters.Args) ([]container.Summary, error) {
	ctx, cancel := context.WithTimeout(context.Background(), constant.ContextTimeoutDockerList)
	defer cancel()

	deployedModules, err := client.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filters,
	})
	if err != nil {
		return nil, err
	}

	return deployedModules, nil
}

func (ms *ModuleSvc) PullModule(client *client.Client, imageName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), constant.ContextTimeoutDockerImagePull)
	defer cancel()

	authorizationToken, err := ms.RegistrySvc.GetAuthorizationToken()
	if err != nil {
		return err
	}

	reader, err := client.ImagePull(ctx, imageName, image.PullOptions{
		RegistryAuth: authorizationToken,
	})
	if err != nil {
		return err
	}
	defer helpers.CloseReader(reader)
	decoder := json.NewDecoder(reader)

	var event *models.Event
	for {
		if err := decoder.Decode(&event); errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return err
		}

		if event.Error == "" {
			current := helpers.ConvertMemory(helpers.BytesToMib, int64(event.ProgressDetail.Current))
			total := helpers.ConvertMemory(helpers.BytesToMib, int64(event.ProgressDetail.Total))
			slog.Debug(ms.Action.Name, "text", "Pulling module", "imageName", imageName, "status", event.Status, "progressCurrent", current, "progressTotal", total)
		} else {
			return appErrors.ModulePullFailed(imageName, errors.New(event.Error))
		}
	}

	return nil
}

func (ms *ModuleSvc) DeployModules(client *client.Client, containers *models.Containers, sidecarImage string, sidecarResources *container.Resources) (map[string]int, error) {
	deployedModules := make(map[string]int)
	networkConfig := helpers.NewModuleNetworkConfig()

	var sidecarWG sync.WaitGroup
	sidecarErrCh := make(chan error, 10)
	for _, registryModules := range containers.RegistryModules {
		for _, registryModule := range registryModules {
			if ms.shouldSkipModule(registryModule, containers.ManagementOnly) {
				continue
			}
			if !ms.shouldDeployModule(registryModule, containers.BackendModules) {
				continue
			}

			backendModule := containers.BackendModules[registryModule.Name]
			version := ms.GetModuleImageVersion(backendModule, registryModule)
			image := ms.GetModuleImage(version, registryModule)
			env := ms.GetModuleEnv(containers, registryModule, backendModule)
			container := models.NewModuleContainer(registryModule.Name, image, env, backendModule, networkConfig)
			if err := ms.DeployModule(client, container); err != nil {
				return nil, err
			}
			deployedModules[registryModule.Name] = backendModule.ModuleExposedServerPort

			if backendModule.DeploySidecar && sidecarImage != "" {
				sidecarWG.Add(1)
				go ms.deploySidecarAsync(&sidecarWG, sidecarErrCh, &SidecarRequest{
					Client:           client,
					Containers:       containers,
					RegistryModule:   registryModule,
					BackendModule:    backendModule,
					SidecarImage:     sidecarImage,
					NetworkConfig:    networkConfig,
					SidecarResources: sidecarResources,
				})
			}
		}
	}

	go func() {
		sidecarWG.Wait()
		close(sidecarErrCh)
	}()
	for err := range sidecarErrCh {
		return nil, err
	}

	return deployedModules, nil
}

func (ms *ModuleSvc) shouldSkipModule(registryModule *models.RegistryModule, managementOnly bool) bool {
	isManagementModule := strings.Contains(registryModule.Name, constant.ManagementModulePattern)
	return (managementOnly && !isManagementModule) || (!managementOnly && isManagementModule)
}

func (ms *ModuleSvc) shouldDeployModule(registryModule *models.RegistryModule, backendModules map[string]models.BackendModule) bool {
	backendModule, exists := backendModules[registryModule.Name]
	return exists && backendModule.DeployModule
}

func (ms *ModuleSvc) deploySidecarAsync(wg *sync.WaitGroup, errCh chan<- error, r *SidecarRequest) {
	defer wg.Done()

	env := ms.GetSidecarEnv(r.Containers, r.RegistryModule, r.BackendModule, "", "")
	container := models.NewSidecarContainer(r.RegistryModule.SidecarName, r.SidecarImage, env, r.BackendModule, r.NetworkConfig, r.SidecarResources)
	if err := ms.DeployModule(r.Client, container); err != nil {
		err := appErrors.SidecarDeployFailed(r.RegistryModule.SidecarName, err)
		select {
		case errCh <- err:
		default:
		}
	}
}

func (ms *ModuleSvc) DeployModule(client *client.Client, c *models.Container) error {
	ctx, cancel := context.WithTimeout(context.Background(), constant.ContextTimeoutDockerDeploy)
	defer cancel()

	containerName := ms.getContainerName(c)
	if c.PullImage {
		err := ms.PullModule(client, c.Image)
		if err != nil {
			return err
		}
	}

	createResponse, err := client.ContainerCreate(ctx, c.Config, c.HostConfig, c.NetworkConfig, c.Platform, containerName)
	if err != nil {
		return err
	}
	if len(createResponse.Warnings) > 0 {
		slog.Warn(ms.Action.Name, "text", "Module created with warning", "container", containerName, "warnings", createResponse.Warnings)
	}

	err = client.ContainerStart(ctx, createResponse.ID, container.StartOptions{})
	if err != nil {
		return err
	}
	slog.Info(ms.Action.Name, "text", "Deployed module", "module", containerName)

	return nil
}

func (ms *ModuleSvc) getContainerName(container *models.Container) string {
	if strings.HasPrefix(container.Name, constant.ManagementModulePattern) {
		return fmt.Sprintf("eureka-%s", container.Name)
	}

	return fmt.Sprintf("eureka-%s-%s", ms.Action.ConfigProfile, container.Name)
}

func (ms *ModuleSvc) UndeployModuleByNamePattern(client *client.Client, pattern string) error {
	deployedModules, err := ms.GetDeployedModules(client, filters.NewArgs(filters.KeyValuePair{
		Key:   "name",
		Value: pattern,
	}))
	if err != nil {
		return err
	}

	for _, deployedModule := range deployedModules {
		err = ms.undeployModule(client, deployedModule)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ms *ModuleSvc) undeployModule(client *client.Client, deployedModule container.Summary) error {
	ctx, cancel := context.WithTimeout(context.Background(), constant.ContextTimeoutDockerUndeploy)
	defer cancel()

	err := client.NetworkDisconnect(ctx, constant.NetworkID, deployedModule.ID, false)
	if err != nil {
		slog.Warn(ms.Action.Name, "text", "Module network is disconnected with warnings", "moduleId", deployedModule.ID, "error", err.Error())
	}

	err = client.ContainerStop(ctx, deployedModule.ID, container.StopOptions{
		Signal: "9",
	})
	if err != nil {
		return err
	}

	err = client.ContainerRemove(ctx, deployedModule.ID, container.RemoveOptions{
		Force:         true,
		RemoveVolumes: true,
	})
	if err != nil {
		slog.Error(ms.Action.Name, "error", err, "module", deployedModule.ID, "operation", "container remove")
	}
	containerName := strings.ReplaceAll(deployedModule.Names[0], "/", "")
	slog.Info(ms.Action.Name, "text", "Undeployed module", "module", containerName)

	return nil
}
