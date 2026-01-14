package modulesvc

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/folio-org/eureka-setup/eureka-cli/constant"
	"github.com/folio-org/eureka-setup/eureka-cli/helpers"
	"github.com/folio-org/eureka-setup/eureka-cli/models"
)

// ModuleCustomizer defines the interface for managing custom module deployment and lifecycle
type ModuleCustomizer interface {
	UndeployModuleAndSidecarPair(client *client.Client, pair *ModulePair) error
	DeployCustomModule(client *client.Client, pair *ModulePair) error
	DeployCustomSidecar(client *client.Client, pair *ModulePair) error
	CheckModuleAndSidecarReadiness(pair *ModulePair) error
}

func (ms *ModuleSvc) UndeployModuleAndSidecarPair(client *client.Client, pair *ModulePair) error {
	slog.Info(ms.Action.Name, "text", "UNDEPLOYING MODULE AND SIDECAR PAIR")
	pattern := fmt.Sprintf(constant.SingleModuleOrSidecarContainerPattern, ms.Action.ConfigProfileName, ms.Action.Param.ModuleName)
	if err := ms.UndeployModuleByNamePattern(client, pattern); err != nil {
		return err
	}

	return nil
}

func (ms *ModuleSvc) DeployCustomModule(client *client.Client, pair *ModulePair) error {
	version := ms.GetModuleImageVersion(*pair.BackendModule, pair.Module)

	var imageName string
	if pair.Namespace != "" && !helpers.IsFolioNamespace(pair.Namespace) {
		imageName = ms.GetLocalModuleImage(pair.Namespace, pair.ModuleName, version)
	} else {
		imageName = ms.GetModuleImage(pair.Module, version)
	}
	slog.Info(ms.Action.Name, "text", "Using module image", "image", imageName)

	return ms.DeployModule(client, &models.Container{
		Name: pair.Module.Metadata.Name,
		Config: &container.Config{
			Image:        imageName,
			Hostname:     pair.Module.Metadata.Name,
			Env:          ms.GetModuleEnv(pair.Containers, pair.Module, *pair.BackendModule),
			ExposedPorts: *pair.BackendModule.ModuleExposedPorts,
		},
		HostConfig: &container.HostConfig{
			PortBindings:  *pair.BackendModule.ModulePortBindings,
			RestartPolicy: *helpers.GetRestartPolicy(),
			Resources:     pair.BackendModule.ModuleResources,
			Binds:         pair.BackendModule.ModuleVolumes,
		},
		NetworkConfig: helpers.GetModuleNetworkConfig(),
		Platform:      helpers.GetPlatform(),
		PullImage:     pair.BackendModule.LocalDescriptorPath == "",
	})
}

func (ms *ModuleSvc) DeployCustomSidecar(client *client.Client, pair *ModulePair) error {
	image, pullImage, err := ms.GetSidecarImage(pair.Containers.Modules.EurekaModules)
	if err != nil {
		return err
	}

	return ms.DeployModule(client, &models.Container{
		Name: pair.Module.Metadata.SidecarName,
		Config: &container.Config{
			Image:        image,
			Hostname:     pair.Module.Metadata.SidecarName,
			Env:          ms.GetSidecarEnv(pair.Containers, pair.Module, *pair.BackendModule, pair.ModuleURL, pair.SidecarURL),
			ExposedPorts: *pair.BackendModule.SidecarExposedPorts,
		},
		HostConfig: &container.HostConfig{
			PortBindings:  *pair.BackendModule.SidecarPortBindings,
			RestartPolicy: *helpers.GetRestartPolicy(),
			Resources:     *helpers.CreateResources(false, ms.Action.ConfigSidecarModuleResources),
		},
		NetworkConfig: helpers.GetModuleNetworkConfig(),
		Platform:      helpers.GetPlatform(),
		PullImage:     pullImage,
	})
}

func (ms *ModuleSvc) CheckModuleAndSidecarReadiness(pair *ModulePair) error {
	slog.Info(ms.Action.Name, "text", "WAITING FOR MODULE AND SIDECAR TO INITIALIZE", "port", pair.BackendModule.ModuleExposedServerPort)
	var interceptModuleWG sync.WaitGroup
	errCh := make(chan error, 2)

	interceptModuleWG.Add(2)
	go ms.CheckModuleReadiness(&interceptModuleWG, errCh, pair.ModuleName, pair.BackendModule.ModuleExposedServerPort)
	go ms.CheckModuleReadiness(&interceptModuleWG, errCh, helpers.GetSidecarName(pair.ModuleName), pair.BackendModule.SidecarExposedServerPort)
	interceptModuleWG.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			return err
		}
	}

	return nil
}
