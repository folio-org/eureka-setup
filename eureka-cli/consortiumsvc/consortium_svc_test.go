package consortiumsvc_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/folio-org/eureka-setup/eureka-cli/consortiumsvc"
	"github.com/folio-org/eureka-setup/eureka-cli/constant"
	"github.com/folio-org/eureka-setup/eureka-cli/field"
	"github.com/folio-org/eureka-setup/eureka-cli/internal/testhelpers"
	"github.com/folio-org/eureka-setup/eureka-cli/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserSvc is a mock for usersvc.UserProcessor
type MockUserSvc struct {
	mock.Mock
}

func (m *MockUserSvc) Get(tenantName string, username string) (*models.User, error) {
	args := m.Called(tenantName, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserSvc) Create(tenantName string, user models.User, username string) error {
	args := m.Called(tenantName, user, username)
	return args.Error(0)
}

func (m *MockUserSvc) Delete(tenantName string, username string) error {
	args := m.Called(tenantName, username)
	return args.Error(0)
}

func (m *MockUserSvc) GetAdminUser() *models.User {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*models.User)
}

func TestNew(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	mockHTTP := &testhelpers.MockHTTPClient{}
	mockUserSvc := &MockUserSvc{}

	// Act
	svc := consortiumsvc.New(action, mockHTTP, mockUserSvc)

	// Assert
	assert.NotNil(t, svc)
	assert.Equal(t, action, svc.Action)
	assert.Equal(t, mockHTTP, svc.HTTPClient)
	assert.Equal(t, mockUserSvc, svc.UserSvc)
}

func TestGetConsortiumByName_Success(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "test-token"
	mockUserSvc := &MockUserSvc{}
	svc := consortiumsvc.New(action, mockHTTP, mockUserSvc)

	centralTenant := "central-tenant"
	consortiumName := "test-consortium"

	expectedResponse := models.ConsortiumResponse{
		Consortia: []models.Consortium{
			{
				ID:   "consortium-123",
				Name: consortiumName,
			},
		},
	}

	mockHTTP.On("GetRetryReturnStruct",
		mock.MatchedBy(func(url string) bool {
			return strings.Contains(url, "/consortia") &&
				strings.Contains(url, "name=="+consortiumName)
		}),
		mock.MatchedBy(func(headers map[string]string) bool {
			return headers[constant.OkapiTenantHeader] == centralTenant &&
				headers[constant.OkapiTokenHeader] == "test-token" &&
				headers[constant.ContentTypeHeader] == constant.ApplicationJSON
		}),
		mock.AnythingOfType("*models.ConsortiumResponse")).
		Run(func(args mock.Arguments) {
			arg := args.Get(2).(*models.ConsortiumResponse)
			*arg = expectedResponse
		}).
		Return(nil)

	// Act
	result, err := svc.GetConsortiumByName(centralTenant, consortiumName)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	resultMap := result.(map[string]any)
	assert.Equal(t, "consortium-123", resultMap["id"])
	assert.Equal(t, consortiumName, resultMap["name"])
	mockHTTP.AssertExpectations(t)
}

func TestGetConsortiumByName_NotFound(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "test-token"
	mockUserSvc := &MockUserSvc{}
	svc := consortiumsvc.New(action, mockHTTP, mockUserSvc)

	centralTenant := "central-tenant"
	consortiumName := "nonexistent-consortium"

	emptyResponse := models.ConsortiumResponse{
		Consortia: []models.Consortium{},
	}

	mockHTTP.On("GetRetryReturnStruct",
		mock.Anything,
		mock.Anything,
		mock.AnythingOfType("*models.ConsortiumResponse")).
		Run(func(args mock.Arguments) {
			arg := args.Get(2).(*models.ConsortiumResponse)
			*arg = emptyResponse
		}).
		Return(nil)

	// Act
	result, err := svc.GetConsortiumByName(centralTenant, consortiumName)

	// Assert
	assert.NoError(t, err)
	assert.Nil(t, result)
	mockHTTP.AssertExpectations(t)
}

func TestGetConsortiumByName_HTTPError(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "test-token"
	mockUserSvc := &MockUserSvc{}
	svc := consortiumsvc.New(action, mockHTTP, mockUserSvc)

	centralTenant := "central-tenant"
	consortiumName := "test-consortium"

	mockHTTP.On("GetRetryReturnStruct",
		mock.Anything,
		mock.Anything,
		mock.AnythingOfType("*models.ConsortiumResponse")).
		Return(errors.New("network error"))

	// Act
	result, err := svc.GetConsortiumByName(centralTenant, consortiumName)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "network error", err.Error())
	mockHTTP.AssertExpectations(t)
}

func TestGetConsortiumByName_HeaderCreationError(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "" // Empty token
	mockUserSvc := &MockUserSvc{}
	svc := consortiumsvc.New(action, mockHTTP, mockUserSvc)

	centralTenant := "central-tenant"
	consortiumName := "test-consortium"

	// Act
	result, err := svc.GetConsortiumByName(centralTenant, consortiumName)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	mockHTTP.AssertNotCalled(t, "GetRetryReturnStruct", mock.Anything, mock.Anything, mock.Anything)
}

func TestGetConsortiumByName_BlankTenantName(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "test-token"
	mockUserSvc := &MockUserSvc{}
	svc := consortiumsvc.New(action, mockHTTP, mockUserSvc)

	centralTenant := "" // Empty tenant
	consortiumName := "test-consortium"

	// Act
	result, err := svc.GetConsortiumByName(centralTenant, consortiumName)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	mockHTTP.AssertNotCalled(t, "GetRetryReturnStruct", mock.Anything, mock.Anything, mock.Anything)
}

func TestGetConsortiumCentralTenant_Found(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	mockUserSvc := &MockUserSvc{}
	svc := consortiumsvc.New(action, mockHTTP, mockUserSvc)

	consortiumName := "test-consortium"
	action.ConfigTenants = map[string]any{
		"member-tenant": map[string]any{
			field.TenantsConsortiumEntry:    consortiumName,
			field.TenantsCentralTenantEntry: false,
		},
		"central-tenant": map[string]any{
			field.TenantsConsortiumEntry:    consortiumName,
			field.TenantsCentralTenantEntry: true,
		},
	}

	// Act
	result := svc.GetConsortiumCentralTenant(consortiumName)

	// Assert
	assert.Equal(t, "central-tenant", result)
}

func TestGetConsortiumCentralTenant_NotFound(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	mockUserSvc := &MockUserSvc{}
	svc := consortiumsvc.New(action, mockHTTP, mockUserSvc)

	consortiumName := "test-consortium"
	action.ConfigTenants = map[string]any{
		"member-tenant": map[string]any{
			field.TenantsConsortiumEntry:    consortiumName,
			field.TenantsCentralTenantEntry: false,
		},
	}

	// Act
	result := svc.GetConsortiumCentralTenant(consortiumName)

	// Assert
	assert.Empty(t, result)
}

func TestGetConsortiumCentralTenant_DifferentConsortium(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	mockUserSvc := &MockUserSvc{}
	svc := consortiumsvc.New(action, mockHTTP, mockUserSvc)

	consortiumName := "test-consortium"
	action.ConfigTenants = map[string]any{
		"central-tenant": map[string]any{
			field.TenantsConsortiumEntry:    "other-consortium",
			field.TenantsCentralTenantEntry: true,
		},
	}

	// Act
	result := svc.GetConsortiumCentralTenant(consortiumName)

	// Assert
	assert.Empty(t, result)
}

func TestGetConsortiumUsers_Success(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	mockUserSvc := &MockUserSvc{}
	svc := consortiumsvc.New(action, mockHTTP, mockUserSvc)

	consortiumName := "test-consortium"
	action.ConfigUsers = map[string]any{
		"user1": map[string]any{
			field.UsersConsortiumEntry: consortiumName,
		},
		"user2": map[string]any{
			field.UsersConsortiumEntry: consortiumName,
		},
		"user3": map[string]any{
			field.UsersConsortiumEntry: "other-consortium",
		},
	}

	// Act
	result := svc.GetConsortiumUsers(consortiumName)

	// Assert
	assert.Len(t, result, 2)
	assert.Contains(t, result, "user1")
	assert.Contains(t, result, "user2")
	assert.NotContains(t, result, "user3")
}

func TestGetConsortiumUsers_NoneFound(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	mockUserSvc := &MockUserSvc{}
	svc := consortiumsvc.New(action, mockHTTP, mockUserSvc)

	consortiumName := "test-consortium"
	action.ConfigUsers = map[string]any{
		"user1": map[string]any{
			field.UsersConsortiumEntry: "other-consortium",
		},
	}

	// Act
	result := svc.GetConsortiumUsers(consortiumName)

	// Assert
	assert.Empty(t, result)
}

func TestGetAdminUsername_Found(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	mockUserSvc := &MockUserSvc{}
	svc := consortiumsvc.New(action, mockHTTP, mockUserSvc)

	centralTenant := "central-tenant"
	consortiumUsers := map[string]any{
		"admin-user": map[string]any{
			field.UsersTenantEntry: centralTenant,
		},
		"member-user": map[string]any{
			field.UsersTenantEntry: "member-tenant",
		},
	}

	// Act
	result := svc.GetAdminUsername(centralTenant, consortiumUsers)

	// Assert
	assert.Equal(t, "admin-user", result)
}

func TestGetAdminUsername_NotFound(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	mockUserSvc := &MockUserSvc{}
	svc := consortiumsvc.New(action, mockHTTP, mockUserSvc)

	centralTenant := "central-tenant"
	consortiumUsers := map[string]any{
		"member-user": map[string]any{
			field.UsersTenantEntry: "member-tenant",
		},
	}

	// Act
	result := svc.GetAdminUsername(centralTenant, consortiumUsers)

	// Assert
	assert.Empty(t, result)
}

func TestGetAdminUsername_NoTenantField(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	mockUserSvc := &MockUserSvc{}
	svc := consortiumsvc.New(action, mockHTTP, mockUserSvc)

	centralTenant := "central-tenant"
	consortiumUsers := map[string]any{
		"user": map[string]any{},
	}

	// Act
	result := svc.GetAdminUsername(centralTenant, consortiumUsers)

	// Assert
	assert.Empty(t, result)
}

func TestCreateConsortium_Success(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "test-token"
	mockUserSvc := &MockUserSvc{}
	svc := consortiumsvc.New(action, mockHTTP, mockUserSvc)

	centralTenant := "central-tenant"
	consortiumName := "new-consortium"

	// Mock GetConsortiumByName returns nil (not found)
	emptyResponse := models.ConsortiumResponse{
		Consortia: []models.Consortium{},
	}

	mockHTTP.On("GetRetryReturnStruct",
		mock.Anything,
		mock.Anything,
		mock.AnythingOfType("*models.ConsortiumResponse")).
		Run(func(args mock.Arguments) {
			arg := args.Get(2).(*models.ConsortiumResponse)
			*arg = emptyResponse
		}).
		Return(nil)

	mockHTTP.On("PostReturnNoContent",
		mock.MatchedBy(func(url string) bool {
			return strings.Contains(url, "/consortia")
		}),
		mock.MatchedBy(func(payload []byte) bool {
			var data map[string]any
			_ = json.Unmarshal(payload, &data)
			return data["name"] == consortiumName && data["id"] != nil
		}),
		mock.MatchedBy(func(headers map[string]string) bool {
			return headers[constant.OkapiTenantHeader] == centralTenant
		})).
		Return(nil)

	// Act
	result, err := svc.CreateConsortium(centralTenant, consortiumName)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	_, uuidErr := uuid.Parse(result)
	assert.NoError(t, uuidErr)
	mockHTTP.AssertExpectations(t)
}

func TestCreateConsortium_AlreadyExists(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "test-token"
	mockUserSvc := &MockUserSvc{}
	svc := consortiumsvc.New(action, mockHTTP, mockUserSvc)

	centralTenant := "central-tenant"
	consortiumName := "existing-consortium"
	existingID := "existing-123"

	// Mock GetConsortiumByName returns existing consortium
	existingResponse := models.ConsortiumResponse{
		Consortia: []models.Consortium{
			{
				ID:   existingID,
				Name: consortiumName,
			},
		},
	}

	mockHTTP.On("GetRetryReturnStruct",
		mock.Anything,
		mock.Anything,
		mock.AnythingOfType("*models.ConsortiumResponse")).
		Run(func(args mock.Arguments) {
			arg := args.Get(2).(*models.ConsortiumResponse)
			*arg = existingResponse
		}).
		Return(nil)

	// Act
	result, err := svc.CreateConsortium(centralTenant, consortiumName)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, existingID, result)
	mockHTTP.AssertExpectations(t)
	// Verify PostReturnNoContent was NOT called
	mockHTTP.AssertNotCalled(t, "PostReturnNoContent", mock.Anything, mock.Anything, mock.Anything)
}

func TestCreateConsortium_GetError(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "test-token"
	mockUserSvc := &MockUserSvc{}
	svc := consortiumsvc.New(action, mockHTTP, mockUserSvc)

	centralTenant := "central-tenant"
	consortiumName := "test-consortium"

	mockHTTP.On("GetRetryReturnStruct",
		mock.Anything,
		mock.Anything,
		mock.AnythingOfType("*models.ConsortiumResponse")).
		Return(errors.New("get failed"))

	// Act
	result, err := svc.CreateConsortium(centralTenant, consortiumName)

	// Assert
	assert.Error(t, err)
	assert.Empty(t, result)
	assert.Equal(t, "get failed", err.Error())
	mockHTTP.AssertExpectations(t)
}

func TestCreateConsortium_PostError(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "test-token"
	mockUserSvc := &MockUserSvc{}
	svc := consortiumsvc.New(action, mockHTTP, mockUserSvc)

	centralTenant := "central-tenant"
	consortiumName := "new-consortium"

	emptyResponse := models.ConsortiumResponse{
		Consortia: []models.Consortium{},
	}

	mockHTTP.On("GetRetryReturnStruct",
		mock.Anything,
		mock.Anything,
		mock.AnythingOfType("*models.ConsortiumResponse")).
		Run(func(args mock.Arguments) {
			arg := args.Get(2).(*models.ConsortiumResponse)
			*arg = emptyResponse
		}).
		Return(nil)

	mockHTTP.On("PostReturnNoContent",
		mock.Anything,
		mock.Anything,
		mock.Anything).
		Return(errors.New("post failed"))

	// Act
	result, err := svc.CreateConsortium(centralTenant, consortiumName)

	// Assert
	assert.Error(t, err)
	assert.Empty(t, result)
	assert.Equal(t, "post failed", err.Error())
	mockHTTP.AssertExpectations(t)
}

func TestCreateConsortium_HeaderCreationError(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "" // Empty token
	mockUserSvc := &MockUserSvc{}
	svc := consortiumsvc.New(action, mockHTTP, mockUserSvc)

	centralTenant := "central-tenant"
	consortiumName := "new-consortium"

	// Act
	result, err := svc.CreateConsortium(centralTenant, consortiumName)

	// Assert
	assert.Error(t, err)
	assert.Empty(t, result)
	mockHTTP.AssertNotCalled(t, "GetRetryReturnStruct", mock.Anything, mock.Anything, mock.Anything)
	mockHTTP.AssertNotCalled(t, "PostReturnNoContent", mock.Anything, mock.Anything, mock.Anything)
}

func TestCreateConsortium_BlankTenantName(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "test-token"
	mockUserSvc := &MockUserSvc{}
	svc := consortiumsvc.New(action, mockHTTP, mockUserSvc)

	centralTenant := "" // Empty tenant
	consortiumName := "new-consortium"

	// Act
	result, err := svc.CreateConsortium(centralTenant, consortiumName)

	// Assert
	assert.Error(t, err)
	assert.Empty(t, result)
	mockHTTP.AssertNotCalled(t, "GetRetryReturnStruct", mock.Anything, mock.Anything, mock.Anything)
	mockHTTP.AssertNotCalled(t, "PostReturnNoContent", mock.Anything, mock.Anything, mock.Anything)
}

func TestGetSortedConsortiumTenants_Success(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	mockUserSvc := &MockUserSvc{}
	svc := consortiumsvc.New(action, mockHTTP, mockUserSvc)

	consortiumName := "test-consortium"
	action.ConfigTenants = map[string]any{
		"member1": map[string]any{
			field.TenantsConsortiumEntry:    consortiumName,
			field.TenantsCentralTenantEntry: false,
		},
		"central": map[string]any{
			field.TenantsConsortiumEntry:    consortiumName,
			field.TenantsCentralTenantEntry: true,
		},
		"member2": map[string]any{
			field.TenantsConsortiumEntry:    consortiumName,
			field.TenantsCentralTenantEntry: false,
		},
		"other-tenant": map[string]any{
			field.TenantsConsortiumEntry:    "other-consortium",
			field.TenantsCentralTenantEntry: false,
		},
	}

	// Act
	result := svc.GetSortedConsortiumTenants(consortiumName)

	// Assert
	assert.Len(t, result, 3)
	// Central tenant should be first
	assert.Equal(t, "central", result[0].Name)
	assert.Equal(t, 1, result[0].IsCentral)
	// Member tenants should follow
	assert.Equal(t, 0, result[1].IsCentral)
	assert.Equal(t, 0, result[2].IsCentral)
}

func TestGetSortedConsortiumTenants_EmptyConsortium(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	mockUserSvc := &MockUserSvc{}
	svc := consortiumsvc.New(action, mockHTTP, mockUserSvc)

	consortiumName := "empty-consortium"
	action.ConfigTenants = map[string]any{
		"other-tenant": map[string]any{
			field.TenantsConsortiumEntry: "other-consortium",
		},
	}

	// Act
	result := svc.GetSortedConsortiumTenants(consortiumName)

	// Assert
	assert.Empty(t, result)
}

func TestGetSortedConsortiumTenants_OnlyCentralTenant(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	mockUserSvc := &MockUserSvc{}
	svc := consortiumsvc.New(action, mockHTTP, mockUserSvc)

	consortiumName := "test-consortium"
	action.ConfigTenants = map[string]any{
		"central-tenant": map[string]any{
			field.TenantsConsortiumEntry:    consortiumName,
			field.TenantsCentralTenantEntry: true,
		},
	}

	// Act
	result := svc.GetSortedConsortiumTenants(consortiumName)

	// Assert
	assert.Len(t, result, 1)
	assert.Equal(t, "central-tenant", result[0].Name)
	assert.Equal(t, 1, result[0].IsCentral)
}

func TestGetSortedConsortiumTenants_OnlyMemberTenants(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	mockUserSvc := &MockUserSvc{}
	svc := consortiumsvc.New(action, mockHTTP, mockUserSvc)

	consortiumName := "test-consortium"
	action.ConfigTenants = map[string]any{
		"member1": map[string]any{
			field.TenantsConsortiumEntry:    consortiumName,
			field.TenantsCentralTenantEntry: false,
		},
		"member2": map[string]any{
			field.TenantsConsortiumEntry:    consortiumName,
			field.TenantsCentralTenantEntry: false,
		},
	}

	// Act
	result := svc.GetSortedConsortiumTenants(consortiumName)

	// Assert
	assert.Len(t, result, 2)
	assert.Equal(t, 0, result[0].IsCentral)
	assert.Equal(t, 0, result[1].IsCentral)
}

func TestGetSortedConsortiumTenants_NilPropertiesHandled(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	mockUserSvc := &MockUserSvc{}
	svc := consortiumsvc.New(action, mockHTTP, mockUserSvc)

	consortiumName := "test-consortium"
	action.ConfigTenants = map[string]any{
		"tenant-with-properties": map[string]any{
			field.TenantsConsortiumEntry:    consortiumName,
			field.TenantsCentralTenantEntry: true,
		},
		"tenant-nil": map[string]any{
			field.TenantsConsortiumEntry: consortiumName,
		},
	}

	// Act
	result := svc.GetSortedConsortiumTenants(consortiumName)

	// Assert
	assert.Len(t, result, 2)
	// Central tenant should be first
	assert.Equal(t, "tenant-with-properties", result[0].Name)
	assert.Equal(t, 1, result[0].IsCentral)
	// Tenant without central property should have IsCentral = 0
	assert.Equal(t, "tenant-nil", result[1].Name)
	assert.Equal(t, 0, result[1].IsCentral)
}

func TestGetSortedConsortiumTenants_MixedConsortiums(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	mockUserSvc := &MockUserSvc{}
	svc := consortiumsvc.New(action, mockHTTP, mockUserSvc)

	consortiumName := "test-consortium"
	action.ConfigTenants = map[string]any{
		"tenant-in-consortium": map[string]any{
			field.TenantsConsortiumEntry:    consortiumName,
			field.TenantsCentralTenantEntry: true,
		},
		"tenant-different-consortium": map[string]any{
			field.TenantsConsortiumEntry:    "different-consortium",
			field.TenantsCentralTenantEntry: false,
		},
		"tenant-no-consortium": map[string]any{
			field.TenantsCentralTenantEntry: false,
		},
	}

	// Act
	result := svc.GetSortedConsortiumTenants(consortiumName)

	// Assert
	// Only tenant-in-consortium should be included
	assert.Len(t, result, 1)
	assert.Equal(t, "tenant-in-consortium", result[0].Name)
	assert.Equal(t, 1, result[0].IsCentral)
}

func TestGetSortedConsortiumTenants_SortingOrder(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	mockUserSvc := &MockUserSvc{}
	svc := consortiumsvc.New(action, mockHTTP, mockUserSvc)

	consortiumName := "test-consortium"
	action.ConfigTenants = map[string]any{
		"member1": map[string]any{
			field.TenantsConsortiumEntry:    consortiumName,
			field.TenantsCentralTenantEntry: false,
		},
		"central": map[string]any{
			field.TenantsConsortiumEntry:    consortiumName,
			field.TenantsCentralTenantEntry: true,
		},
		"member2": map[string]any{
			field.TenantsConsortiumEntry:    consortiumName,
			field.TenantsCentralTenantEntry: false,
		},
	}

	// Act
	result := svc.GetSortedConsortiumTenants(consortiumName)

	// Assert
	assert.Len(t, result, 3)
	// Central tenant should be first due to sorting by IsCentral descending
	assert.Equal(t, "central", result[0].Name)
	assert.Equal(t, 1, result[0].IsCentral)
	// Member tenants follow (order between members may vary)
	assert.Equal(t, 0, result[1].IsCentral)
	assert.Equal(t, 0, result[2].IsCentral)
}

func TestGetSortedConsortiumTenants_SkipsNilProperties(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	mockUserSvc := &MockUserSvc{}
	svc := consortiumsvc.New(action, mockHTTP, mockUserSvc)

	consortiumName := "test-consortium"
	action.ConfigTenants = map[string]any{
		"tenant-with-nil-properties": nil,
		"tenant-valid": map[string]any{
			field.TenantsConsortiumEntry:    consortiumName,
			field.TenantsCentralTenantEntry: true,
		},
	}

	// Act
	result := svc.GetSortedConsortiumTenants(consortiumName)

	// Assert
	// Should only return tenant-valid, nil properties are skipped
	assert.Len(t, result, 1)
	assert.Equal(t, "tenant-valid", result[0].Name)
	assert.Equal(t, 1, result[0].IsCentral)
}

func TestCreateConsortiumTenants_CentralTenant(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "test-token"
	mockUserSvc := &MockUserSvc{}
	svc := consortiumsvc.New(action, mockHTTP, mockUserSvc)

	centralTenant := "central-tenant"
	consortiumID := "consortium-123"
	adminUsername := "admin-user"

	consortiumTenants := models.SortedConsortiumTenants{
		&models.SortedConsortiumTenant{Name: centralTenant, IsCentral: 1},
	}

	// Mock getConsortiumTenantByIDAndName returns nil (not found)
	emptyTenantsResponse := models.ConsortiumTenantsResponse{
		Tenants: []models.ConsortiumTenant{},
	}

	mockHTTP.On("GetRetryReturnStruct",
		mock.MatchedBy(func(url string) bool {
			return strings.Contains(url, fmt.Sprintf("/consortia/%s/tenants", consortiumID))
		}),
		mock.Anything,
		mock.AnythingOfType("*models.ConsortiumTenantsResponse")).
		Run(func(args mock.Arguments) {
			arg := args.Get(2).(*models.ConsortiumTenantsResponse)
			*arg = emptyTenantsResponse
		}).
		Return(nil)

	mockHTTP.On("PostReturnNoContent",
		mock.MatchedBy(func(url string) bool {
			return strings.Contains(url, fmt.Sprintf("/consortia/%s/tenants", consortiumID)) &&
				!strings.Contains(url, "adminUserId")
		}),
		mock.MatchedBy(func(payload []byte) bool {
			var data map[string]any
			_ = json.Unmarshal(payload, &data)
			return data["id"] == centralTenant && data["isCentral"] == 1.0
		}),
		mock.Anything).
		Return(nil)

	// Mock checkConsortiumTenantStatus
	statusResponse := models.ConsortiumTenantStatus{
		SetupStatus: "COMPLETED",
		IsCentral:   true,
	}

	mockHTTP.On("GetRetryReturnStruct",
		mock.MatchedBy(func(url string) bool {
			return strings.Contains(url, fmt.Sprintf("/consortia/%s/tenants/%s", consortiumID, centralTenant))
		}),
		mock.Anything,
		mock.AnythingOfType("*models.ConsortiumTenantStatus")).
		Run(func(args mock.Arguments) {
			arg := args.Get(2).(*models.ConsortiumTenantStatus)
			*arg = statusResponse
		}).
		Return(nil)

	// Act
	err := svc.CreateConsortiumTenants(centralTenant, consortiumID, consortiumTenants, adminUsername)

	// Assert
	assert.NoError(t, err)
	mockHTTP.AssertExpectations(t)
}

func TestCreateConsortiumTenants_MemberTenant(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "test-token"
	mockUserSvc := &MockUserSvc{}
	svc := consortiumsvc.New(action, mockHTTP, mockUserSvc)

	centralTenant := "central-tenant"
	memberTenant := "member-tenant"
	consortiumID := "consortium-123"
	adminUsername := "admin-user"
	adminUserID := "admin-123"

	consortiumTenants := models.SortedConsortiumTenants{
		&models.SortedConsortiumTenant{Name: memberTenant, IsCentral: 0},
	}

	// Mock getConsortiumTenantByIDAndName returns nil (not found)
	emptyTenantsResponse := models.ConsortiumTenantsResponse{
		Tenants: []models.ConsortiumTenant{},
	}

	mockHTTP.On("GetRetryReturnStruct",
		mock.MatchedBy(func(url string) bool {
			return strings.Contains(url, fmt.Sprintf("/consortia/%s/tenants", consortiumID)) &&
				!strings.Contains(url, adminUsername) &&
				!strings.Contains(url, memberTenant)
		}),
		mock.Anything,
		mock.AnythingOfType("*models.ConsortiumTenantsResponse")).
		Run(func(args mock.Arguments) {
			arg := args.Get(2).(*models.ConsortiumTenantsResponse)
			*arg = emptyTenantsResponse
		}).
		Return(nil)

	// Mock UserSvc.Get
	mockUserSvc.On("Get", centralTenant, adminUsername).
		Return(&models.User{ID: adminUserID}, nil)

	mockHTTP.On("PostReturnNoContent",
		mock.MatchedBy(func(url string) bool {
			return strings.Contains(url, fmt.Sprintf("/consortia/%s/tenants", consortiumID)) &&
				strings.Contains(url, fmt.Sprintf("adminUserId=%s", adminUserID))
		}),
		mock.MatchedBy(func(payload []byte) bool {
			var data map[string]any
			_ = json.Unmarshal(payload, &data)
			return data["id"] == memberTenant && data["isCentral"] == 0.0
		}),
		mock.Anything).
		Return(nil)

	// Mock checkConsortiumTenantStatus
	statusResponse := models.ConsortiumTenantStatus{
		SetupStatus: "COMPLETED",
		IsCentral:   false,
	}

	mockHTTP.On("GetRetryReturnStruct",
		mock.MatchedBy(func(url string) bool {
			return strings.Contains(url, fmt.Sprintf("/consortia/%s/tenants/%s", consortiumID, memberTenant))
		}),
		mock.Anything,
		mock.AnythingOfType("*models.ConsortiumTenantStatus")).
		Run(func(args mock.Arguments) {
			arg := args.Get(2).(*models.ConsortiumTenantStatus)
			*arg = statusResponse
		}).
		Return(nil)

	// Act
	err := svc.CreateConsortiumTenants(centralTenant, consortiumID, consortiumTenants, adminUsername)

	// Assert
	assert.NoError(t, err)
	mockHTTP.AssertExpectations(t)
	mockUserSvc.AssertExpectations(t)
}

func TestCreateConsortiumTenants_AlreadyExists(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "test-token"
	mockUserSvc := &MockUserSvc{}
	svc := consortiumsvc.New(action, mockHTTP, mockUserSvc)

	centralTenant := "central-tenant"
	consortiumID := "consortium-123"
	adminUsername := "admin-user"

	consortiumTenants := models.SortedConsortiumTenants{
		&models.SortedConsortiumTenant{Name: centralTenant, IsCentral: 1},
	}

	// Mock getConsortiumTenantByIDAndName returns existing tenant
	existingTenantsResponse := models.ConsortiumTenantsResponse{
		Tenants: []models.ConsortiumTenant{
			{Name: centralTenant},
		},
	}

	mockHTTP.On("GetRetryReturnStruct",
		mock.Anything,
		mock.Anything,
		mock.AnythingOfType("*models.ConsortiumTenantsResponse")).
		Run(func(args mock.Arguments) {
			arg := args.Get(2).(*models.ConsortiumTenantsResponse)
			*arg = existingTenantsResponse
		}).
		Return(nil)

	// Act
	err := svc.CreateConsortiumTenants(centralTenant, consortiumID, consortiumTenants, adminUsername)

	// Assert
	assert.NoError(t, err)
	mockHTTP.AssertExpectations(t)
	// Verify PostReturnNoContent was NOT called
	mockHTTP.AssertNotCalled(t, "PostReturnNoContent", mock.Anything, mock.Anything, mock.Anything)
}

func TestCreateConsortiumTenants_GetTenantByIDError(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "test-token"
	mockUserSvc := &MockUserSvc{}
	svc := consortiumsvc.New(action, mockHTTP, mockUserSvc)

	centralTenant := "central-tenant"
	consortiumID := "consortium-123"
	adminUsername := "admin-user"

	consortiumTenants := models.SortedConsortiumTenants{
		&models.SortedConsortiumTenant{Name: centralTenant, IsCentral: 1},
	}

	mockHTTP.On("GetRetryReturnStruct",
		mock.Anything,
		mock.Anything,
		mock.AnythingOfType("*models.ConsortiumTenantsResponse")).
		Return(errors.New("get failed"))

	// Act
	err := svc.CreateConsortiumTenants(centralTenant, consortiumID, consortiumTenants, adminUsername)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "get failed", err.Error())
	mockHTTP.AssertExpectations(t)
}

func TestCreateConsortiumTenants_UserGetError(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "test-token"
	mockUserSvc := &MockUserSvc{}
	svc := consortiumsvc.New(action, mockHTTP, mockUserSvc)

	centralTenant := "central-tenant"
	memberTenant := "member-tenant"
	consortiumID := "consortium-123"
	adminUsername := "admin-user"

	consortiumTenants := models.SortedConsortiumTenants{
		&models.SortedConsortiumTenant{Name: memberTenant, IsCentral: 0},
	}

	emptyTenantsResponse := models.ConsortiumTenantsResponse{
		Tenants: []models.ConsortiumTenant{},
	}

	mockHTTP.On("GetRetryReturnStruct",
		mock.Anything,
		mock.Anything,
		mock.AnythingOfType("*models.ConsortiumTenantsResponse")).
		Run(func(args mock.Arguments) {
			arg := args.Get(2).(*models.ConsortiumTenantsResponse)
			*arg = emptyTenantsResponse
		}).
		Return(nil)

	mockUserSvc.On("Get", centralTenant, adminUsername).
		Return(nil, errors.New("user not found"))

	// Act
	err := svc.CreateConsortiumTenants(centralTenant, consortiumID, consortiumTenants, adminUsername)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "user not found", err.Error())
	mockHTTP.AssertExpectations(t)
	mockUserSvc.AssertExpectations(t)
}

func TestCreateConsortiumTenants_PostError(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "test-token"
	mockUserSvc := &MockUserSvc{}
	svc := consortiumsvc.New(action, mockHTTP, mockUserSvc)

	centralTenant := "central-tenant"
	consortiumID := "consortium-123"
	adminUsername := "admin-user"

	consortiumTenants := models.SortedConsortiumTenants{
		&models.SortedConsortiumTenant{Name: centralTenant, IsCentral: 1},
	}

	emptyTenantsResponse := models.ConsortiumTenantsResponse{
		Tenants: []models.ConsortiumTenant{},
	}

	mockHTTP.On("GetRetryReturnStruct",
		mock.Anything,
		mock.Anything,
		mock.AnythingOfType("*models.ConsortiumTenantsResponse")).
		Run(func(args mock.Arguments) {
			arg := args.Get(2).(*models.ConsortiumTenantsResponse)
			*arg = emptyTenantsResponse
		}).
		Return(nil)

	mockHTTP.On("PostReturnNoContent",
		mock.Anything,
		mock.Anything,
		mock.Anything).
		Return(errors.New("post failed"))

	// Act
	err := svc.CreateConsortiumTenants(centralTenant, consortiumID, consortiumTenants, adminUsername)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "post failed", err.Error())
	mockHTTP.AssertExpectations(t)
}

func TestCreateConsortiumTenants_StatusCheckError(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "test-token"
	mockUserSvc := &MockUserSvc{}
	svc := consortiumsvc.New(action, mockHTTP, mockUserSvc)

	centralTenant := "central-tenant"
	consortiumID := "consortium-123"
	adminUsername := "admin-user"

	consortiumTenants := models.SortedConsortiumTenants{
		&models.SortedConsortiumTenant{Name: centralTenant, IsCentral: 1},
	}

	emptyTenantsResponse := models.ConsortiumTenantsResponse{
		Tenants: []models.ConsortiumTenant{},
	}

	mockHTTP.On("GetRetryReturnStruct",
		mock.MatchedBy(func(url string) bool {
			return !strings.Contains(url, fmt.Sprintf("/%s", centralTenant))
		}),
		mock.Anything,
		mock.AnythingOfType("*models.ConsortiumTenantsResponse")).
		Run(func(args mock.Arguments) {
			arg := args.Get(2).(*models.ConsortiumTenantsResponse)
			*arg = emptyTenantsResponse
		}).
		Return(nil)

	mockHTTP.On("PostReturnNoContent",
		mock.Anything,
		mock.Anything,
		mock.Anything).
		Return(nil)

	mockHTTP.On("GetRetryReturnStruct",
		mock.MatchedBy(func(url string) bool {
			return strings.Contains(url, fmt.Sprintf("/%s", centralTenant))
		}),
		mock.Anything,
		mock.AnythingOfType("*models.ConsortiumTenantStatus")).
		Return(errors.New("status check failed"))

	// Act
	err := svc.CreateConsortiumTenants(centralTenant, consortiumID, consortiumTenants, adminUsername)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "status check failed", err.Error())
	mockHTTP.AssertExpectations(t)
}

func TestCreateConsortiumTenants_HeaderCreationError(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "" // Empty token
	mockUserSvc := &MockUserSvc{}
	svc := consortiumsvc.New(action, mockHTTP, mockUserSvc)

	centralTenant := "central-tenant"
	consortiumID := "consortium-123"
	adminUsername := "admin-user"

	consortiumTenants := models.SortedConsortiumTenants{
		&models.SortedConsortiumTenant{Name: centralTenant, IsCentral: 1},
	}

	// Act
	err := svc.CreateConsortiumTenants(centralTenant, consortiumID, consortiumTenants, adminUsername)

	// Assert
	assert.Error(t, err)
	mockHTTP.AssertNotCalled(t, "GetRetryReturnStruct", mock.Anything, mock.Anything, mock.Anything)
	mockHTTP.AssertNotCalled(t, "PostReturnNoContent", mock.Anything, mock.Anything, mock.Anything)
}

func TestCreateConsortiumTenants_BlankTenantName(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "test-token"
	mockUserSvc := &MockUserSvc{}
	svc := consortiumsvc.New(action, mockHTTP, mockUserSvc)

	centralTenant := "" // Empty tenant
	consortiumID := "consortium-123"
	adminUsername := "admin-user"

	consortiumTenants := models.SortedConsortiumTenants{
		&models.SortedConsortiumTenant{Name: "test-tenant", IsCentral: 1},
	}

	// Act
	err := svc.CreateConsortiumTenants(centralTenant, consortiumID, consortiumTenants, adminUsername)

	// Assert
	assert.Error(t, err)
	mockHTTP.AssertNotCalled(t, "GetRetryReturnStruct", mock.Anything, mock.Anything, mock.Anything)
	mockHTTP.AssertNotCalled(t, "PostReturnNoContent", mock.Anything, mock.Anything, mock.Anything)
}

func TestEnableCentralOrdering_Success(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "test-token"
	mockUserSvc := &MockUserSvc{}
	svc := consortiumsvc.New(action, mockHTTP, mockUserSvc)

	centralTenant := "central-tenant"

	// Mock getEnableCentralOrderingByKey returns false (not enabled)
	emptyResponse := models.SettingsResponse{
		Settings: []models.Setting{},
	}

	mockHTTP.On("GetRetryReturnStruct",
		mock.MatchedBy(func(url string) bool {
			return strings.Contains(url, "/orders-storage/settings") &&
				strings.Contains(url, "ALLOW_ORDERING_WITH_AFFILIATED_LOCATIONS")
		}),
		mock.Anything,
		mock.AnythingOfType("*models.SettingsResponse")).
		Run(func(args mock.Arguments) {
			arg := args.Get(2).(*models.SettingsResponse)
			*arg = emptyResponse
		}).
		Return(nil)

	mockHTTP.On("PostReturnNoContent",
		mock.MatchedBy(func(url string) bool {
			return strings.Contains(url, "/orders-storage/settings")
		}),
		mock.MatchedBy(func(payload []byte) bool {
			var data map[string]any
			_ = json.Unmarshal(payload, &data)
			return data["key"] == "ALLOW_ORDERING_WITH_AFFILIATED_LOCATIONS" &&
				data["value"] == "true"
		}),
		mock.Anything).
		Return(nil)

	// Act
	err := svc.EnableCentralOrdering(centralTenant)

	// Assert
	assert.NoError(t, err)
	mockHTTP.AssertExpectations(t)
}

func TestEnableCentralOrdering_AlreadyEnabled(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "test-token"
	mockUserSvc := &MockUserSvc{}
	svc := consortiumsvc.New(action, mockHTTP, mockUserSvc)

	centralTenant := "central-tenant"

	// Mock getEnableCentralOrderingByKey returns true (already enabled)
	enabledResponse := models.SettingsResponse{
		Settings: []models.Setting{
			{
				Key:   "ALLOW_ORDERING_WITH_AFFILIATED_LOCATIONS",
				Value: "true",
			},
		},
	}

	mockHTTP.On("GetRetryReturnStruct",
		mock.Anything,
		mock.Anything,
		mock.AnythingOfType("*models.SettingsResponse")).
		Run(func(args mock.Arguments) {
			arg := args.Get(2).(*models.SettingsResponse)
			*arg = enabledResponse
		}).
		Return(nil)

	// Act
	err := svc.EnableCentralOrdering(centralTenant)

	// Assert
	assert.NoError(t, err)
	mockHTTP.AssertExpectations(t)
	// Verify PostReturnNoContent was NOT called
	mockHTTP.AssertNotCalled(t, "PostReturnNoContent", mock.Anything, mock.Anything, mock.Anything)
}

func TestEnableCentralOrdering_GetError(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "test-token"
	mockUserSvc := &MockUserSvc{}
	svc := consortiumsvc.New(action, mockHTTP, mockUserSvc)

	centralTenant := "central-tenant"

	mockHTTP.On("GetRetryReturnStruct",
		mock.Anything,
		mock.Anything,
		mock.AnythingOfType("*models.SettingsResponse")).
		Return(errors.New("get failed"))

	// Act
	err := svc.EnableCentralOrdering(centralTenant)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "get failed", err.Error())
	mockHTTP.AssertExpectations(t)
}

func TestEnableCentralOrdering_PostError(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "test-token"
	mockUserSvc := &MockUserSvc{}
	svc := consortiumsvc.New(action, mockHTTP, mockUserSvc)

	centralTenant := "central-tenant"

	emptyResponse := models.SettingsResponse{
		Settings: []models.Setting{},
	}

	mockHTTP.On("GetRetryReturnStruct",
		mock.Anything,
		mock.Anything,
		mock.AnythingOfType("*models.SettingsResponse")).
		Run(func(args mock.Arguments) {
			arg := args.Get(2).(*models.SettingsResponse)
			*arg = emptyResponse
		}).
		Return(nil)

	mockHTTP.On("PostReturnNoContent",
		mock.Anything,
		mock.Anything,
		mock.Anything).
		Return(errors.New("post failed"))

	// Act
	err := svc.EnableCentralOrdering(centralTenant)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "post failed", err.Error())
	mockHTTP.AssertExpectations(t)
}

func TestEnableCentralOrdering_HeaderCreationError(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "" // Empty token
	mockUserSvc := &MockUserSvc{}
	svc := consortiumsvc.New(action, mockHTTP, mockUserSvc)

	centralTenant := "central-tenant"

	// Act
	err := svc.EnableCentralOrdering(centralTenant)

	// Assert
	assert.Error(t, err)
	mockHTTP.AssertNotCalled(t, "GetRetryReturnStruct", mock.Anything, mock.Anything, mock.Anything)
	mockHTTP.AssertNotCalled(t, "PostReturnNoContent", mock.Anything, mock.Anything, mock.Anything)
}

func TestEnableCentralOrdering_BlankTenantName(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "test-token"
	mockUserSvc := &MockUserSvc{}
	svc := consortiumsvc.New(action, mockHTTP, mockUserSvc)

	centralTenant := "" // Empty tenant

	// Act
	err := svc.EnableCentralOrdering(centralTenant)

	// Assert
	assert.Error(t, err)
	mockHTTP.AssertNotCalled(t, "GetRetryReturnStruct", mock.Anything, mock.Anything, mock.Anything)
	mockHTTP.AssertNotCalled(t, "PostReturnNoContent", mock.Anything, mock.Anything, mock.Anything)
}

func TestConsortiumTenantString(t *testing.T) {
	// Arrange & Act
	central := models.SortedConsortiumTenant{
		Name:      "central-tenant",
		IsCentral: 1,
	}
	member := models.SortedConsortiumTenant{
		Name:      "member-tenant",
		IsCentral: 0,
	}

	// Assert
	assert.Equal(t, "central-tenant (central)", central.String())
	assert.Equal(t, "member-tenant", member.String())
}

func TestConsortiumTenantsString(t *testing.T) {
	// Arrange
	tenants := models.SortedConsortiumTenants{
		&models.SortedConsortiumTenant{Name: "tenant1", IsCentral: 1},
		&models.SortedConsortiumTenant{Name: "tenant2", IsCentral: 0},
		&models.SortedConsortiumTenant{Name: "tenant3", IsCentral: 0},
	}

	// Act
	result := tenants.String()

	// Assert
	assert.Equal(t, "tenant1, tenant2, tenant3", result)
}

func TestConsortiumTenantsString_Single(t *testing.T) {
	// Arrange
	tenants := models.SortedConsortiumTenants{
		&models.SortedConsortiumTenant{Name: "tenant1", IsCentral: 1},
	}

	// Act
	result := tenants.String()

	// Assert
	assert.Equal(t, "tenant1", result)
}

func TestConsortiumTenantsString_Empty(t *testing.T) {
	// Arrange
	tenants := models.SortedConsortiumTenants{}

	// Act
	result := tenants.String()

	// Assert
	assert.Empty(t, result)
}
