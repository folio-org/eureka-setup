package usersvc

import (
	"fmt"

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/httpclient"
)

type UserSvc struct {
	Action     *action.Action
	HTTPClient *httpclient.HTTPClient
}

func New(action *action.Action, httpClient *httpclient.HTTPClient) *UserSvc {
	return &UserSvc{
		Action:     action,
		HTTPClient: httpClient,
	}
}

func (us *UserSvc) GetUser(tenant string, accessToken string, username string) (any, error) {
	requestURL := us.Action.CreateURL(constant.KongPort, fmt.Sprintf("/users?query=username==%s", username))

	headers := map[string]string{
		constant.ContentTypeHeader: constant.ApplicationJSON,
		constant.OkapiTenantHeader: tenant,
		constant.OkapiTokenHeader:  accessToken,
	}

	data, err := us.HTTPClient.GetDecodeReturnMapStringAny(requestURL, headers)
	if err != nil {
		return nil, err
	}

	if data["users"] == nil || len(data["users"].([]any)) == 0 {
		return nil, nil
	}

	return data["users"].([]any)[0], nil
}
