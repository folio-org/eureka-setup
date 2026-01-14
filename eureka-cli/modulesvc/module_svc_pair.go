package modulesvc

import (
	"github.com/docker/docker/api/types/network"
	"github.com/j011195/eureka-setup/eureka-cli/action"
	"github.com/j011195/eureka-setup/eureka-cli/helpers"
	"github.com/j011195/eureka-setup/eureka-cli/models"
)

// ModulePair represents a module configured for traffic interception and debugging
type ModulePair struct {
	ID            string
	ModuleName    string
	ModuleVersion string
	ModuleURL     string
	SidecarURL    string
	Namespace     string
	Module        *models.ProxyModule
	Containers    *models.Containers
	NetworkConfig *network.NetworkingConfig
	BackendModule *models.BackendModule
}

// NewModulePair creates a new ModulePair instance with configured URLs for interception
func NewModulePair(a *action.Action, p *action.Param) (*ModulePair, error) {
	var moduleURL, sidecarURL = p.ModuleURL, p.SidecarURL
	if p.DefaultGateway {
		gatewayURL, err := action.GetGatewayURL(a.Name)
		if err != nil {
			return nil, err
		}

		moduleURL = helpers.ConstructURL(p.ModuleURL, gatewayURL)
		sidecarURL = helpers.ConstructURL(p.SidecarURL, gatewayURL)
	}

	moduleVersion := helpers.GetModuleVersionFromID(p.ID)
	return &ModulePair{
		ID:            p.ID,
		ModuleName:    p.ModuleName,
		ModuleVersion: moduleVersion,
		ModuleURL:     moduleURL,
		SidecarURL:    sidecarURL,
		Namespace:     p.Namespace,
	}, nil
}

// ClearModuleURL clears the module URL from the intercept module
func (mp *ModulePair) ClearModuleURL() {
	mp.ModuleURL = ""
}

// ClearSidecarURL clears the sidecar URL from the intercept module
func (mp *ModulePair) ClearSidecarURL() {
	mp.SidecarURL = ""
}
