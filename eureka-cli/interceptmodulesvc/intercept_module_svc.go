package interceptmodulesvc

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/folio-org/eureka-cli/managementsvc"
	"github.com/folio-org/eureka-cli/models"
	"github.com/folio-org/eureka-cli/modulesvc"
)

// InterceptModuleProcessor defines the interface for module interception operations
type InterceptModuleProcessor interface {
	DeployDefaultModuleAndSidecarPair(pair *ModulePair, client *client.Client) error
	DeployCustomSidecarForInterception(pair *ModulePair, client *client.Client) error
}

// InterceptModuleSvc provides functionality for intercepting and redirecting module traffic
type InterceptModuleSvc struct {
	Action        *action.Action
	ModuleSvc     modulesvc.ModuleProcessor
	ManagementSvc managementsvc.ManagementProcessor
	pair          *ModulePair
	client        *client.Client
}

// New creates a new InterceptSvc instance
func New(action *action.Action, ModuleSvc modulesvc.ModuleProcessor, managementSvc managementsvc.ManagementProcessor) *InterceptModuleSvc {
	return &InterceptModuleSvc{Action: action, ModuleSvc: ModuleSvc, ManagementSvc: managementSvc}
}

func (is *InterceptModuleSvc) updateModuleDiscovery() error {
	slog.Info(is.Action.Name, "text", "UPDATING MODULE DISCOVERY", "module", is.Action.Param.ModuleName, "id", is.Action.Param.ID, "port", is.pair.BackendModule.PrivatePort)
	err := is.ManagementSvc.UpdateModuleDiscovery(is.Action.Param.ID, is.Action.Param.Restore, is.pair.BackendModule.PrivatePort, is.pair.SidecarURL)
	if err != nil {
		return err
	}

	return nil
}

func (is *InterceptModuleSvc) undeployModuleAndSidecarPair() error {
	slog.Info(is.Action.Name, "text", "UNDEPLOYING MODULE AND SIDECAR PAIR")
	pattern := fmt.Sprintf(constant.SingleModuleOrSidecarContainerPattern, is.Action.ConfigProfile, is.Action.Param.ModuleName)
	if err := is.ModuleSvc.UndeployModuleByNamePattern(is.client, pattern); err != nil {
		return err
	}

	return nil
}

func (is *InterceptModuleSvc) deployModule() error {
	version := is.ModuleSvc.GetModuleImageVersion(*is.pair.BackendModule, is.pair.Module)
	if err := is.ModuleSvc.DeployModule(is.client, &models.Container{
		Name: is.pair.Module.Metadata.Name,
		Config: &container.Config{
			Image:        is.ModuleSvc.GetModuleImage(version, is.pair.Module),
			Hostname:     is.pair.Module.Metadata.Name,
			Env:          is.ModuleSvc.GetModuleEnv(is.pair.Containers, is.pair.Module, *is.pair.BackendModule),
			ExposedPorts: *is.pair.BackendModule.ModuleExposedPorts,
		},
		HostConfig: &container.HostConfig{
			PortBindings:  *is.pair.BackendModule.ModulePortBindings,
			RestartPolicy: *helpers.GetRestartPolicy(),
			Resources:     is.pair.BackendModule.ModuleResources,
			Binds:         is.pair.BackendModule.ModuleVolumes,
		},
		NetworkConfig: helpers.GetModuleNetworkConfig(),
		Platform:      helpers.GetPlatform(),
		PullImage:     is.pair.BackendModule.LocalDescriptorPath == "",
	}); err != nil {
		return err
	}

	return nil
}

func (is *InterceptModuleSvc) deploySidecar() error {
	image, pullImage, err := is.ModuleSvc.GetSidecarImage(is.pair.Containers.Modules.EurekaModules)
	if err != nil {
		return err
	}

	if err := is.ModuleSvc.DeployModule(is.client, &models.Container{
		Name: is.pair.Module.Metadata.SidecarName,
		Config: &container.Config{
			Image:        image,
			Hostname:     is.pair.Module.Metadata.SidecarName,
			Env:          is.ModuleSvc.GetSidecarEnv(is.pair.Containers, is.pair.Module, *is.pair.BackendModule, is.pair.ModuleURL, is.pair.SidecarURL),
			ExposedPorts: *is.pair.BackendModule.SidecarExposedPorts,
		},
		HostConfig: &container.HostConfig{
			PortBindings:  *is.pair.BackendModule.SidecarPortBindings,
			RestartPolicy: *helpers.GetRestartPolicy(),
			Resources:     *helpers.CreateResources(false, is.Action.ConfigSidecarResources),
		},
		NetworkConfig: helpers.GetModuleNetworkConfig(),
		Platform:      helpers.GetPlatform(),
		PullImage:     pullImage,
	}); err != nil {
		return err
	}

	return nil
}

func (is *InterceptModuleSvc) checkModuleAndSidecarReadiness() error {
	slog.Info(is.Action.Name, "text", "WAITING FOR MODULE AND SIDECAR TO INITIALIZE", "port", is.pair.BackendModule.ModuleExposedServerPort)
	var interceptModuleWG sync.WaitGroup
	errCh := make(chan error, 2)

	interceptModuleWG.Add(2)
	go is.ModuleSvc.CheckModuleReadiness(&interceptModuleWG, errCh, is.pair.ModuleName, is.pair.BackendModule.ModuleExposedServerPort)
	go is.ModuleSvc.CheckModuleReadiness(&interceptModuleWG, errCh, helpers.GetSidecarName(is.pair.ModuleName), is.pair.BackendModule.SidecarExposedServerPort)
	interceptModuleWG.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			return err
		}
	}

	return nil
}
