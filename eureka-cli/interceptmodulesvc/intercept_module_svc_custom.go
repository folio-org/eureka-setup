package interceptmodulesvc

import (
	"log/slog"

	"github.com/docker/docker/client"
	"github.com/folio-org/eureka-cli/helpers"
)

func (is *InterceptModuleSvc) DeployCustomSidecarForInterception(pair *ModulePair, client *client.Client) error {
	is.pair = pair
	is.client = client
	if err := is.undeployModuleAndSidecarPair(); err != nil {
		return err
	}
	if err := is.prepareSidecarNetwork(); err != nil {
		return err
	}

	slog.Info(is.Action.Name, "text", "DEPLOYING CUSTOM SIDECAR FOR INTERCEPTION")
	if err := is.deploySidecar(); err != nil {
		return err
	}

	return is.checkModuleAndSidecarReadiness()
}

func (is *InterceptModuleSvc) prepareSidecarNetwork() error {
	slog.Info(is.Action.Name, "text", "PREPARING SIDECAR NETWORK")
	moduleServerPort, err := helpers.ExtractPortFromURL(is.pair.ModuleURL)
	if err != nil {
		return err
	}

	sidecarServerPort, err := helpers.ExtractPortFromURL(is.pair.SidecarURL)
	if err != nil {
		return err
	}

	sidecarDebugPort, err := is.Action.GetPreReservedPort()
	if err != nil {
		return err
	}

	is.pair.BackendModule, is.pair.Module = is.ModuleSvc.GetBackendModule(is.pair.Containers, is.pair.ModuleName)
	is.pair.BackendModule.ModuleExposedServerPort = moduleServerPort
	is.pair.BackendModule.SidecarExposedServerPort = sidecarServerPort
	is.pair.BackendModule.SidecarExposedDebugPort = sidecarDebugPort

	is.pair.BackendModule.SidecarPortBindings = helpers.CreatePortBindings(sidecarServerPort, sidecarDebugPort, is.pair.BackendModule.PrivatePort)
	if err := is.updateModuleDiscovery(); err != nil {
		return err
	}

	return nil
}
