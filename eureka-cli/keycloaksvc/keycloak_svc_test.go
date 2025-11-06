package keycloaksvc_test

import (
  "context"
  "encoding/json"
  "errors"
  "net/url"
  "strings"
  "testing"

  "github.com/folio-org/eureka-cli/constant"
  "github.com/folio-org/eureka-cli/internal/testhelpers"
  "github.com/folio-org/eureka-cli/keycloaksvc"
  "github.com/folio-org/eureka-cli/models"
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
  return args.Get(0).(models.ApplicationsResponse), args.Error(1)
}

func (m *MockManagementSvc) CreateApplications(extract *models.RegistryExtract) error {
  args := m.Called(extract)
  return args.Error(0)
}

func (m *MockManagementSvc) RemoveApplication(applicationID string) error {
  args := m.Called(applicationID)
  return args.Error(0)
}

func (m *MockManagementSvc) GetModuleDiscovery(name string) (models.ModuleDiscoveryResponse, error) {
  args := m.Called(name)
  return args.Get(0).(models.ModuleDiscoveryResponse), args.Error(1)
}

func (m *MockManagementSvc) UpdateModuleDiscovery(id string, restore bool, privatePort int, sidecarURL string) error {
  args := m.Called(id, restore, privatePort, sidecarURL)
  return args.Error(0)
}

func (m *MockManagementSvc) CreateTenantEntitlement(consortiumName string, tenantType constant.TenantType) error {
  args := m.Called(consortiumName, tenantType)
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

func TestGetKeycloakMasterAccessToken_Success(t *testing.T) {
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
  token, err := svc.GetKeycloakMasterAccessToken()

  // Assert
  assert.NoError(t, err)
  assert.Equal(t, expectedToken, token)
  mockHTTP.AssertExpectations(t)
}

func TestGetKeycloakMasterAccessToken_HTTPError(t *testing.T) {
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
  token, err := svc.GetKeycloakMasterAccessToken()

  // Assert
  assert.Error(t, err)
  assert.Equal(t, expectedError, err)
  assert.Empty(t, token)
  mockHTTP.AssertExpectations(t)
}

func TestGetKeycloakMasterAccessToken_NoToken(t *testing.T) {
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
  token, err := svc.GetKeycloakMasterAccessToken()

  // Assert
  assert.Error(t, err)
  assert.Contains(t, err.Error(), "access token")
  assert.Empty(t, token)
  mockHTTP.AssertExpectations(t)
}

func TestUpdateKeycloakPublicClientParams_Success(t *testing.T) {
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
  err := svc.UpdateKeycloakPublicClientParams(tenantName, baseURL)

  // Assert
  assert.NoError(t, err)
  mockHTTP.AssertExpectations(t)
}

func TestUpdateKeycloakPublicClientParams_HTTPError(t *testing.T) {
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
  err := svc.UpdateKeycloakPublicClientParams("test-tenant", "http://test.com")

  // Assert
  assert.Error(t, err)
  assert.Equal(t, expectedError, err)
  mockHTTP.AssertExpectations(t)
}
