package managementstep

import (
	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/httpclient"
	"github.com/folio-org/eureka-cli/tenantstep"
)

type ManagementStep struct {
	Action     *action.Action
	HTTPClient *httpclient.HTTPClient
	TenantStep *tenantstep.TenantStep
}

func New(action *action.Action, httpClient *httpclient.HTTPClient, tenantStep *tenantstep.TenantStep) *ManagementStep {
	return &ManagementStep{
		Action:     action,
		HTTPClient: httpClient,
		TenantStep: tenantStep,
	}
}
