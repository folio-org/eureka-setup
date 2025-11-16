package interceptmodulesvc

import (
	"log/slog"

	"github.com/docker/docker/client"
	"github.com/folio-org/eureka-cli/helpers"
)

func (is *InterceptModuleSvc) DeployDefaultModuleAndSidecarPair(pair *ModulePair, client *client.Client) error {
	is.pair = pair
	is.client = client
	is.pair.clearModuleURL()
	is.pair.clearSidecarURL()
	if err := is.undeployModuleAndSidecarPair(); err != nil {
		return err
	}
	if err := is.prepareModuleAndSidecarPairNetwork(); err != nil {
		return err
	}

	slog.Info(is.Action.Name, "text", "DEPLOYING DEFAULT MODULE AND SIDECAR PAIR")
	if err := is.deployModule(); err != nil {
		return err
	}
	if err := is.deploySidecar(); err != nil {
		return err
	}

	return is.checkModuleAndSidecarReadiness()
}

func (is *InterceptModuleSvc) prepareModuleAndSidecarPairNetwork() error {
	slog.Info(is.Action.Name, "text", "PREPARING MODULE AND SIDECAR PAIR NETWORK")
	ports, err := is.Action.GetPreReservedPortSet(4)
	if err != nil {
		return err
	}

	is.pair.BackendModule, is.pair.Module = is.ModuleSvc.GetBackendModule(is.pair.Containers, is.pair.ModuleName)
	is.pair.BackendModule.ModuleExposedServerPort = ports[0]
	is.pair.BackendModule.ModuleExposedDebugPort = ports[1]
	is.pair.BackendModule.SidecarExposedServerPort = ports[2]
	is.pair.BackendModule.SidecarExposedDebugPort = ports[3]

	is.pair.BackendModule.ModulePortBindings = helpers.CreatePortBindings(ports[0], ports[1], is.pair.BackendModule.PrivatePort)
	is.pair.BackendModule.SidecarPortBindings = helpers.CreatePortBindings(ports[2], ports[3], is.pair.BackendModule.PrivatePort)
	if err := is.updateModuleDiscovery(); err != nil {
		return err
	}

	return nil
}
