package models

import (
	"strings"

	"github.com/docker/docker/api/types/network"
	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/helpers"
)

// InterceptModule represents a module configured for traffic interception and debugging
type InterceptModule struct {
	ID                string
	ModuleName        string
	ModuleURL         *string
	SidecarURL        *string
	SidecarServerPort int
	PortStart         int
	PortEnd           int
	RegistryModule    *RegistryModule
	Containers        *Containers
	NetworkConfig     *network.NetworkingConfig
	BackendModule     *BackendModule
}

// NewInterceptModule creates a new InterceptModule instance with configured URLs for interception
func NewInterceptModule(a *action.Action, id string, defaultGateway bool, moduleURL, sidecarURL string) (*InterceptModule, error) {
	var (
		moduleURLTemp  = moduleURL
		sidecarURLTemp = sidecarURL
	)
	if defaultGateway {
		gatewayURL, err := action.GetGatewayURL(a.Name)
		if err != nil {
			return nil, err
		}

		moduleURLTemp = helpers.ConstructURL(moduleURL, gatewayURL)
		sidecarURLTemp = helpers.ConstructURL(sidecarURL, gatewayURL)
	}

	id = strings.ReplaceAll(id, ":", "-")
	moduleName := helpers.GetModuleNameFromID(id)
	return &InterceptModule{
		ID:         id,
		ModuleName: moduleName,
		ModuleURL:  &moduleURLTemp,
		SidecarURL: &sidecarURLTemp,
		PortStart:  a.ConfigApplicationPortStart,
		PortEnd:    a.ConfigApplicationPortEnd,
	}, nil
}

// ClearURLs clears the module and sidecar URLs from the intercept module
func (im *InterceptModule) ClearURLs() {
	im.ModuleURL = nil
	im.SidecarURL = nil
}
