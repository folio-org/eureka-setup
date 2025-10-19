package userstep

import (
	"fmt"

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/httpclient"
)

type UserStep struct {
	Action     *action.Action
	HTTPClient *httpclient.HTTPClient
}

func New(action *action.Action, httpClient *httpclient.HTTPClient) *UserStep {
	return &UserStep{
		Action:     action,
		HTTPClient: httpClient,
	}
}

func (us *UserStep) GetUser(panicOnError bool, tenant string, accessToken string, username string) any {
	requestURL := fmt.Sprintf(us.HTTPClient.GetGatewayURL(), constant.GatewayPort, fmt.Sprintf("/users?query=username==%s", username))

	headers := map[string]string{
		constant.ContentTypeHeader: constant.JsonContentType,
		constant.TenantHeader:      tenant,
		constant.TokenHeader:       accessToken,
	}

	data := us.HTTPClient.DoGetDecodeReturnMapStringAny(requestURL, panicOnError, headers)
	if data["users"] == nil || len(data["users"].([]any)) == 0 {
		return nil
	}

	return data["users"].([]any)[0]
}
