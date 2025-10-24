package action

import (
	"fmt"

	"github.com/folio-org/eureka-cli/actionparams"
)

// Action is a container that holds the state of the deployment
type Action struct {
	Name          string
	GatewayURL    string
	StartPort     int
	EndPort       int
	ReservedPorts []int
	Params        *actionparams.ActionParams
}

func New(name, gatewayURL string, runParams *actionparams.ActionParams) *Action {
	return newGeneric(name, gatewayURL, 30000, 30999, runParams)
}

func NewCustom(name, gatewayURL string, startPort, endPort int, runParams *actionparams.ActionParams) *Action {
	return newGeneric(name, gatewayURL, startPort, endPort, runParams)
}

func newGeneric(name, gatewayURL string, startPort, endPort int, runParams *actionparams.ActionParams) *Action {
	return &Action{
		Name:          name,
		GatewayURL:    gatewayURL,
		StartPort:     startPort,
		EndPort:       endPort,
		ReservedPorts: []int{},
		Params:        runParams,
	}
}

func (a *Action) CreateURL(port string, route string) string {
	return fmt.Sprintf(a.GatewayURL, port) + route
}
