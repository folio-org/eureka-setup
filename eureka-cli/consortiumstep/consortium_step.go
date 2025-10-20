package consortiumstep

import (
	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/httpclient"
	"github.com/folio-org/eureka-cli/userstep"
)

type ConsortiumStep struct {
	Action     *action.Action
	HTTPClient *httpclient.HTTPClient
	UserStep   *userstep.UserStep
}

func New(action *action.Action, httpClient *httpclient.HTTPClient, userStep *userstep.UserStep) *ConsortiumStep {
	return &ConsortiumStep{
		Action:     action,
		HTTPClient: httpClient,
		UserStep:   userStep,
	}
}
