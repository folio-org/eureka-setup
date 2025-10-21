package consortiumsvc

import (
	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/httpclient"
	"github.com/folio-org/eureka-cli/usersvc"
)

type ConsortiumSvc struct {
	Action     *action.Action
	HTTPClient *httpclient.HTTPClient
	UserSvc    *usersvc.UserSvc
}

func New(action *action.Action, httpClient *httpclient.HTTPClient, userSvc *usersvc.UserSvc) *ConsortiumSvc {
	return &ConsortiumSvc{
		Action:     action,
		HTTPClient: httpClient,
		UserSvc:    userSvc,
	}
}
