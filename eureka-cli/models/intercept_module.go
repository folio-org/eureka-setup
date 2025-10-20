package models

import (
	"strings"

	"github.com/docker/docker/api/types/network"
	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/helpers"
)

type InterceptModule struct {
	Id                string
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

func NewInterceptModule(
	action *action.Action,
	id string,
	defaultGateway bool,
	moduleUrl,
	sidecarUrl string,
	portStart,
	portEnd int,
) *InterceptModule {
	var moduleURLTemp, sidecarURLTemp = moduleUrl, sidecarUrl
	if defaultGateway {
		baseSchemaAndUrl := helpers.GetGatewayProtoAndBaseURL(action.Name)
		moduleURLTemp = helpers.ConstructURL(moduleUrl, baseSchemaAndUrl)
		sidecarURLTemp = helpers.ConstructURL(sidecarUrl, baseSchemaAndUrl)
	}

	id = strings.ReplaceAll(id, ":", "-")
	return &InterceptModule{
		Id:         id,
		ModuleName: helpers.GetModuleNameFromID(id),
		ModuleUrl:  &moduleURLTemp,
		SidecarUrl: &sidecarURLTemp,
		PortStart:  portStart,
		PortEnd:    portEnd,
	}
}

func (im *InterceptModule) ClearURLs() {
	im.ModuleUrl = nil
	im.SidecarUrl = nil
}
