package interceptmodulesvc

import (
	"github.com/docker/docker/api/types/network"
	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/folio-org/eureka-cli/models"
)

// ModulePair represents a module configured for traffic interception and debugging
type ModulePair struct {
	ID            string
	ModuleName    string
	ModuleURL     string
	SidecarURL    string
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

	return &ModulePair{ID: p.ID, ModuleName: p.ModuleName, ModuleURL: moduleURL, SidecarURL: sidecarURL}, nil
}

// ClearModuleURL clears the module URL from the intercept module
func (mp *ModulePair) clearModuleURL() {
	mp.ModuleURL = ""
}

// ClearSidecarURL clears the sidecar URL from the intercept module
func (mp *ModulePair) clearSidecarURL() {
	mp.SidecarURL = ""
}
