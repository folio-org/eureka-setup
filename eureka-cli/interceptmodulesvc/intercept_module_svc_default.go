package interceptmodulesvc

import (
	"log/slog"

	"github.com/docker/docker/client"
	"github.com/j011195/eureka-setup/eureka-cli/helpers"
	"github.com/j011195/eureka-setup/eureka-cli/modulesvc"
)

func (is *InterceptModuleSvc) DeployDefaultModuleAndSidecarPair(client *client.Client, pair *modulesvc.ModulePair) error {
	pair.ClearModuleURL()
	pair.ClearSidecarURL()
	if err := is.ModuleSvc.UndeployModuleAndSidecarPair(client, pair); err != nil {
		return err
	}
	if err := is.prepareModuleAndSidecarPairNetwork(pair); err != nil {
		return err
	}

	slog.Info(is.Action.Name, "text", "DEPLOYING DEFAULT MODULE AND SIDECAR PAIR")
	if err := is.ModuleSvc.DeployCustomModule(client, pair); err != nil {
		return err
	}
	if err := is.ModuleSvc.DeployCustomSidecar(client, pair); err != nil {
		return err
	}

	return is.ModuleSvc.CheckModuleAndSidecarReadiness(pair)
}

func (is *InterceptModuleSvc) prepareModuleAndSidecarPairNetwork(pair *modulesvc.ModulePair) error {
	slog.Info(is.Action.Name, "text", "PREPARING MODULE AND SIDECAR PAIR NETWORK")
	ports, err := is.Action.GetPreReservedPortSet(4)
	if err != nil {
		return err
	}

	pair.BackendModule, pair.Module = is.ModuleSvc.GetBackendModule(pair.Containers, pair.ModuleName)
	pair.BackendModule.ModuleVersion = &pair.ModuleVersion
	pair.BackendModule.ModuleExposedServerPort = ports[0]
	pair.BackendModule.ModuleExposedDebugPort = ports[1]
	pair.BackendModule.SidecarExposedServerPort = ports[2]
	pair.BackendModule.SidecarExposedDebugPort = ports[3]

	pair.BackendModule.ModulePortBindings = helpers.CreatePortBindings(ports[0], ports[1], pair.BackendModule.PrivatePort)
	pair.BackendModule.SidecarPortBindings = helpers.CreatePortBindings(ports[2], ports[3], pair.BackendModule.PrivatePort)
	if err := is.updateModuleDiscovery(pair); err != nil {
		return err
	}
	pair.Module.Metadata.Version = pair.BackendModule.ModuleVersion

	return nil
}
