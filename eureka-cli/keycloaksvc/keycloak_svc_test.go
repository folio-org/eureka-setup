package keycloaksvc_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"testing"

	"github.com/folio-org/eureka-setup/eureka-cli/constant"
	apperrors "github.com/folio-org/eureka-setup/eureka-cli/errors"
	"github.com/folio-org/eureka-setup/eureka-cli/internal/testhelpers"
	"github.com/folio-org/eureka-setup/eureka-cli/keycloaksvc"
	"github.com/folio-org/eureka-setup/eureka-cli/models"
	vault "github.com/hashicorp/vault-client-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockVaultClient is a mock for vaultclient.VaultClientRunner
type MockVaultClient struct {
	mock.Mock
}

func (m *MockVaultClient) Create() (*vault.Client, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*vault.Client), args.Error(1)
}

func (m *MockVaultClient) GetSecretKey(ctx context.Context, client *vault.Client, vaultRootToken string, secretPath string) (map[string]any, error) {
	args := m.Called(ctx, client, vaultRootToken, secretPath)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]any), args.Error(1)
}

// MockManagementSvc is a mock for managementsvc.ManagementProcessor
type MockManagementSvc struct {
	mock.Mock
}

func (m *MockManagementSvc) GetTenants(consortiumName string, tenantType constant.TenantType) ([]any, error) {
	args := m.Called(consortiumName, tenantType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]any), args.Error(1)
}

func (m *MockManagementSvc) CreateTenants() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockManagementSvc) RemoveTenants(consortiumName string, tenantType constant.TenantType) error {
	args := m.Called(consortiumName, tenantType)
	return args.Error(0)
}

func (m *MockManagementSvc) GetApplications() (models.ApplicationsResponse, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return models.ApplicationsResponse{}, args.Error(1)
	}
	return args.Get(0).(models.ApplicationsResponse), args.Error(1)
}

func (m *MockManagementSvc) GetLatestApplication() (map[string]any, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]any), args.Error(1)
}

func (m *MockManagementSvc) CreateApplication(extract *models.RegistryExtract) error {
	args := m.Called(extract)
	return args.Error(0)
}

func (m *MockManagementSvc) CreateNewApplication(r *models.ApplicationUpgradeRequest) error {
	args := m.Called(r)
	return args.Error(0)
}

func (m *MockManagementSvc) RemoveApplication(applicationID string) error {
	args := m.Called(applicationID)
	return args.Error(0)
}

func (m *MockManagementSvc) RemoveApplications(applicationName, ignoreAppID string) error {
	args := m.Called(applicationName, ignoreAppID)
	return args.Error(0)
}

func (m *MockManagementSvc) GetModuleDiscovery(name string) (models.ModuleDiscoveryResponse, error) {
	args := m.Called(name)
	return args.Get(0).(models.ModuleDiscoveryResponse), args.Error(1)
}

func (m *MockManagementSvc) CreateNewModuleDiscovery(newDiscoveryModules []map[string]string) error {
	args := m.Called(newDiscoveryModules)
	return args.Error(0)
}

func (m *MockManagementSvc) UpdateModuleDiscovery(id string, restore bool, privatePort int, sidecarURL string) error {
	args := m.Called(id, restore, privatePort, sidecarURL)
	return args.Error(0)
}

func (m *MockManagementSvc) GetTenantEntitlements(tenantName string, includeModules bool) (models.TenantEntitlementResponse, error) {
	args := m.Called(tenantName, includeModules)
	return args.Get(0).(models.TenantEntitlementResponse), args.Error(1)
}

func (m *MockManagementSvc) CreateTenantEntitlement(consortiumName string, tenantType constant.TenantType) error {
	args := m.Called(consortiumName, tenantType)
	return args.Error(0)
}

func (m *MockManagementSvc) UpgradeTenantEntitlement(consortiumName string, tenantType constant.TenantType, newApplicationID string) error {
	args := m.Called(consortiumName, tenantType, newApplicationID)
	return args.Error(0)
}

func (m *MockManagementSvc) RemoveTenantEntitlements(consortiumName string, tenantType constant.TenantType, purgeSchemas bool) error {
	args := m.Called(consortiumName, tenantType, purgeSchemas)
	return args.Error(0)
}

func TestNew(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	mockHTTP := &testhelpers.MockHTTPClient{}
	mockVault := &MockVaultClient{}
	mockMgmt := &MockManagementSvc{}

	// Act
	svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

	// Assert
	assert.NotNil(t, svc)
	assert.Equal(t, action, svc.Action)
	assert.Equal(t, mockHTTP, svc.HTTPClient)
	assert.Equal(t, mockVault, svc.VaultClient)
	assert.Equal(t, mockMgmt, svc.ManagementSvc)
}

func TestGetMasterAccessToken_Success(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	mockVault := &MockVaultClient{}
	mockMgmt := &MockManagementSvc{}
	svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

	expectedToken := "master-token-123"
	tokenData := map[string]any{
		"access_token": expectedToken,
	}

	mockHTTP.On("PostFormDataReturnStruct",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/realms/master/protocol/openid-connect/token")
		}),
		mock.MatchedBy(func(formData url.Values) bool {
			return formData.Get("grant_type") == "password" &&
				formData.Get("client_id") == "admin-cli" &&
				formData.Get("username") == "admin" &&
				formData.Get("password") == "admin"
		}),
		mock.Anything,
		mock.Anything).
		Run(func(args mock.Arguments) {
			target := args.Get(3).(*map[string]any)
			*target = tokenData
		}).
		Return(nil)

	// Act
	token, err := svc.GetMasterAccessToken(constant.Password)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedToken, token)
	mockHTTP.AssertExpectations(t)
}

func TestGetMasterAccessToken_HTTPError(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	mockVault := &MockVaultClient{}
	mockMgmt := &MockManagementSvc{}
	svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

	expectedError := errors.New("HTTP request failed")
	mockHTTP.On("PostFormDataReturnStruct",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything).
		Return(expectedError)

	// Act
	token, err := svc.GetMasterAccessToken(constant.Password)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Empty(t, token)
	mockHTTP.AssertExpectations(t)
}

func TestGetMasterAccessToken_NoToken(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	mockVault := &MockVaultClient{}
	mockMgmt := &MockManagementSvc{}
	svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

	tokenData := map[string]any{} // Empty response, no access_token

	mockHTTP.On("PostFormDataReturnStruct",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything).
		Run(func(args mock.Arguments) {
			target := args.Get(3).(*map[string]any)
			*target = tokenData
		}).
		Return(nil)

	// Act
	token, err := svc.GetMasterAccessToken(constant.Password)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "access token")
	assert.Empty(t, token)
	mockHTTP.AssertExpectations(t)
}

func TestGetMasterAccessToken_ClientCredentials_Success(t *testing.T) {
	t.Run("TestGetMasterAccessToken_ClientCredentials_Success", func(t *testing.T) {
		// Arrange
		mockHTTP := &testhelpers.MockHTTPClient{}
		action := testhelpers.NewMockAction()
		action.ConfigGlobalEnv = map[string]string{
			"kc_admin_client_id":     "test-admin-client",
			"kc_admin_client_secret": "test-admin-secret",
		}
		mockVault := &MockVaultClient{}
		mockMgmt := &MockManagementSvc{}
		svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

		expectedToken := "master-client-credentials-token"
		tokenData := map[string]any{
			"access_token": expectedToken,
		}

		mockHTTP.On("PostFormDataReturnStruct",
			mock.MatchedBy(func(urlStr string) bool {
				return strings.Contains(urlStr, "/realms/master/protocol/openid-connect/token")
			}),
			mock.MatchedBy(func(formData url.Values) bool {
				return formData.Get("grant_type") == "client_credentials" &&
					formData.Get("client_id") == "test-admin-client" &&
					formData.Get("client_secret") == "test-admin-secret"
			}),
			mock.Anything,
			mock.Anything).
			Run(func(args mock.Arguments) {
				target := args.Get(3).(*map[string]any)
				*target = tokenData
			}).
			Return(nil)

		// Act
		token, err := svc.GetMasterAccessToken(constant.ClientCredentials)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedToken, token)
		mockHTTP.AssertExpectations(t)
	})
}

func TestGetMasterAccessToken_ClientCredentials_HTTPError(t *testing.T) {
	t.Run("TestGetMasterAccessToken_ClientCredentials_HTTPError", func(t *testing.T) {
		// Arrange
		mockHTTP := &testhelpers.MockHTTPClient{}
		action := testhelpers.NewMockAction()
		action.ConfigGlobalEnv = map[string]string{
			"kc_admin_client_id":     "test-admin-client",
			"kc_admin_client_secret": "test-admin-secret",
		}
		mockVault := &MockVaultClient{}
		mockMgmt := &MockManagementSvc{}
		svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

		expectedError := errors.New("HTTP request failed")
		mockHTTP.On("PostFormDataReturnStruct",
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything).
			Return(expectedError)

		// Act
		token, err := svc.GetMasterAccessToken(constant.ClientCredentials)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Empty(t, token)
		mockHTTP.AssertExpectations(t)
	})
}

func TestGetMasterAccessToken_ClientCredentials_NoToken(t *testing.T) {
	t.Run("TestGetMasterAccessToken_ClientCredentials_NoToken", func(t *testing.T) {
		// Arrange
		mockHTTP := &testhelpers.MockHTTPClient{}
		action := testhelpers.NewMockAction()
		action.ConfigGlobalEnv = map[string]string{
			"kc_admin_client_id":     "test-admin-client",
			"kc_admin_client_secret": "test-admin-secret",
		}
		mockVault := &MockVaultClient{}
		mockMgmt := &MockManagementSvc{}
		svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

		tokenData := map[string]any{} // No access_token

		mockHTTP.On("PostFormDataReturnStruct",
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything).
			Run(func(args mock.Arguments) {
				target := args.Get(3).(*map[string]any)
				*target = tokenData
			}).
			Return(nil)

		// Act
		token, err := svc.GetMasterAccessToken(constant.ClientCredentials)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "access token")
		assert.Empty(t, token)
		mockHTTP.AssertExpectations(t)
	})
}

func TestUpdateRealmAccessTokenSettings_Success(t *testing.T) {
	t.Run("TestUpdateRealmAccessTokenSettings_Success", func(t *testing.T) {
		// Arrange
		mockHTTP := &testhelpers.MockHTTPClient{}
		action := testhelpers.NewMockAction()
		action.KeycloakMasterAccessToken = "test-master-token"
		mockVault := &MockVaultClient{}
		mockMgmt := &MockManagementSvc{}
		svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

		lifespan := 1800

		tenantName := "test-tenant"

		mockHTTP.On("PutReturnNoContent",
			mock.MatchedBy(func(urlStr string) bool {
				return strings.Contains(urlStr, "/admin/realms/"+tenantName)
			}),
			mock.MatchedBy(func(payload []byte) bool {
				var data map[string]any
				err := json.Unmarshal(payload, &data)
				if err != nil {
					return false
				}
				return data["accessTokenLifespan"].(float64) == 1800
			}),
			mock.MatchedBy(func(headers map[string]string) bool {
				return strings.Contains(headers[constant.AuthorizationHeader], "test-master-token")
			})).
			Return(nil)

		// Act
		err := svc.UpdateRealmAccessTokenSettings(tenantName, lifespan)

		// Assert
		assert.NoError(t, err)
		mockHTTP.AssertExpectations(t)
	})
}

func TestUpdateRealmAccessTokenSettings_HTTPError(t *testing.T) {
	t.Run("TestUpdateRealmAccessTokenSettings_HTTPError", func(t *testing.T) {
		// Arrange
		mockHTTP := &testhelpers.MockHTTPClient{}
		action := testhelpers.NewMockAction()
		action.KeycloakMasterAccessToken = "test-master-token"
		mockVault := &MockVaultClient{}
		mockMgmt := &MockManagementSvc{}
		svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

		expectedError := errors.New("HTTP PUT failed")
		mockHTTP.On("PutReturnNoContent",
			mock.Anything,
			mock.Anything,
			mock.Anything).
			Return(expectedError)

		// Act
		err := svc.UpdateRealmAccessTokenSettings("test-tenant", 1800)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		mockHTTP.AssertExpectations(t)
	})
}

func TestUpdateRealmAccessTokenSettings_ZeroLifespan(t *testing.T) {
	t.Run("TestUpdateRealmAccessTokenSettings_ZeroLifespan", func(t *testing.T) {
		// Arrange
		mockHTTP := &testhelpers.MockHTTPClient{}
		action := testhelpers.NewMockAction()
		action.KeycloakMasterAccessToken = "test-master-token"
		mockVault := &MockVaultClient{}
		mockMgmt := &MockManagementSvc{}
		svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

		mockHTTP.On("PutReturnNoContent",
			mock.Anything,
			mock.MatchedBy(func(payload []byte) bool {
				var data map[string]any
				err := json.Unmarshal(payload, &data)
				if err != nil {
					return false
				}
				return data["accessTokenLifespan"].(float64) == 0
			}),
			mock.Anything).
			Return(nil)

		// Act
		err := svc.UpdateRealmAccessTokenSettings("test-tenant", 0)

		// Assert
		assert.NoError(t, err)
		mockHTTP.AssertExpectations(t)
	})
}

func TestUpdateRealmAccessTokenSettings_LargeLifespan(t *testing.T) {
	t.Run("TestUpdateRealmAccessTokenSettings_LargeLifespan", func(t *testing.T) {
		// Arrange
		mockHTTP := &testhelpers.MockHTTPClient{}
		action := testhelpers.NewMockAction()
		action.KeycloakMasterAccessToken = "test-master-token"
		mockVault := &MockVaultClient{}
		mockMgmt := &MockManagementSvc{}
		svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

		lifespan := 86400 // 24 hours

		mockHTTP.On("PutReturnNoContent",
			mock.Anything,
			mock.MatchedBy(func(payload []byte) bool {
				var data map[string]any
				err := json.Unmarshal(payload, &data)
				if err != nil {
					return false
				}
				return data["accessTokenLifespan"].(float64) == 86400
			}),
			mock.Anything).
			Return(nil)

		// Act
		err := svc.UpdateRealmAccessTokenSettings("test-tenant", lifespan)

		// Assert
		assert.NoError(t, err)
		mockHTTP.AssertExpectations(t)
	})
}

func TestUpdateRealmAccessTokenSettings_HeaderCreationError(t *testing.T) {
	t.Run("TestUpdateRealmAccessTokenSettings_HeaderCreationError", func(t *testing.T) {
		// Arrange
		mockHTTP := &testhelpers.MockHTTPClient{}
		action := testhelpers.NewMockAction()
		action.KeycloakMasterAccessToken = "" // Empty token
		mockVault := &MockVaultClient{}
		mockMgmt := &MockManagementSvc{}
		svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

		// Act
		err := svc.UpdateRealmAccessTokenSettings("test-tenant", 1800)

		// Assert
		assert.Error(t, err)
		mockHTTP.AssertNotCalled(t, "PutReturnNoContent", mock.Anything, mock.Anything, mock.Anything)
	})
}

func TestUpdatePublicClientSettings_Success(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakMasterAccessToken = "test-master-token"
	action.ConfigGlobalEnv = map[string]string{
		"KC_LOGIN_CLIENT_SUFFIX": "",
	}
	mockVault := &MockVaultClient{}
	mockMgmt := &MockManagementSvc{}
	svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

	tenantName := "test-tenant"
	baseURL := "http://test.com"

	clientsResponse := models.KeycloakClientsResponse{
		{
			ID:       "client-uuid-123",
			ClientID: "test-application",
		},
	}

	mockHTTP.On("GetRetryReturnStruct",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/admin/realms/"+tenantName+"/clients")
		}),
		mock.Anything,
		mock.Anything).
		Run(func(args mock.Arguments) {
			target := args.Get(2).(*models.KeycloakClientsResponse)
			*target = clientsResponse
		}).
		Return(nil)

	mockHTTP.On("PutReturnNoContent",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/admin/realms/"+tenantName+"/clients/client-uuid-123")
		}),
		mock.MatchedBy(func(payload []byte) bool {
			var data map[string]any
			_ = json.Unmarshal(payload, &data)
			redirectURIs := data["redirectUris"].([]any)
			webOrigins := data["webOrigins"].([]any)
			attrs := data["attributes"].(map[string]any)
			return redirectURIs[0] == baseURL+"/*" &&
				webOrigins[0] == "/*" &&
				attrs["post.logout.redirect.uris"] == baseURL+"/*" &&
				attrs["login_theme"] == "custom-theme"
		}),
		mock.Anything).
		Return(nil)

	// Act
	err := svc.UpdatePublicClientSettings(tenantName, baseURL)

	// Assert
	assert.NoError(t, err)
	mockHTTP.AssertExpectations(t)
}

func TestUpdatePublicClientSettings_HTTPError(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakMasterAccessToken = "test-master-token"
	action.ConfigGlobalEnv = map[string]string{
		"KC_LOGIN_CLIENT_SUFFIX": "",
	}
	mockVault := &MockVaultClient{}
	mockMgmt := &MockManagementSvc{}
	svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

	expectedError := errors.New("HTTP GET failed")
	mockHTTP.On("GetRetryReturnStruct",
		mock.Anything,
		mock.Anything,
		mock.Anything).
		Return(expectedError)

	// Act
	err := svc.UpdatePublicClientSettings("test-tenant", "http://test.com")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockHTTP.AssertExpectations(t)
}

func TestUpdatePublicClientSettings_HeaderCreationError(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakMasterAccessToken = "" // Empty token
	action.ConfigGlobalEnv = map[string]string{
		"KC_LOGIN_CLIENT_SUFFIX": "",
	}
	mockVault := &MockVaultClient{}
	mockMgmt := &MockManagementSvc{}
	svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

	// Act
	err := svc.UpdatePublicClientSettings("test-tenant", "http://test.com")

	// Assert
	assert.Error(t, err)
	mockHTTP.AssertNotCalled(t, "GetRetryReturnStruct", mock.Anything, mock.Anything, mock.Anything)
	mockHTTP.AssertNotCalled(t, "PutReturnNoContent", mock.Anything, mock.Anything, mock.Anything)
}

// ==================== GetAccessToken Tests ====================

func TestGetAccessToken_Success(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.VaultRootToken = "root-token"
	action.ConfigGlobalEnv = map[string]string{
		"kc_service_client_id": "test-client-id",
	}
	mockVault := &MockVaultClient{}
	mockMgmt := &MockManagementSvc{}
	svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

	vaultClient := &vault.Client{}
	secrets := map[string]any{
		"test-client-id":          "client-secret-123",
		"test-tenant-system-user": "system-user-password",
	}

	mockVault.On("Create").Return(vaultClient, nil)
	mockVault.On("GetSecretKey",
		mock.Anything,
		vaultClient,
		"root-token",
		"folio/test-tenant").
		Return(secrets, nil)

	tokenData := map[string]any{
		"access_token": "tenant-access-token",
	}

	mockHTTP.On("PostFormDataReturnStruct",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/test-tenant/protocol/openid-connect/token")
		}),
		mock.Anything,
		mock.Anything,
		mock.Anything).
		Run(func(args mock.Arguments) {
			target := args.Get(3).(*map[string]any)
			*target = tokenData
		}).
		Return(nil)

	// Act
	token, err := svc.GetAccessToken("test-tenant")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "tenant-access-token", token)
	mockVault.AssertExpectations(t)
	mockHTTP.AssertExpectations(t)
}

func TestGetAccessToken_VaultCreateError(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	mockVault := &MockVaultClient{}
	mockMgmt := &MockManagementSvc{}
	svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

	expectedError := errors.New("vault create failed")
	mockVault.On("Create").Return(nil, expectedError)

	// Act
	token, err := svc.GetAccessToken("test-tenant")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Empty(t, token)
	mockVault.AssertExpectations(t)
}

func TestGetAccessToken_VaultGetSecretError(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.VaultRootToken = "root-token"
	mockVault := &MockVaultClient{}
	mockMgmt := &MockManagementSvc{}
	svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

	vaultClient := &vault.Client{}
	expectedError := errors.New("secret not found")

	mockVault.On("Create").Return(vaultClient, nil)
	mockVault.On("GetSecretKey",
		mock.Anything,
		vaultClient,
		"root-token",
		"folio/test-tenant").
		Return(nil, expectedError)

	// Act
	token, err := svc.GetAccessToken("test-tenant")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Empty(t, token)
	mockVault.AssertExpectations(t)
}

func TestGetAccessToken_HTTPError(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.VaultRootToken = "root-token"
	action.ConfigGlobalEnv = map[string]string{
		"kc_service_client_id": "test-client-id",
	}
	mockVault := &MockVaultClient{}
	mockMgmt := &MockManagementSvc{}
	svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

	vaultClient := &vault.Client{}
	secrets := map[string]any{
		"test-client-id":          "client-secret-123",
		"test-tenant-system-user": "system-user-password",
	}

	mockVault.On("Create").Return(vaultClient, nil)
	mockVault.On("GetSecretKey",
		mock.Anything,
		vaultClient,
		"root-token",
		"folio/test-tenant").
		Return(secrets, nil)

	expectedError := errors.New("HTTP error")
	mockHTTP.On("PostFormDataReturnStruct",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything).
		Return(expectedError)

	// Act
	token, err := svc.GetAccessToken("test-tenant")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Empty(t, token)
	mockVault.AssertExpectations(t)
	mockHTTP.AssertExpectations(t)
}

func TestGetAccessToken_NoToken(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.VaultRootToken = "root-token"
	action.ConfigGlobalEnv = map[string]string{
		"kc_service_client_id": "test-client-id",
	}
	mockVault := &MockVaultClient{}
	mockMgmt := &MockManagementSvc{}
	svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

	vaultClient := &vault.Client{}
	secrets := map[string]any{
		"test-client-id":          "client-secret-123",
		"test-tenant-system-user": "system-user-password",
	}

	mockVault.On("Create").Return(vaultClient, nil)
	mockVault.On("GetSecretKey",
		mock.Anything,
		vaultClient,
		"root-token",
		"folio/test-tenant").
		Return(secrets, nil)

	tokenData := map[string]any{} // No access_token

	mockHTTP.On("PostFormDataReturnStruct",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything).
		Run(func(args mock.Arguments) {
			target := args.Get(3).(*map[string]any)
			*target = tokenData
		}).
		Return(nil)

	// Act
	token, err := svc.GetAccessToken("test-tenant")

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "access token")
	assert.Empty(t, token)
}

// ==================== Role Tests ====================

func TestGetRoles_Success(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	mockVault := &MockVaultClient{}
	mockMgmt := &MockManagementSvc{}
	svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

	rolesResponse := models.KeycloakRolesResponse{
		Roles: []models.KeycloakRole{
			{ID: "role-1", Name: "Admin", Description: "Admin role"},
			{ID: "role-2", Name: "User", Description: "User role"},
		},
	}

	mockHTTP.On("GetRetryReturnStruct",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/roles?offset=0&limit=10000")
		}),
		mock.Anything,
		mock.Anything).
		Run(func(args mock.Arguments) {
			target := args.Get(2).(*models.KeycloakRolesResponse)
			*target = rolesResponse
		}).
		Return(nil)

	// Act
	roles, err := svc.GetRoles(map[string]string{})

	// Assert
	assert.NoError(t, err)
	assert.Len(t, roles, 2)
	role1 := roles[0].(map[string]any)
	assert.Equal(t, "role-1", role1["id"])
	assert.Equal(t, "Admin", role1["name"])
	mockHTTP.AssertExpectations(t)
}

func TestGetRoles_EmptyResponse(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	mockVault := &MockVaultClient{}
	mockMgmt := &MockManagementSvc{}
	svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

	rolesResponse := models.KeycloakRolesResponse{Roles: []models.KeycloakRole{}}

	mockHTTP.On("GetRetryReturnStruct",
		mock.Anything,
		mock.Anything,
		mock.Anything).
		Run(func(args mock.Arguments) {
			target := args.Get(2).(*models.KeycloakRolesResponse)
			*target = rolesResponse
		}).
		Return(nil)

	// Act
	roles, err := svc.GetRoles(map[string]string{})

	// Assert
	assert.NoError(t, err)
	assert.Nil(t, roles)
	mockHTTP.AssertExpectations(t)
}

func TestGetRoles_HTTPError(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	mockVault := &MockVaultClient{}
	mockMgmt := &MockManagementSvc{}
	svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

	expectedError := errors.New("HTTP GET failed")
	mockHTTP.On("GetRetryReturnStruct",
		mock.Anything,
		mock.Anything,
		mock.Anything).
		Return(expectedError)

	// Act
	roles, err := svc.GetRoles(map[string]string{})

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, roles)
	mockHTTP.AssertExpectations(t)
}

func TestGetRoleByName_Success(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	mockVault := &MockVaultClient{}
	mockMgmt := &MockManagementSvc{}
	svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

	rolesResponse := models.KeycloakRolesResponse{
		Roles: []models.KeycloakRole{
			{ID: "role-1", Name: "Admin", Description: "Admin role"},
		},
	}

	mockHTTP.On("GetRetryReturnStruct",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/roles?query=name==Admin&limit=1")
		}),
		mock.Anything,
		mock.Anything).
		Run(func(args mock.Arguments) {
			target := args.Get(2).(*models.KeycloakRolesResponse)
			*target = rolesResponse
		}).
		Return(nil)

	// Act
	role, err := svc.GetRoleByName("Admin", map[string]string{})

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, role)
	assert.Equal(t, "role-1", role["id"])
	assert.Equal(t, "Admin", role["name"])
	mockHTTP.AssertExpectations(t)
}

func TestGetRoleByName_NotFound(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	mockVault := &MockVaultClient{}
	mockMgmt := &MockManagementSvc{}
	svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

	rolesResponse := models.KeycloakRolesResponse{Roles: []models.KeycloakRole{}}

	mockHTTP.On("GetRetryReturnStruct",
		mock.Anything,
		mock.Anything,
		mock.Anything).
		Run(func(args mock.Arguments) {
			target := args.Get(2).(*models.KeycloakRolesResponse)
			*target = rolesResponse
		}).
		Return(nil)

	// Act
	role, err := svc.GetRoleByName("NonExistent", map[string]string{})

	// Assert
	assert.NoError(t, err)
	assert.Nil(t, role)
	mockHTTP.AssertExpectations(t)
}

func TestGetRoleByName_MultipleFound(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	mockVault := &MockVaultClient{}
	mockMgmt := &MockManagementSvc{}
	svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

	rolesResponse := models.KeycloakRolesResponse{
		Roles: []models.KeycloakRole{
			{ID: "role-1", Name: "Admin", Description: "Admin role 1"},
			{ID: "role-2", Name: "Admin", Description: "Admin role 2"},
		},
	}

	mockHTTP.On("GetRetryReturnStruct",
		mock.Anything,
		mock.Anything,
		mock.Anything).
		Run(func(args mock.Arguments) {
			target := args.Get(2).(*models.KeycloakRolesResponse)
			*target = rolesResponse
		}).
		Return(nil)

	// Act
	role, err := svc.GetRoleByName("Admin", map[string]string{})

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Admin")
	assert.Nil(t, role)
	mockHTTP.AssertExpectations(t)
}

func TestCreateRoles_Success(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "test-token"
	action.ConfigRoles = map[string]any{
		"admin": map[string]any{
			"tenant": "test-tenant",
		},
		"user": map[string]any{
			"tenant": "test-tenant",
		},
	}
	mockVault := &MockVaultClient{}
	mockMgmt := &MockManagementSvc{}
	svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

	mockHTTP.On("PostReturnNoContent",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/roles")
		}),
		mock.MatchedBy(func(payload []byte) bool {
			var data map[string]string
			_ = json.Unmarshal(payload, &data)
			return data["description"] == "Default"
		}),
		mock.Anything).
		Return(nil).Times(2)

	// Act
	err := svc.CreateRoles("test-tenant")

	// Assert
	assert.NoError(t, err)
	mockHTTP.AssertExpectations(t)
}

func TestCreateRoles_SkipsDifferentTenant(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.ConfigRoles = map[string]any{
		"admin": map[string]any{
			"tenant": "other-tenant",
		},
	}
	mockVault := &MockVaultClient{}
	mockMgmt := &MockManagementSvc{}
	svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

	// Act
	err := svc.CreateRoles("test-tenant")

	// Assert
	assert.NoError(t, err)
	mockHTTP.AssertNotCalled(t, "PostReturnNoContent")
}

func TestCreateRoles_HTTPError(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "test-token"
	action.ConfigRoles = map[string]any{
		"admin": map[string]any{
			"tenant": "test-tenant",
		},
	}
	mockVault := &MockVaultClient{}
	mockMgmt := &MockManagementSvc{}
	svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

	expectedError := errors.New("HTTP POST failed")
	mockHTTP.On("PostReturnNoContent",
		mock.Anything,
		mock.Anything,
		mock.Anything).
		Return(expectedError)

	// Act
	err := svc.CreateRoles("test-tenant")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockHTTP.AssertExpectations(t)
}

func TestCreateRoles_HeaderCreationError(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "" // Empty token
	action.ConfigRoles = map[string]any{
		"admin": map[string]any{
			"tenant": "test-tenant",
		},
	}
	mockVault := &MockVaultClient{}
	mockMgmt := &MockManagementSvc{}
	svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

	// Act
	err := svc.CreateRoles("test-tenant")

	// Assert
	assert.Error(t, err)
	mockHTTP.AssertNotCalled(t, "PostReturnNoContent", mock.Anything, mock.Anything, mock.Anything)
}

func TestRemoveRoles_Success(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "test-token"
	action.ConfigRoles = map[string]any{
		"admin": map[string]any{
			"tenant": "test-tenant",
		},
	}
	mockVault := &MockVaultClient{}
	mockMgmt := &MockManagementSvc{}
	svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

	rolesResponse := models.KeycloakRolesResponse{
		Roles: []models.KeycloakRole{
			{ID: "role-1", Name: "admin", Description: "Admin role"},
		},
	}

	mockHTTP.On("GetRetryReturnStruct",
		mock.Anything,
		mock.Anything,
		mock.Anything).
		Run(func(args mock.Arguments) {
			target := args.Get(2).(*models.KeycloakRolesResponse)
			*target = rolesResponse
		}).
		Return(nil)

	mockHTTP.On("Delete",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/roles/role-1")
		}),
		mock.Anything).
		Return(nil)

	// Act
	err := svc.RemoveRoles("test-tenant")

	// Assert
	assert.NoError(t, err)
	mockHTTP.AssertExpectations(t)
}

func TestRemoveRoles_GetRolesError(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "test-token"
	mockVault := &MockVaultClient{}
	mockMgmt := &MockManagementSvc{}
	svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

	expectedError := errors.New("Get roles failed")
	mockHTTP.On("GetRetryReturnStruct",
		mock.Anything,
		mock.Anything,
		mock.Anything).
		Return(expectedError)

	// Act
	err := svc.RemoveRoles("test-tenant")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockHTTP.AssertExpectations(t)
}

func TestRemoveRoles_HeaderCreationError(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "" // Empty token
	mockVault := &MockVaultClient{}
	mockMgmt := &MockManagementSvc{}
	svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

	// Act
	err := svc.RemoveRoles("test-tenant")

	// Assert
	assert.Error(t, err)
	mockHTTP.AssertNotCalled(t, "GetRetryReturnStruct", mock.Anything, mock.Anything, mock.Anything)
	mockHTTP.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
}

// ==================== User Tests ====================

func TestGetUsers_Success(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "test-token"
	mockVault := &MockVaultClient{}
	mockMgmt := &MockManagementSvc{}
	svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

	usersResponse := models.KeycloakUsersResponse{
		Users: []models.KeycloakUser{
			{
				ID:       "user-1",
				Username: "admin",
				Active:   true,
				Type:     "staff",
				Personal: map[string]any{"firstName": "John"},
			},
		},
	}

	mockHTTP.On("GetRetryReturnStruct",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/users?offset=0&limit=10000")
		}),
		mock.MatchedBy(func(headers map[string]string) bool {
			return headers[constant.OkapiTenantHeader] == "test-tenant" &&
				headers[constant.OkapiTokenHeader] == "test-token" &&
				headers[constant.ContentTypeHeader] == constant.ApplicationJSON
		}),
		mock.Anything).
		Run(func(args mock.Arguments) {
			target := args.Get(2).(*models.KeycloakUsersResponse)
			*target = usersResponse
		}).
		Return(nil)

	// Act
	users, err := svc.GetUsers("test-tenant")

	// Assert
	assert.NoError(t, err)
	assert.Len(t, users, 1)
	user := users[0].(map[string]any)
	assert.Equal(t, "user-1", user["id"])
	assert.Equal(t, "admin", user["username"])
	mockHTTP.AssertExpectations(t)
}

func TestGetUsers_EmptyResponse(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "test-token"
	mockVault := &MockVaultClient{}
	mockMgmt := &MockManagementSvc{}
	svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

	usersResponse := models.KeycloakUsersResponse{Users: []models.KeycloakUser{}}

	mockHTTP.On("GetRetryReturnStruct",
		mock.Anything,
		mock.Anything,
		mock.Anything).
		Run(func(args mock.Arguments) {
			target := args.Get(2).(*models.KeycloakUsersResponse)
			*target = usersResponse
		}).
		Return(nil)

	// Act
	users, err := svc.GetUsers("test-tenant")

	// Assert
	assert.NoError(t, err)
	assert.Nil(t, users)
	mockHTTP.AssertExpectations(t)
}

func TestGetUsers_HTTPError(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "test-token"
	mockVault := &MockVaultClient{}
	mockMgmt := &MockManagementSvc{}
	svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

	expectedError := errors.New("HTTP GET failed")
	mockHTTP.On("GetRetryReturnStruct",
		mock.Anything,
		mock.Anything,
		mock.Anything).
		Return(expectedError)

	// Act
	users, err := svc.GetUsers("test-tenant")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, users)
	mockHTTP.AssertExpectations(t)
}

func TestCreateUsers_Success(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "test-token"
	action.ConfigUsers = map[string]any{
		"testuser": map[string]any{
			"tenant":     "test-tenant",
			"password":   "pass123",
			"first-name": "Test",
			"last-name":  "User",
			"roles":      []any{"admin"},
		},
	}
	mockVault := &MockVaultClient{}
	mockMgmt := &MockManagementSvc{}
	svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

	createdUser := map[string]any{"id": "user-123"}

	mockHTTP.On("PostReturnStruct",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/users-keycloak/users")
		}),
		mock.Anything,
		mock.MatchedBy(func(headers map[string]string) bool {
			return headers[constant.OkapiTenantHeader] == "test-tenant" &&
				headers[constant.OkapiTokenHeader] == "test-token" &&
				headers[constant.ContentTypeHeader] == constant.ApplicationJSON
		}),
		mock.Anything).
		Run(func(args mock.Arguments) {
			target := args.Get(3).(*map[string]any)
			*target = createdUser
		}).
		Return(nil)

	mockHTTP.On("PostReturnNoContent",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/authn/credentials")
		}),
		mock.Anything,
		mock.MatchedBy(func(headers map[string]string) bool {
			return headers[constant.OkapiTenantHeader] == "test-tenant" &&
				headers[constant.OkapiTokenHeader] == "test-token" &&
				headers[constant.ContentTypeHeader] == constant.ApplicationJSON
		})).
		Return(nil)

	rolesResponse := models.KeycloakRolesResponse{
		Roles: []models.KeycloakRole{
			{ID: "role-1", Name: "admin", Description: "Admin"},
		},
	}

	mockHTTP.On("GetRetryReturnStruct",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/roles?query=name==admin")
		}),
		mock.MatchedBy(func(headers map[string]string) bool {
			return headers[constant.OkapiTenantHeader] == "test-tenant" &&
				headers[constant.OkapiTokenHeader] == "test-token" &&
				headers[constant.ContentTypeHeader] == constant.ApplicationJSON
		}),
		mock.Anything).
		Run(func(args mock.Arguments) {
			target := args.Get(2).(*models.KeycloakRolesResponse)
			*target = rolesResponse
		}).
		Return(nil)

	mockHTTP.On("PostReturnNoContent",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/roles/users")
		}),
		mock.Anything,
		mock.MatchedBy(func(headers map[string]string) bool {
			return headers[constant.OkapiTenantHeader] == "test-tenant" &&
				headers[constant.OkapiTokenHeader] == "test-token" &&
				headers[constant.ContentTypeHeader] == constant.ApplicationJSON
		})).
		Return(nil)

	// Act
	err := svc.CreateUsers("test-tenant")

	// Assert
	assert.NoError(t, err)
	mockHTTP.AssertExpectations(t)
}

func TestCreateUsers_SkipsDifferentTenant(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.ConfigUsers = map[string]any{
		"testuser": map[string]any{
			"tenant": "other-tenant",
		},
	}
	mockVault := &MockVaultClient{}
	mockMgmt := &MockManagementSvc{}
	svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

	// Act
	err := svc.CreateUsers("test-tenant")

	// Assert
	assert.NoError(t, err)
	mockHTTP.AssertNotCalled(t, "PostReturnStruct")
}

func TestCreateUsers_HeaderCreationError(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "" // Empty token
	action.ConfigUsers = map[string]any{
		"testuser": map[string]any{
			"tenant":     "test-tenant",
			"password":   "pass123",
			"first-name": "Test",
			"last-name":  "User",
			"roles":      []any{"admin"},
		},
	}
	mockVault := &MockVaultClient{}
	mockMgmt := &MockManagementSvc{}
	svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

	// Act
	err := svc.CreateUsers("test-tenant")

	// Assert
	assert.Error(t, err)
	mockHTTP.AssertNotCalled(t, "PostReturnStruct", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	mockHTTP.AssertNotCalled(t, "PostReturnNoContent", mock.Anything, mock.Anything, mock.Anything)
}

func TestCreateUsers_BlankTenantName(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "test-token"
	action.ConfigUsers = map[string]any{
		"testuser": map[string]any{
			"tenant":     "", // Empty tenant in user config
			"password":   "pass123",
			"first-name": "Test",
			"last-name":  "User",
			"roles":      []any{"admin"},
		},
	}
	mockVault := &MockVaultClient{}
	mockMgmt := &MockManagementSvc{}
	svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

	// Act
	err := svc.CreateUsers("") // Call with empty tenant to match the user's tenant

	// Assert
	assert.Error(t, err) // Error because tenant is blank when creating headers
	mockHTTP.AssertNotCalled(t, "PostReturnStruct", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestRemoveUsers_Success(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "test-token"
	action.ConfigUsers = map[string]any{
		"testuser": map[string]any{},
	}
	mockVault := &MockVaultClient{}
	mockMgmt := &MockManagementSvc{}
	svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

	usersResponse := models.KeycloakUsersResponse{
		Users: []models.KeycloakUser{
			{ID: "user-1", Username: "testuser", Active: true},
		},
	}

	mockHTTP.On("GetRetryReturnStruct",
		mock.Anything,
		mock.Anything,
		mock.Anything).
		Run(func(args mock.Arguments) {
			target := args.Get(2).(*models.KeycloakUsersResponse)
			*target = usersResponse
		}).
		Return(nil)

	mockHTTP.On("Delete",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/users-keycloak/users/user-1")
		}),
		mock.Anything).
		Return(nil)

	// Act
	err := svc.RemoveUsers("test-tenant")

	// Assert
	assert.NoError(t, err)
	mockHTTP.AssertExpectations(t)
}

func TestRemoveUsers_GetUsersError(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "test-token"
	mockVault := &MockVaultClient{}
	mockMgmt := &MockManagementSvc{}
	svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

	expectedError := errors.New("Get users failed")
	mockHTTP.On("GetRetryReturnStruct",
		mock.Anything,
		mock.Anything,
		mock.Anything).
		Return(expectedError)

	// Act
	err := svc.RemoveUsers("test-tenant")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockHTTP.AssertExpectations(t)
}

func TestRemoveUsers_HeaderCreationError(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "" // Empty token
	mockVault := &MockVaultClient{}
	mockMgmt := &MockManagementSvc{}
	svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

	// Act
	err := svc.RemoveUsers("test-tenant")

	// Assert
	assert.Error(t, err)
	mockHTTP.AssertNotCalled(t, "GetRetryReturnStruct", mock.Anything, mock.Anything, mock.Anything)
	mockHTTP.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
}

func TestRemoveUsers_BlankTenantName(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "test-token"
	mockVault := &MockVaultClient{}
	mockMgmt := &MockManagementSvc{}
	svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

	// Act
	err := svc.RemoveUsers("") // Empty tenant

	// Assert
	assert.Error(t, err)
	mockHTTP.AssertNotCalled(t, "GetRetryReturnStruct", mock.Anything, mock.Anything, mock.Anything)
	mockHTTP.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
}

// ==================== Capability Set Tests ====================

func TestGetCapabilitySets_Success(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	mockVault := &MockVaultClient{}
	mockMgmt := &MockManagementSvc{}
	svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

	applications := models.ApplicationsResponse{
		ApplicationDescriptors: []map[string]any{
			{"id": "app-1"},
		},
	}

	mockMgmt.On("GetApplications").Return(applications, nil)

	capSetsResponse := models.KeycloakCapabilitySetsResponse{
		CapabilitySets: []models.KeycloakCapabilitySet{
			{
				ID:            "cap-1",
				Name:          "users.read",
				ApplicationID: "app-1",
				Resource:      "users",
				Action:        "read",
			},
		},
	}

	mockHTTP.On("GetRetryReturnStruct",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/capability-sets?query=applicationId==app-1")
		}),
		mock.Anything,
		mock.Anything).
		Run(func(args mock.Arguments) {
			target := args.Get(2).(*models.KeycloakCapabilitySetsResponse)
			*target = capSetsResponse
		}).
		Return(nil)

	// Act
	capSets, err := svc.GetCapabilitySets(map[string]string{})

	// Assert
	assert.NoError(t, err)
	assert.Len(t, capSets, 1)
	capSet := capSets[0].(map[string]any)
	assert.Equal(t, "cap-1", capSet["id"])
	assert.Equal(t, "users.read", capSet["name"])
	mockMgmt.AssertExpectations(t)
	mockHTTP.AssertExpectations(t)
}

func TestGetCapabilitySets_GetApplicationsError(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	mockVault := &MockVaultClient{}
	mockMgmt := &MockManagementSvc{}
	svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

	expectedError := errors.New("Get applications failed")
	mockMgmt.On("GetApplications").Return(nil, expectedError)

	// Act
	capSets, err := svc.GetCapabilitySets(map[string]string{})

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, capSets)
	mockMgmt.AssertExpectations(t)
}

func TestGetCapabilitySetsByName_Success(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	mockVault := &MockVaultClient{}
	mockMgmt := &MockManagementSvc{}
	svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

	capSetsResponse := models.KeycloakCapabilitySetsResponse{
		CapabilitySets: []models.KeycloakCapabilitySet{
			{
				ID:            "cap-1",
				Name:          "users.read",
				ApplicationID: "app-1",
			},
		},
	}

	mockHTTP.On("GetRetryReturnStruct",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/capability-sets?query=name==users.read")
		}),
		mock.Anything,
		mock.Anything).
		Run(func(args mock.Arguments) {
			target := args.Get(2).(*models.KeycloakCapabilitySetsResponse)
			*target = capSetsResponse
		}).
		Return(nil)

	// Act
	capSets, err := svc.GetCapabilitySetsByName(map[string]string{}, "users.read")

	// Assert
	assert.NoError(t, err)
	assert.Len(t, capSets, 1)
	mockHTTP.AssertExpectations(t)
}

func TestGetCapabilitySetsByName_EmptyResponse(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	mockVault := &MockVaultClient{}
	mockMgmt := &MockManagementSvc{}
	svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

	capSetsResponse := models.KeycloakCapabilitySetsResponse{CapabilitySets: []models.KeycloakCapabilitySet{}}

	mockHTTP.On("GetRetryReturnStruct",
		mock.Anything,
		mock.Anything,
		mock.Anything).
		Run(func(args mock.Arguments) {
			target := args.Get(2).(*models.KeycloakCapabilitySetsResponse)
			*target = capSetsResponse
		}).
		Return(nil)

	// Act
	capSets, err := svc.GetCapabilitySetsByName(map[string]string{}, "nonexistent")

	// Assert
	assert.NoError(t, err)
	assert.Nil(t, capSets)
	mockHTTP.AssertExpectations(t)
}

func TestAttachCapabilitySetsToRoles_Success(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "test-token"
	action.ConfigRoles = map[string]any{
		"admin": map[string]any{
			"tenant":          "test-tenant",
			"capability-sets": []any{"users.read"},
		},
	}
	mockVault := &MockVaultClient{}
	mockMgmt := &MockManagementSvc{}
	svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

	rolesResponse := models.KeycloakRolesResponse{
		Roles: []models.KeycloakRole{
			{ID: "role-1", Name: "admin"},
		},
	}

	mockHTTP.On("GetRetryReturnStruct",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/roles?offset=0&limit=10000")
		}),
		mock.Anything,
		mock.Anything).
		Run(func(args mock.Arguments) {
			target := args.Get(2).(*models.KeycloakRolesResponse)
			*target = rolesResponse
		}).
		Return(nil)

	capSetsResponse := models.KeycloakCapabilitySetsResponse{
		CapabilitySets: []models.KeycloakCapabilitySet{
			{ID: "cap-1", Name: "users.read"},
		},
	}

	mockHTTP.On("GetRetryReturnStruct",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/capability-sets?query=name==users.read")
		}),
		mock.Anything,
		mock.Anything).
		Run(func(args mock.Arguments) {
			target := args.Get(2).(*models.KeycloakCapabilitySetsResponse)
			*target = capSetsResponse
		}).
		Return(nil)

	mockHTTP.On("PostRetryReturnNoContent",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/roles/capability-sets")
		}),
		mock.Anything,
		mock.Anything).
		Return(nil)

	// Act
	err := svc.AttachCapabilitySetsToRoles("test-tenant")

	// Assert
	assert.NoError(t, err)
	mockHTTP.AssertExpectations(t)
}

func TestAttachCapabilitySetsToRoles_NoRoles(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "test-token"
	mockVault := &MockVaultClient{}
	mockMgmt := &MockManagementSvc{}
	svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

	rolesResponse := models.KeycloakRolesResponse{Roles: []models.KeycloakRole{}}

	mockHTTP.On("GetRetryReturnStruct",
		mock.Anything,
		mock.Anything,
		mock.Anything).
		Run(func(args mock.Arguments) {
			target := args.Get(2).(*models.KeycloakRolesResponse)
			*target = rolesResponse
		}).
		Return(nil)

	// Act
	err := svc.AttachCapabilitySetsToRoles("test-tenant")

	// Assert
	assert.NoError(t, err)
	mockHTTP.AssertExpectations(t)
}

func TestAttachCapabilitySetsToRoles_GetRolesError(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "test-token"
	mockVault := &MockVaultClient{}
	mockMgmt := &MockManagementSvc{}
	svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

	expectedError := errors.New("Get roles failed")
	mockHTTP.On("GetRetryReturnStruct",
		mock.Anything,
		mock.Anything,
		mock.Anything).
		Return(expectedError)

	// Act
	err := svc.AttachCapabilitySetsToRoles("test-tenant")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockHTTP.AssertExpectations(t)
}

func TestAttachCapabilitySetsToRoles_HeaderCreationError(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "" // Empty token
	mockVault := &MockVaultClient{}
	mockMgmt := &MockManagementSvc{}
	svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

	// Act
	err := svc.AttachCapabilitySetsToRoles("test-tenant")

	// Assert
	assert.Error(t, err)
	mockHTTP.AssertNotCalled(t, "GetRetryReturnStruct", mock.Anything, mock.Anything, mock.Anything)
	mockHTTP.AssertNotCalled(t, "PostRetryReturnNoContent", mock.Anything, mock.Anything, mock.Anything)
}

func TestDetachCapabilitySetsFromRoles_Success(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "test-token"
	action.ConfigRoles = map[string]any{
		"admin": map[string]any{},
	}
	mockVault := &MockVaultClient{}
	mockMgmt := &MockManagementSvc{}
	svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

	rolesResponse := models.KeycloakRolesResponse{
		Roles: []models.KeycloakRole{
			{ID: "role-1", Name: "admin"},
		},
	}

	mockHTTP.On("GetRetryReturnStruct",
		mock.Anything,
		mock.Anything,
		mock.Anything).
		Run(func(args mock.Arguments) {
			target := args.Get(2).(*models.KeycloakRolesResponse)
			*target = rolesResponse
		}).
		Return(nil)

	mockHTTP.On("Delete",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/roles/role-1/capability-sets")
		}),
		mock.Anything).
		Return(nil)

	// Act
	err := svc.DetachCapabilitySetsFromRoles("test-tenant")

	// Assert
	assert.NoError(t, err)
	mockHTTP.AssertExpectations(t)
}

func TestDetachCapabilitySetsFromRoles_NoRoles(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "test-token"
	mockVault := &MockVaultClient{}
	mockMgmt := &MockManagementSvc{}
	svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

	rolesResponse := models.KeycloakRolesResponse{Roles: []models.KeycloakRole{}}

	mockHTTP.On("GetRetryReturnStruct",
		mock.Anything,
		mock.Anything,
		mock.Anything).
		Run(func(args mock.Arguments) {
			target := args.Get(2).(*models.KeycloakRolesResponse)
			*target = rolesResponse
		}).
		Return(nil)

	// Act
	err := svc.DetachCapabilitySetsFromRoles("test-tenant")

	// Assert
	assert.NoError(t, err)
	mockHTTP.AssertExpectations(t)
}

func TestDetachCapabilitySetsFromRoles_HeaderCreationError(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "" // Empty token
	mockVault := &MockVaultClient{}
	mockMgmt := &MockManagementSvc{}
	svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

	// Act
	err := svc.DetachCapabilitySetsFromRoles("test-tenant")

	// Assert
	assert.Error(t, err)
	mockHTTP.AssertNotCalled(t, "GetRetryReturnStruct", mock.Anything, mock.Anything, mock.Anything)
	mockHTTP.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
}

func TestDetachCapabilitySetsFromRoles_GetRolesError(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "test-token"
	mockVault := &MockVaultClient{}
	mockMgmt := &MockManagementSvc{}
	svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

	mockHTTP.On("GetRetryReturnStruct",
		mock.Anything,
		mock.Anything,
		mock.Anything).
		Return(errors.New("get roles failed"))

	// Act
	err := svc.DetachCapabilitySetsFromRoles("test-tenant")

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "get roles failed")
	mockHTTP.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
}

func TestDetachCapabilitySetsFromRoles_NotFoundIgnored(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "test-token"
	action.ConfigRoles = map[string]any{
		"admin": map[string]any{},
	}
	mockVault := &MockVaultClient{}
	mockMgmt := &MockManagementSvc{}
	svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

	rolesResponse := models.KeycloakRolesResponse{
		Roles: []models.KeycloakRole{
			{ID: "role-1", Name: "admin"},
		},
	}

	mockHTTP.On("GetRetryReturnStruct",
		mock.Anything,
		mock.Anything,
		mock.Anything).
		Run(func(args mock.Arguments) {
			target := args.Get(2).(*models.KeycloakRolesResponse)
			*target = rolesResponse
		}).
		Return(nil)

	mockHTTP.On("Delete",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/roles/role-1/capability-sets")
		}),
		mock.Anything).
		Return(apperrors.ErrHTTP404NotFound)

	// Act
	err := svc.DetachCapabilitySetsFromRoles("test-tenant")

	// Assert
	// 404 errors should be logged but not returned as errors
	assert.NoError(t, err)
	mockHTTP.AssertExpectations(t)
}

func TestDetachCapabilitySetsFromRoles_DeleteError(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "test-token"
	action.ConfigRoles = map[string]any{
		"admin": map[string]any{},
	}
	mockVault := &MockVaultClient{}
	mockMgmt := &MockManagementSvc{}
	svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

	rolesResponse := models.KeycloakRolesResponse{
		Roles: []models.KeycloakRole{
			{ID: "role-1", Name: "admin"},
		},
	}

	mockHTTP.On("GetRetryReturnStruct",
		mock.Anything,
		mock.Anything,
		mock.Anything).
		Run(func(args mock.Arguments) {
			target := args.Get(2).(*models.KeycloakRolesResponse)
			*target = rolesResponse
		}).
		Return(nil)

	mockHTTP.On("Delete",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/roles/role-1/capability-sets")
		}),
		mock.Anything).
		Return(errors.New("delete failed"))

	// Act
	err := svc.DetachCapabilitySetsFromRoles("test-tenant")

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delete failed")
	mockHTTP.AssertExpectations(t)
}

func TestDetachCapabilitySetsFromRoles_RoleNotInConfig(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "test-token"
	action.ConfigRoles = map[string]any{
		"admin": map[string]any{},
	}
	mockVault := &MockVaultClient{}
	mockMgmt := &MockManagementSvc{}
	svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

	rolesResponse := models.KeycloakRolesResponse{
		Roles: []models.KeycloakRole{
			{ID: "role-1", Name: "user"}, // Not in config
			{ID: "role-2", Name: "admin"},
		},
	}

	mockHTTP.On("GetRetryReturnStruct",
		mock.Anything,
		mock.Anything,
		mock.Anything).
		Run(func(args mock.Arguments) {
			target := args.Get(2).(*models.KeycloakRolesResponse)
			*target = rolesResponse
		}).
		Return(nil)

	// Only admin role should be deleted, not user role
	mockHTTP.On("Delete",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/roles/role-2/capability-sets")
		}),
		mock.Anything).
		Return(nil)

	// Act
	err := svc.DetachCapabilitySetsFromRoles("test-tenant")

	// Assert
	assert.NoError(t, err)
	mockHTTP.AssertExpectations(t)
	// Verify Delete was NOT called for role-1
	mockHTTP.AssertNotCalled(t, "Delete",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/roles/role-1/capability-sets")
		}),
		mock.Anything)
}

func TestAttachCapabilitySetsToRoles_WithAllCapabilitySets(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "test-token"
	action.ConfigRoles = map[string]any{
		"admin": map[string]any{
			"tenant":          "test-tenant",
			"capability-sets": []any{"all"},
		},
	}
	mockVault := &MockVaultClient{}
	mockMgmt := &MockManagementSvc{}
	svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

	rolesResponse := models.KeycloakRolesResponse{
		Roles: []models.KeycloakRole{
			{ID: "role-1", Name: "admin"},
		},
	}

	applicationsResponse := models.ApplicationsResponse{
		ApplicationDescriptors: []map[string]any{
			{"id": "app-1"},
		},
	}

	capSetsResponse := models.KeycloakCapabilitySetsResponse{
		CapabilitySets: []models.KeycloakCapabilitySet{
			{ID: "cap-1", Name: "users.read"},
			{ID: "cap-2", Name: "users.write"},
		},
	}

	mockHTTP.On("GetRetryReturnStruct",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/roles?offset=0&limit=10000")
		}),
		mock.Anything,
		mock.Anything).
		Run(func(args mock.Arguments) {
			target := args.Get(2).(*models.KeycloakRolesResponse)
			*target = rolesResponse
		}).
		Return(nil)

	mockMgmt.On("GetApplications").Return(applicationsResponse, nil)

	mockHTTP.On("GetRetryReturnStruct",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/capability-sets?query=applicationId==app-1")
		}),
		mock.Anything,
		mock.Anything).
		Run(func(args mock.Arguments) {
			target := args.Get(2).(*models.KeycloakCapabilitySetsResponse)
			*target = capSetsResponse
		}).
		Return(nil)

	mockHTTP.On("PostRetryReturnNoContent",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/roles/capability-sets")
		}),
		mock.Anything,
		mock.Anything).
		Return(nil)

	// Act
	err := svc.AttachCapabilitySetsToRoles("test-tenant")

	// Assert
	assert.NoError(t, err)
	mockHTTP.AssertExpectations(t)
	mockMgmt.AssertExpectations(t)
}

func TestAttachCapabilitySetsToRoles_LargeBatchSplitting(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "test-token"
	action.ConfigRoles = map[string]any{
		"admin": map[string]any{
			"tenant":          "test-tenant",
			"capability-sets": []any{"all"},
		},
	}
	mockVault := &MockVaultClient{}
	mockMgmt := &MockManagementSvc{}
	svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

	rolesResponse := models.KeycloakRolesResponse{
		Roles: []models.KeycloakRole{
			{ID: "role-1", Name: "admin"},
		},
	}

	applicationsResponse := models.ApplicationsResponse{
		ApplicationDescriptors: []map[string]any{
			{"id": "app-1"},
		},
	}

	// Create 300 capability sets to test batch splitting (batch size is 250)
	capSets := make([]models.KeycloakCapabilitySet, 300)
	for i := 0; i < 300; i++ {
		capSets[i] = models.KeycloakCapabilitySet{
			ID:   fmt.Sprintf("cap-%d", i),
			Name: fmt.Sprintf("capability-%d", i),
		}
	}
	capSetsResponse := models.KeycloakCapabilitySetsResponse{CapabilitySets: capSets}

	mockHTTP.On("GetRetryReturnStruct",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/roles?offset=0&limit=10000")
		}),
		mock.Anything,
		mock.Anything).
		Run(func(args mock.Arguments) {
			target := args.Get(2).(*models.KeycloakRolesResponse)
			*target = rolesResponse
		}).
		Return(nil)

	mockMgmt.On("GetApplications").Return(applicationsResponse, nil)

	mockHTTP.On("GetRetryReturnStruct",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/capability-sets?query=applicationId==app-1")
		}),
		mock.Anything,
		mock.Anything).
		Run(func(args mock.Arguments) {
			target := args.Get(2).(*models.KeycloakCapabilitySetsResponse)
			*target = capSetsResponse
		}).
		Return(nil)

	// Expect 2 batches: 250 + 50
	mockHTTP.On("PostRetryReturnNoContent",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/roles/capability-sets")
		}),
		mock.Anything,
		mock.Anything).
		Return(nil).Times(2)

	// Act
	err := svc.AttachCapabilitySetsToRoles("test-tenant")

	// Assert
	assert.NoError(t, err)
	mockHTTP.AssertExpectations(t)
	mockMgmt.AssertExpectations(t)
}

func TestAttachCapabilitySetsToRoles_SkipsDifferentTenant(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "test-token"
	action.ConfigRoles = map[string]any{
		"admin": map[string]any{
			"tenant":          "different-tenant",
			"capability-sets": []any{"users.read"},
		},
	}
	mockVault := &MockVaultClient{}
	mockMgmt := &MockManagementSvc{}
	svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

	rolesResponse := models.KeycloakRolesResponse{
		Roles: []models.KeycloakRole{
			{ID: "role-1", Name: "admin"},
		},
	}

	mockHTTP.On("GetRetryReturnStruct",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/roles?offset=0&limit=10000")
		}),
		mock.Anything,
		mock.Anything).
		Run(func(args mock.Arguments) {
			target := args.Get(2).(*models.KeycloakRolesResponse)
			*target = rolesResponse
		}).
		Return(nil)

	// Act
	err := svc.AttachCapabilitySetsToRoles("test-tenant")

	// Assert
	assert.NoError(t, err)
	mockHTTP.AssertExpectations(t)
	// Should not call GetCapabilitySetsByName or PostRetryReturnNoContent
}

func TestAttachCapabilitySetsToRoles_SkipsRoleNotInConfig(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "test-token"
	action.ConfigRoles = map[string]any{}
	mockVault := &MockVaultClient{}
	mockMgmt := &MockManagementSvc{}
	svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

	rolesResponse := models.KeycloakRolesResponse{
		Roles: []models.KeycloakRole{
			{ID: "role-1", Name: "admin"},
		},
	}

	mockHTTP.On("GetRetryReturnStruct",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/roles?offset=0&limit=10000")
		}),
		mock.Anything,
		mock.Anything).
		Run(func(args mock.Arguments) {
			target := args.Get(2).(*models.KeycloakRolesResponse)
			*target = rolesResponse
		}).
		Return(nil)

	// Act
	err := svc.AttachCapabilitySetsToRoles("test-tenant")

	// Assert
	assert.NoError(t, err)
	mockHTTP.AssertExpectations(t)
}

func TestAttachCapabilitySetsToRoles_NoCapabilitySetsFound(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "test-token"
	action.ConfigRoles = map[string]any{
		"admin": map[string]any{
			"tenant":          "test-tenant",
			"capability-sets": []any{"nonexistent"},
		},
	}
	mockVault := &MockVaultClient{}
	mockMgmt := &MockManagementSvc{}
	svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

	rolesResponse := models.KeycloakRolesResponse{
		Roles: []models.KeycloakRole{
			{ID: "role-1", Name: "admin"},
		},
	}

	capSetsResponse := models.KeycloakCapabilitySetsResponse{CapabilitySets: []models.KeycloakCapabilitySet{}}

	mockHTTP.On("GetRetryReturnStruct",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/roles?offset=0&limit=10000")
		}),
		mock.Anything,
		mock.Anything).
		Run(func(args mock.Arguments) {
			target := args.Get(2).(*models.KeycloakRolesResponse)
			*target = rolesResponse
		}).
		Return(nil)

	mockHTTP.On("GetRetryReturnStruct",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/capability-sets?query=name==nonexistent")
		}),
		mock.Anything,
		mock.Anything).
		Run(func(args mock.Arguments) {
			target := args.Get(2).(*models.KeycloakCapabilitySetsResponse)
			*target = capSetsResponse
		}).
		Return(nil)

	// Act
	err := svc.AttachCapabilitySetsToRoles("test-tenant")

	// Assert
	assert.NoError(t, err)
	mockHTTP.AssertExpectations(t)
	// Should not call PostRetryReturnNoContent since no capability sets were found
}

func TestAttachCapabilitySetsToRoles_PopulateCapabilitySetsError(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "test-token"
	action.ConfigRoles = map[string]any{
		"admin": map[string]any{
			"tenant":          "test-tenant",
			"capability-sets": []any{"users.read"},
		},
	}
	mockVault := &MockVaultClient{}
	mockMgmt := &MockManagementSvc{}
	svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

	rolesResponse := models.KeycloakRolesResponse{
		Roles: []models.KeycloakRole{
			{ID: "role-1", Name: "admin"},
		},
	}

	expectedError := errors.New("HTTP error")
	mockHTTP.On("GetRetryReturnStruct",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/roles?offset=0&limit=10000")
		}),
		mock.Anything,
		mock.Anything).
		Run(func(args mock.Arguments) {
			target := args.Get(2).(*models.KeycloakRolesResponse)
			*target = rolesResponse
		}).
		Return(nil)

	mockHTTP.On("GetRetryReturnStruct",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/capability-sets?query=name==users.read")
		}),
		mock.Anything,
		mock.Anything).
		Return(expectedError)

	// Act
	err := svc.AttachCapabilitySetsToRoles("test-tenant")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockHTTP.AssertExpectations(t)
}

func TestAttachCapabilitySetsToRoles_PostError(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	action.KeycloakAccessToken = "test-token"
	action.ConfigRoles = map[string]any{
		"admin": map[string]any{
			"tenant":          "test-tenant",
			"capability-sets": []any{"users.read"},
		},
	}
	mockVault := &MockVaultClient{}
	mockMgmt := &MockManagementSvc{}
	svc := keycloaksvc.New(action, mockHTTP, mockVault, mockMgmt)

	rolesResponse := models.KeycloakRolesResponse{
		Roles: []models.KeycloakRole{
			{ID: "role-1", Name: "admin"},
		},
	}

	capSetsResponse := models.KeycloakCapabilitySetsResponse{
		CapabilitySets: []models.KeycloakCapabilitySet{
			{ID: "cap-1", Name: "users.read"},
		},
	}

	expectedError := errors.New("Post failed")
	mockHTTP.On("GetRetryReturnStruct",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/roles?offset=0&limit=10000")
		}),
		mock.Anything,
		mock.Anything).
		Run(func(args mock.Arguments) {
			target := args.Get(2).(*models.KeycloakRolesResponse)
			*target = rolesResponse
		}).
		Return(nil)

	mockHTTP.On("GetRetryReturnStruct",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/capability-sets?query=name==users.read")
		}),
		mock.Anything,
		mock.Anything).
		Run(func(args mock.Arguments) {
			target := args.Get(2).(*models.KeycloakCapabilitySetsResponse)
			*target = capSetsResponse
		}).
		Return(nil)

	mockHTTP.On("PostRetryReturnNoContent",
		mock.Anything,
		mock.Anything,
		mock.Anything).
		Return(expectedError)

	// Act
	err := svc.AttachCapabilitySetsToRoles("test-tenant")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockHTTP.AssertExpectations(t)
}
