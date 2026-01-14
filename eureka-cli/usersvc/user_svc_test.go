package usersvc_test

import (
	"errors"
	"testing"

	"github.com/j011195/eureka-setup/eureka-cli/constant"
	"github.com/j011195/eureka-setup/eureka-cli/internal/testhelpers"
	"github.com/j011195/eureka-setup/eureka-cli/models"
	"github.com/j011195/eureka-setup/eureka-cli/usersvc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNew(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	mockHTTP := &testhelpers.MockHTTPClient{}

	// Act
	svc := usersvc.New(action, mockHTTP)

	// Assert
	assert.NotNil(t, svc)
	assert.Equal(t, action, svc.Action)
	assert.Equal(t, mockHTTP, svc.HTTPClient)
}

func TestGet_Success(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "test-token"
	svc := usersvc.New(action, mockHTTP)

	tenantName := "test-tenant"
	username := "testuser"

	expectedUser := models.User{
		ID:       "user-123",
		Username: username,
		Active:   true,
		Type:     "patron",
	}

	expectedResponse := models.UserResponse{
		Users:        []models.User{expectedUser},
		TotalRecords: 1,
	}

	mockHTTP.On("GetReturnStruct",
		mock.MatchedBy(func(url string) bool {
			return assert.Contains(t, url, "username=="+username) &&
				assert.Contains(t, url, "limit=1")
		}),
		mock.MatchedBy(func(headers map[string]string) bool {
			return headers[constant.OkapiTenantHeader] == tenantName &&
				headers[constant.OkapiTokenHeader] == action.KeycloakAccessToken &&
				headers[constant.ContentTypeHeader] == constant.ApplicationJSON
		}),
		mock.Anything).
		Run(func(args mock.Arguments) {
			target := args.Get(2).(*models.UserResponse)
			*target = expectedResponse
		}).
		Return(nil)

	// Act
	user, err := svc.Get(tenantName, username)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, expectedUser.ID, user.ID)
	assert.Equal(t, expectedUser.Username, user.Username)
	assert.Equal(t, expectedUser.Active, user.Active)
	assert.Equal(t, expectedUser.Type, user.Type)
	mockHTTP.AssertExpectations(t)
}

func TestGet_UserNotFound(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "test-token"
	svc := usersvc.New(action, mockHTTP)

	tenantName := "test-tenant"
	username := "nonexistent"

	emptyResponse := models.UserResponse{
		Users:        []models.User{},
		TotalRecords: 0,
	}

	mockHTTP.On("GetReturnStruct",
		mock.Anything,
		mock.Anything,
		mock.Anything).
		Run(func(args mock.Arguments) {
			target := args.Get(2).(*models.UserResponse)
			*target = emptyResponse
		}).
		Return(nil)

	// Act
	user, err := svc.Get(tenantName, username)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "resource not found")
	assert.Contains(t, err.Error(), username)
	assert.Contains(t, err.Error(), tenantName)
	mockHTTP.AssertExpectations(t)
}

func TestGet_HeaderCreationError(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "" // Empty token will cause header creation to fail
	svc := usersvc.New(action, mockHTTP)

	tenantName := "test-tenant"
	username := "testuser"

	// Act
	user, err := svc.Get(tenantName, username)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "access token")
	// HTTP client should not be called since header creation failed
	mockHTTP.AssertNotCalled(t, "GetReturnStruct")
}

func TestGet_HTTPError(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "test-token"
	svc := usersvc.New(action, mockHTTP)

	tenantName := "test-tenant"
	username := "testuser"
	expectedError := errors.New("HTTP request failed")

	mockHTTP.On("GetReturnStruct",
		mock.Anything,
		mock.Anything,
		mock.Anything).
		Return(expectedError)

	// Act
	user, err := svc.Get(tenantName, username)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, expectedError, err)
	mockHTTP.AssertExpectations(t)
}

func TestGet_WithPersonalInfo(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "test-token"
	svc := usersvc.New(action, mockHTTP)

	tenantName := "test-tenant"
	username := "testuser"

	expectedUser := models.User{
		ID:       "user-123",
		Username: username,
		Active:   true,
		Type:     "patron",
		Personal: &struct {
			FirstName              string `json:"firstName"`
			LastName               string `json:"lastName"`
			Email                  string `json:"email"`
			PreferredContactTypeId string `json:"preferredContactTypeId"`
		}{
			FirstName:              "John",
			LastName:               "Doe",
			Email:                  "john.doe@example.com",
			PreferredContactTypeId: "email",
		},
	}

	expectedResponse := models.UserResponse{
		Users:        []models.User{expectedUser},
		TotalRecords: 1,
	}

	mockHTTP.On("GetReturnStruct",
		mock.Anything,
		mock.Anything,
		mock.Anything).
		Run(func(args mock.Arguments) {
			target := args.Get(2).(*models.UserResponse)
			*target = expectedResponse
		}).
		Return(nil)

	// Act
	user, err := svc.Get(tenantName, username)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.NotNil(t, user.Personal)
	assert.Equal(t, "John", user.Personal.FirstName)
	assert.Equal(t, "Doe", user.Personal.LastName)
	assert.Equal(t, "john.doe@example.com", user.Personal.Email)
	mockHTTP.AssertExpectations(t)
}
