package interceptmodulesvc

import (
	"log/slog"

	"github.com/docker/docker/client"
	"github.com/folio-org/eureka-setup/eureka-cli/action"
	"github.com/folio-org/eureka-setup/eureka-cli/managementsvc"
	"github.com/folio-org/eureka-setup/eureka-cli/modulesvc"
)

// TODO Add testcontainers tests
// InterceptModuleProcessor defines the interface for module interception operations
type InterceptModuleProcessor interface {
	DeployDefaultModuleAndSidecarPair(client *client.Client, pair *modulesvc.ModulePair) error
	DeployCustomSidecarForInterception(client *client.Client, pair *modulesvc.ModulePair) error
}

// InterceptModuleSvc provides functionality for intercepting and redirecting module traffic
type InterceptModuleSvc struct {
	Action        *action.Action
	ModuleSvc     modulesvc.ModuleProcessor
	ManagementSvc managementsvc.ManagementProcessor
}

// New creates a new InterceptSvc instance
func New(action *action.Action, ModuleSvc modulesvc.ModuleProcessor, managementSvc managementsvc.ManagementProcessor) *InterceptModuleSvc {
	return &InterceptModuleSvc{Action: action, ModuleSvc: ModuleSvc, ManagementSvc: managementSvc}
}

func (is *InterceptModuleSvc) updateModuleDiscovery(pair *modulesvc.ModulePair) error {
	slog.Info(is.Action.Name, "text", "UPDATING MODULE DISCOVERY", "module", is.Action.Param.ModuleName, "id", is.Action.Param.ID, "port", pair.BackendModule.PrivatePort)
	err := is.ManagementSvc.UpdateModuleDiscovery(is.Action.Param.ID, is.Action.Param.Restore, pair.BackendModule.PrivatePort, pair.SidecarURL)
	if err != nil {
		return err
	}

	return nil
}
