package interceptmodulesvc

import (
	"log/slog"

	"github.com/docker/docker/client"
	"github.com/j011195/eureka-setup/eureka-cli/helpers"
	"github.com/j011195/eureka-setup/eureka-cli/modulesvc"
)

func (is *InterceptModuleSvc) DeployCustomSidecarForInterception(client *client.Client, pair *modulesvc.ModulePair) error {
	if err := is.ModuleSvc.UndeployModuleAndSidecarPair(client, pair); err != nil {
		return err
	}
	if err := is.prepareSidecarNetwork(pair); err != nil {
		return err
	}

	slog.Info(is.Action.Name, "text", "DEPLOYING CUSTOM SIDECAR FOR INTERCEPTION")
	if err := is.ModuleSvc.DeployCustomSidecar(client, pair); err != nil {
		return err
	}

	return is.ModuleSvc.CheckModuleAndSidecarReadiness(pair)
}

func (is *InterceptModuleSvc) prepareSidecarNetwork(pair *modulesvc.ModulePair) error {
	slog.Info(is.Action.Name, "text", "PREPARING SIDECAR NETWORK")
	moduleServerPort, err := helpers.GetPortFromURL(pair.ModuleURL)
	if err != nil {
		return err
	}

	sidecarServerPort, err := helpers.GetPortFromURL(pair.SidecarURL)
	if err != nil {
		return err
	}

	sidecarDebugPort, err := is.Action.GetPreReservedPort()
	if err != nil {
		return err
	}

	pair.BackendModule, pair.Module = is.ModuleSvc.GetBackendModule(pair.Containers, pair.ModuleName)
	pair.BackendModule.ModuleVersion = &pair.ModuleVersion
	pair.BackendModule.ModuleExposedServerPort = moduleServerPort
	pair.BackendModule.SidecarExposedServerPort = sidecarServerPort
	pair.BackendModule.SidecarExposedDebugPort = sidecarDebugPort

	pair.BackendModule.SidecarPortBindings = helpers.CreatePortBindings(sidecarServerPort, sidecarDebugPort, pair.BackendModule.PrivatePort)
	if err := is.updateModuleDiscovery(pair); err != nil {
		return err
	}
	pair.Module.Metadata.Version = pair.BackendModule.ModuleVersion

	return nil
}
