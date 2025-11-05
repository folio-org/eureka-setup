package interceptmodulesvc

import (
	"fmt"
	"log/slog"
	"sync"

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
	slog.Info(is.Action.Name, "text", "UPDATING MODULE DISCOVERY", "module", is.Action.Params.ModuleName, "id", is.Action.Params.ID, "port", is.pair.BackendModule.ModuleServerPort)
	if err := is.ManagementSvc.UpdateModuleDiscovery(is.Action.Params.ID, is.Action.Params.Restore, is.pair.BackendModule.ModuleServerPort, is.pair.SidecarURL); err != nil {
		return err
	}

	return nil
}

func (is *InterceptModuleSvc) undeployModuleAndSidecarPair() error {
	slog.Info(is.Action.Name, "text", "UNDEPLOYING DEFAULT MODULE AND SIDECAR PAIR")
	pattern := fmt.Sprintf(constant.SingleModuleOrSidecarContainerPattern, is.Action.ConfigProfile, is.Action.Params.ModuleName)
	if err := is.ModuleSvc.UndeployModuleByNamePattern(is.client, pattern); err != nil {
		return err
	}

	return nil
}

func (is *InterceptModuleSvc) deployModule() error {
	version := is.ModuleSvc.GetModuleImageVersion(*is.pair.BackendModule, is.pair.RegistryModule)
	image := is.ModuleSvc.GetModuleImage(version, is.pair.RegistryModule)
	env := is.ModuleSvc.GetModuleEnv(is.pair.Containers, is.pair.RegistryModule, *is.pair.BackendModule)
	container := models.NewModuleContainer(is.pair.RegistryModule.Name, image, env, *is.pair.BackendModule, is.pair.NetworkConfig)
	if err := is.ModuleSvc.DeployModule(is.client, container); err != nil {
		return err
	}

	return nil
}

func (is *InterceptModuleSvc) deploySidecar() error {
	image, pullImage, err := is.ModuleSvc.GetSidecarImage(is.pair.Containers.RegistryModules[constant.EurekaRegistry])
	if err != nil {
		return err
	}

	resources := helpers.CreateResources(false, is.Action.ConfigSidecarResources)
	env := is.ModuleSvc.GetSidecarEnv(is.pair.Containers, is.pair.RegistryModule, *is.pair.BackendModule, is.pair.ModuleURL, is.pair.SidecarURL)
	container := models.NewSidecarContainer(is.pair.RegistryModule.SidecarName, image, env, *is.pair.BackendModule, is.pair.NetworkConfig, resources)
	container.PullImage = pullImage
	if err := is.ModuleSvc.DeployModule(is.client, container); err != nil {
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
