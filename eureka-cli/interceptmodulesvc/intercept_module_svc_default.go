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
	ports, err := is.Action.GetPreReservedPortSet([]func() (int, error){
		func() (int, error) { return is.Action.GetPreReservedPort() },
		func() (int, error) { return is.Action.GetPreReservedPort() },
		func() (int, error) { return is.Action.GetPreReservedPort() },
		func() (int, error) { return is.Action.GetPreReservedPort() },
	})
	if err != nil {
		return err
	}
	moduleServerPort := ports[0]
	moduleDebugPort := ports[1]
	sidecarServerPort := ports[2]
	sidecarDebugPort := ports[3]

	is.pair.NetworkConfig = helpers.GetModuleNetworkConfig()
	is.pair.BackendModule, is.pair.Module = is.ModuleSvc.GetBackendModule(is.pair.Containers, is.pair.ModuleName)
	is.pair.BackendModule.ModuleExposedServerPort = moduleServerPort
	is.pair.BackendModule.ModuleExposedDebugPort = moduleDebugPort
	is.pair.BackendModule.SidecarExposedServerPort = sidecarServerPort
	is.pair.BackendModule.SidecarExposedDebugPort = sidecarDebugPort
	is.pair.BackendModule.ModulePortBindings = helpers.CreatePortBindings(moduleServerPort, moduleDebugPort, is.pair.BackendModule.PrivatePort)
	is.pair.BackendModule.SidecarPortBindings = helpers.CreatePortBindings(sidecarServerPort, sidecarDebugPort, is.pair.BackendModule.PrivatePort)
	if err := is.updateModuleDiscovery(); err != nil {
		return err
	}

	return nil
}
