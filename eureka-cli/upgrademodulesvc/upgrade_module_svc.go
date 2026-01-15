package upgrademodulesvc

import (
	"log/slog"

	"github.com/docker/docker/client"
	"github.com/folio-org/eureka-setup/eureka-cli/action"
	"github.com/folio-org/eureka-setup/eureka-cli/execsvc"
	"github.com/folio-org/eureka-setup/eureka-cli/helpers"
	"github.com/folio-org/eureka-setup/eureka-cli/managementsvc"
	"github.com/folio-org/eureka-setup/eureka-cli/modulesvc"
)

// TODO Add testcontainers tests
// UpgradeModuleProcessor defines the interface combining all operations needed to upgrade a module
type UpgradeModuleProcessor interface {
	UpgradeModuleVersionManager
	UpgradeModuleBuildManager
	UpgradeModuleApplicationBuilder
	UpgradeModuleDeploymentManager
}

// UpgradeModuleDeploymentManager defines the interface for deploying upgraded modules and their sidecars
type UpgradeModuleDeploymentManager interface {
	DeployModuleAndSidecarPair(client *client.Client, pair *modulesvc.ModulePair) error
}

// UpgradeModuleSvc defines the service for upgrading or downgrading modules
type UpgradeModuleSvc struct {
	Action        *action.Action
	ExecSvc       execsvc.CommandRunner
	ModuleSvc     modulesvc.ModuleProcessor
	ManagementSvc managementsvc.ManagementProcessor
}

// New creates a new UpgradeModuleSvc instance
func New(action *action.Action, execSvc execsvc.CommandRunner, ModuleSvc modulesvc.ModuleProcessor, managementSvc managementsvc.ManagementProcessor) *UpgradeModuleSvc {
	return &UpgradeModuleSvc{Action: action, ExecSvc: execSvc, ModuleSvc: ModuleSvc, ManagementSvc: managementSvc}
}

func (um *UpgradeModuleSvc) DeployModuleAndSidecarPair(client *client.Client, pair *modulesvc.ModulePair) error {
	if err := um.ModuleSvc.UndeployModuleAndSidecarPair(client, pair); err != nil {
		return err
	}
	if err := um.prepareModuleAndSidecarPairNetwork(pair); err != nil {
		return err
	}

	slog.Info(um.Action.Name, "text", "DEPLOYING DEFAULT MODULE AND SIDECAR PAIR")
	if err := um.ModuleSvc.DeployCustomModule(client, pair); err != nil {
		return err
	}
	if err := um.ModuleSvc.DeployCustomSidecar(client, pair); err != nil {
		return err
	}

	return um.ModuleSvc.CheckModuleAndSidecarReadiness(pair)
}

func (um *UpgradeModuleSvc) prepareModuleAndSidecarPairNetwork(pair *modulesvc.ModulePair) error {
	slog.Info(um.Action.Name, "text", "PREPARING MODULE AND SIDECAR PAIR NETWORK")
	ports, err := um.Action.GetPreReservedPortSet(4)
	if err != nil {
		return err
	}

	pair.BackendModule, pair.Module = um.ModuleSvc.GetBackendModule(pair.Containers, pair.ModuleName)
	pair.BackendModule.ModuleVersion = &pair.ModuleVersion
	pair.BackendModule.ModuleExposedServerPort = ports[0]
	pair.BackendModule.ModuleExposedDebugPort = ports[1]
	pair.BackendModule.SidecarExposedServerPort = ports[2]
	pair.BackendModule.SidecarExposedDebugPort = ports[3]

	pair.BackendModule.ModulePortBindings = helpers.CreatePortBindings(ports[0], ports[1], pair.BackendModule.PrivatePort)
	pair.BackendModule.SidecarPortBindings = helpers.CreatePortBindings(ports[2], ports[3], pair.BackendModule.PrivatePort)

	pair.Module.Metadata.Version = pair.BackendModule.ModuleVersion

	return nil
}
