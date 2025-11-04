package interceptmodulesvc

import (
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
}

// New creates a new InterceptSvc instance
func New(action *action.Action, ModuleSvc modulesvc.ModuleProcessor, managementSvc managementsvc.ManagementProcessor) *InterceptModuleSvc {
	return &InterceptModuleSvc{Action: action, ModuleSvc: ModuleSvc, ManagementSvc: managementSvc}
}

func (is *InterceptModuleSvc) DeployDefaultModuleAndSidecarPair(pair *ModulePair, client *client.Client) error {
	slog.Info(is.Action.Name, "text", "DEPLOYING DEFAULT MODULE AND SIDECAR PAIR")
	is.pair = pair
	if err := is.prepareModuleAndSidecarPairContainerNetwork(); err != nil {
		return err
	}

	is.pair.ClearModuleURL()
	if err := is.deployModule(client); err != nil {
		return err
	}

	is.pair.ClearSidecarURL()
	if err := is.deploySidecar(client); err != nil {
		return err
	}

	slog.Info(is.Action.Name, "text", "WAITING FOR MODULE TO INITIALIZE")
	var deployModuleWG sync.WaitGroup
	errCh := make(chan error, 1)

	deployModuleWG.Add(1)
	go is.ModuleSvc.CheckModuleReadiness(&deployModuleWG, errCh, is.pair.ModuleName, is.pair.BackendModule.ModuleExposedServerPort)
	deployModuleWG.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			return err
		}
	}

	return nil
}

func (is *InterceptModuleSvc) DeployCustomSidecarForInterception(pair *ModulePair, client *client.Client) error {
	slog.Info(is.Action.Name, "text", "DEPLOYING CUSTOM SIDECAR FOR INTERCEPTION")
	is.pair = pair
	if err := is.prepareSidecarContainerNetwork(); err != nil {
		return err
	}

	return is.deploySidecar(client)
}

func (is *InterceptModuleSvc) prepareModuleAndSidecarPairContainerNetwork() error {
	is.pair.NetworkConfig = helpers.NewModuleNetworkConfig()
	moduleServerPort, err := is.Action.GetPreReservedPort()
	if err != nil {
		return err
	}

	moduleDebugPort, err := is.Action.GetPreReservedPort()
	if err != nil {
		return err
	}

	sidecarServerPort, err := is.Action.GetPreReservedPort()
	if err != nil {
		return err
	}

	sidecarDebugPort, err := is.Action.GetPreReservedPort()
	if err != nil {
		return err
	}
	is.pair.BackendModule, is.pair.RegistryModule = is.ModuleSvc.GetBackendModule(is.pair.Containers, is.pair.ModuleName)
	is.pair.BackendModule.ModulePortBindings = helpers.CreatePortBindings(moduleServerPort, moduleDebugPort, is.pair.BackendModule.ModuleServerPort)
	is.pair.BackendModule.SidecarPortBindings = helpers.CreatePortBindings(sidecarServerPort, sidecarDebugPort, is.pair.BackendModule.ModuleServerPort)
	is.pair.BackendModule.ModuleExposedServerPort = moduleServerPort

	return nil
}

func (is *InterceptModuleSvc) prepareSidecarContainerNetwork() error {
	is.pair.NetworkConfig = helpers.NewModuleNetworkConfig()
	sidecarServerPort, err := helpers.ExtractPortFromURL(*is.pair.SidecarURL)
	if err != nil {
		return err
	}

	sidecarDebugPort, err := is.Action.GetPreReservedPort()
	if err != nil {
		return err
	}
	is.pair.SidecarServerPort = sidecarServerPort
	is.pair.BackendModule, is.pair.RegistryModule = is.ModuleSvc.GetBackendModule(is.pair.Containers, is.pair.ModuleName)
	is.pair.BackendModule.SidecarPortBindings = helpers.CreatePortBindings(sidecarServerPort, sidecarDebugPort, is.pair.BackendModule.ModuleServerPort)

	return nil
}

func (is *InterceptModuleSvc) deployModule(client *client.Client) error {
	version := is.ModuleSvc.GetModuleImageVersion(*is.pair.BackendModule, is.pair.RegistryModule)
	image := is.ModuleSvc.GetModuleImage(version, is.pair.RegistryModule)
	env := is.ModuleSvc.GetModuleEnv(is.pair.Containers, is.pair.RegistryModule, *is.pair.BackendModule)
	container := models.NewModuleContainer(is.pair.RegistryModule.Name, image, env, *is.pair.BackendModule, is.pair.NetworkConfig)
	if err := is.ModuleSvc.DeployModule(client, container); err != nil {
		return err
	}

	return nil
}

func (is *InterceptModuleSvc) deploySidecar(client *client.Client) error {
	image, pullImage, err := is.ModuleSvc.GetSidecarImage(is.pair.Containers.RegistryModules[constant.EurekaRegistry])
	if err != nil {
		return err
	}

	resources := helpers.CreateResources(false, is.Action.ConfigSidecarResources)
	env := is.ModuleSvc.GetSidecarEnv(is.pair.Containers, is.pair.RegistryModule, *is.pair.BackendModule, is.pair.ModuleURL, is.pair.SidecarURL)
	container := models.NewSidecarContainer(is.pair.RegistryModule.SidecarName, image, env, *is.pair.BackendModule, is.pair.NetworkConfig, resources)
	container.PullImage = pullImage
	if err := is.ModuleSvc.DeployModule(client, container); err != nil {
		return err
	}

	return nil
}
