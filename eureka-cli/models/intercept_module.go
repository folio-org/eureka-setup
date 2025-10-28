package models

import (
	"strings"

	"github.com/docker/docker/api/types/network"
	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/helpers"
)

type InterceptModule struct {
	ID                string
	ModuleName        string
	ModuleUrl         *string
	SidecarUrl        *string
	SidecarServerPort int
	PortStart         int
	PortEnd           int
	RegistryModule    *RegistryModule
	Containers        *Containers
	NetworkConfig     *network.NetworkingConfig
	BackendModule     *BackendModule
}

func NewInterceptModule(a *action.Action, id string, defaultGateway bool, moduleURL, sidecarURL string) (*InterceptModule, error) {
	var moduleURLTemp, sidecarURLTemp = moduleURL, sidecarURL
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
		ModuleUrl:  &moduleURLTemp,
		SidecarUrl: &sidecarURLTemp,
		PortStart:  a.ConfigApplicationPortStart,
		PortEnd:    a.ConfigApplicationPortEnd,
	}, nil
}

func (im *InterceptModule) ClearURLs() {
	im.ModuleUrl = nil
	im.SidecarUrl = nil
}
