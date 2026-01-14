package usersvc

import (
	"fmt"

	"github.com/folio-org/eureka-setup/eureka-cli/action"
	"github.com/folio-org/eureka-setup/eureka-cli/constant"
	"github.com/folio-org/eureka-setup/eureka-cli/errors"
	"github.com/folio-org/eureka-setup/eureka-cli/helpers"
	"github.com/folio-org/eureka-setup/eureka-cli/httpclient"
	"github.com/folio-org/eureka-setup/eureka-cli/models"
)

// UserProcessor defines the interface for user-related operations
type UserProcessor interface {
	Get(tenantName string, username string) (*models.User, error)
}

// UserSvc provides functionality for managing users
type UserSvc struct {
	Action     *action.Action
	HTTPClient httpclient.HTTPClientRunner
}

// New creates a new UserSvc instance
func New(action *action.Action, httpClient httpclient.HTTPClientRunner) *UserSvc {
	return &UserSvc{Action: action, HTTPClient: httpClient}
}

func (us *UserSvc) Get(tenantName string, username string) (*models.User, error) {
	requestURL := us.Action.GetRequestURL(constant.KongPort, fmt.Sprintf("/users?query=username==%s&limit=1", username))
	headers, err := helpers.SecureOkapiTenantApplicationJSONHeaders(tenantName, us.Action.KeycloakAccessToken)
	if err != nil {
		return nil, err
	}

	var decodedResponse models.UserResponse
	if err := us.HTTPClient.GetReturnStruct(requestURL, headers, &decodedResponse); err != nil {
		return nil, err
	}
	if len(decodedResponse.Users) == 0 {
		return nil, errors.UserNotFound(username, tenantName)
	}

	return &decodedResponse.Users[0], nil
}
