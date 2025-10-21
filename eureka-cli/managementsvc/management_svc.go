package managementsvc

import (
	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/httpclient"
	tenantsvc "github.com/folio-org/eureka-cli/tenantsvc"
)

type ManagementSvc struct {
	Action     *action.Action
	HTTPClient *httpclient.HTTPClient
	TenantSvc  *tenantsvc.TenantSvc
}

func New(action *action.Action, httpClient *httpclient.HTTPClient, tenantSvc *tenantsvc.TenantSvc) *ManagementSvc {
	return &ManagementSvc{
		Action:     action,
		HTTPClient: httpClient,
		TenantSvc:  tenantSvc,
	}
}
