package registrysvc_test

import (
	"errors"
	"testing"

	"github.com/folio-org/eureka-setup/eureka-cli/constant"
	"github.com/folio-org/eureka-setup/eureka-cli/internal/testhelpers"
	"github.com/folio-org/eureka-setup/eureka-cli/models"
	"github.com/folio-org/eureka-setup/eureka-cli/registrysvc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAWSSvc is a mock for awssvc.AWSProcessor
type MockAWSSvc struct {
	mock.Mock
}

func (m *MockAWSSvc) GetAuthorizationToken() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockAWSSvc) GetECRNamespace() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockAWSSvc) IsECRConfigured() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockAWSSvc) GetRegion() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockAWSSvc) GetECRRepositoryURI(namespace string, repositoryName string) (string, error) {
	args := m.Called(namespace, repositoryName)
	return args.String(0), args.Error(1)
}

// TestGetAuthorizationToken_Success tests successful token retrieval
func TestGetAuthorizationToken_Success(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	mockAWS := &MockAWSSvc{}
	action := testhelpers.NewMockAction()
	svc := registrysvc.New(action, mockHTTP, mockAWS)

	expectedToken := "auth-token-123"
	mockAWS.On("GetAuthorizationToken").Return(expectedToken, nil)

	// Act
	token, err := svc.GetAuthorizationToken()

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedToken, token)
	mockAWS.AssertExpectations(t)
}

func TestGetAuthorizationToken_Error(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	mockAWS := &MockAWSSvc{}
	action := testhelpers.NewMockAction()
	svc := registrysvc.New(action, mockHTTP, mockAWS)

	mockAWS.On("GetAuthorizationToken").Return("", errors.New("AWS error"))

	// Act
	token, err := svc.GetAuthorizationToken()

	// Assert
	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Contains(t, err.Error(), "AWS error")
	mockAWS.AssertExpectations(t)
}

func TestGetNamespace_WithECRNamespace(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	mockAWS := &MockAWSSvc{}
	action := testhelpers.NewMockAction()
	svc := registrysvc.New(action, mockHTTP, mockAWS)

	ecrNamespace := "my-ecr-namespace"
	mockAWS.On("GetECRNamespace").Return(ecrNamespace)

	// Act
	namespace := svc.GetNamespace("1.0.0")

	// Assert
	assert.Equal(t, ecrNamespace, namespace)
	mockAWS.AssertExpectations(t)
}

func TestGetNamespace_SnapshotVersion(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	mockAWS := &MockAWSSvc{}
	action := testhelpers.NewMockAction()
	svc := registrysvc.New(action, mockHTTP, mockAWS)

	mockAWS.On("GetECRNamespace").Return("")

	// Act
	namespace := svc.GetNamespace("1.0.0-SNAPSHOT")

	// Assert
	assert.Equal(t, constant.SnapshotNamespace, namespace)
	mockAWS.AssertExpectations(t)
}

func TestGetNamespace_ReleaseVersion(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	mockAWS := &MockAWSSvc{}
	action := testhelpers.NewMockAction()
	svc := registrysvc.New(action, mockHTTP, mockAWS)

	mockAWS.On("GetECRNamespace").Return("")

	// Act
	namespace := svc.GetNamespace("1.0.0")

	// Assert
	assert.Equal(t, constant.ReleaseNamespace, namespace)
	mockAWS.AssertExpectations(t)
}

// ==================== GetModules (LSP-based) Tests ====================

func buildLSPResponse(required, optional, experimental []models.PlatformApplication, components []models.PlatformApplication) models.PlatformDescriptor {
	return models.PlatformDescriptor{
		EurekaComponents: components,
		Applications: models.PlatformApplications{
			Required:     required,
			Optional:     optional,
			Experimental: experimental,
		},
	}
}

func stubLSP(mockHTTP *testhelpers.MockHTTPClient, lspURL string, descriptor models.PlatformDescriptor) {
	mockHTTP.On("GetRetryReturnStruct", lspURL, mock.Anything, mock.AnythingOfType("*models.PlatformDescriptor")).
		Run(func(args mock.Arguments) {
			ptr := args.Get(2).(*models.PlatformDescriptor)
			*ptr = descriptor
		}).Return(nil)
}

func stubFAR(mockHTTP *testhelpers.MockHTTPClient, farBase string, appName, appVersion string, mods []any) {
	appID := appName + "-" + appVersion
	url := farBase + "/applications?query=id==" + appID
	resp := models.ApplicationsResponse{ApplicationDescriptors: []map[string]any{{"modules": mods}}}
	mockHTTP.On("GetRetryReturnStruct", url, mock.Anything, mock.AnythingOfType("*models.ApplicationsResponse")).
		Run(func(args mock.Arguments) {
			ptr := args.Get(2).(*models.ApplicationsResponse)
			*ptr = resp
		}).Return(nil)
}

func stubFARWithUI(mockHTTP *testhelpers.MockHTTPClient, farBase string, appName, appVersion string, mods, uiMods []any) {
	appID := appName + "-" + appVersion
	url := farBase + "/applications?query=id==" + appID
	resp := models.ApplicationsResponse{ApplicationDescriptors: []map[string]any{{"modules": mods, "uiModules": uiMods}}}
	mockHTTP.On("GetRetryReturnStruct", url, mock.Anything, mock.AnythingOfType("*models.ApplicationsResponse")).
		Run(func(args mock.Arguments) {
			ptr := args.Get(2).(*models.ApplicationsResponse)
			*ptr = resp
		}).Return(nil)
}

func TestGetModules_Success(t *testing.T) {
	mockHTTP := &testhelpers.MockHTTPClient{}
	mockAWS := &MockAWSSvc{}
	act := testhelpers.NewMockAction()
	act.ConfigLspURL = "http://lsp.example.com/descriptor.json"
	act.ConfigFarURL = "http://far.example.com"
	svc := registrysvc.New(act, mockHTTP, mockAWS)

	descriptor := buildLSPResponse(
		[]models.PlatformApplication{{Name: "app-core", Version: "1.0.0"}},
		nil, nil, nil,
	)
	stubLSP(mockHTTP, act.ConfigLspURL, descriptor)
	stubFAR(mockHTTP, act.ConfigFarURL, "app-core", "1.0.0", []any{
		map[string]any{"id": "mod-inventory-1.0.0", "name": "mod-inventory", "version": "1.0.0"},
		map[string]any{"id": "mod-users-2.0.0", "name": "mod-users", "version": "2.0.0"},
	})

	result, err := svc.GetModules(false)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.FolioModules, 2)
	assert.Empty(t, result.EurekaModules)
	mockHTTP.AssertExpectations(t)
}

func TestGetModules_Verbose(t *testing.T) {
	mockHTTP := &testhelpers.MockHTTPClient{}
	mockAWS := &MockAWSSvc{}
	act := testhelpers.NewMockAction()
	act.ConfigLspURL = "http://lsp.example.com/descriptor.json"
	act.ConfigFarURL = "http://far.example.com"
	svc := registrysvc.New(act, mockHTTP, mockAWS)

	descriptor := buildLSPResponse(
		[]models.PlatformApplication{{Name: "app-core", Version: "1.0.0"}},
		nil, nil, nil,
	)
	stubLSP(mockHTTP, act.ConfigLspURL, descriptor)
	stubFAR(mockHTTP, act.ConfigFarURL, "app-core", "1.0.0", []any{
		map[string]any{"id": "mod-inventory-1.0.0", "name": "mod-inventory", "version": "1.0.0"},
	})

	result, err := svc.GetModules(true)

	assert.NoError(t, err)
	assert.Len(t, result.FolioModules, 1)
	mockHTTP.AssertExpectations(t)
}

func TestGetModules_LSPFetchError(t *testing.T) {
	mockHTTP := &testhelpers.MockHTTPClient{}
	mockAWS := &MockAWSSvc{}
	act := testhelpers.NewMockAction()
	act.ConfigLspURL = "http://lsp.example.com/descriptor.json"
	act.ConfigFarURL = "http://far.example.com"
	svc := registrysvc.New(act, mockHTTP, mockAWS)

	mockHTTP.On("GetRetryReturnStruct", act.ConfigLspURL, mock.Anything, mock.AnythingOfType("*models.PlatformDescriptor")).
		Return(errors.New("LSP unreachable"))

	result, err := svc.GetModules(false)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "LSP unreachable")
	mockHTTP.AssertExpectations(t)
}

func TestGetModules_FARFetchError(t *testing.T) {
	mockHTTP := &testhelpers.MockHTTPClient{}
	mockAWS := &MockAWSSvc{}
	act := testhelpers.NewMockAction()
	act.ConfigLspURL = "http://lsp.example.com/descriptor.json"
	act.ConfigFarURL = "http://far.example.com"
	svc := registrysvc.New(act, mockHTTP, mockAWS)

	descriptor := buildLSPResponse(
		[]models.PlatformApplication{{Name: "app-core", Version: "1.0.0"}},
		nil, nil, nil,
	)
	stubLSP(mockHTTP, act.ConfigLspURL, descriptor)

	farURL := act.ConfigFarURL + "/applications?query=id==app-core-1.0.0"
	mockHTTP.On("GetRetryReturnStruct", farURL, mock.Anything, mock.AnythingOfType("*models.ApplicationsResponse")).
		Return(errors.New("FAR timeout"))

	result, err := svc.GetModules(false)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "app-core-1.0.0")
	mockHTTP.AssertExpectations(t)
}

func TestGetModules_EurekaComponentsIncluded(t *testing.T) {
	mockHTTP := &testhelpers.MockHTTPClient{}
	mockAWS := &MockAWSSvc{}
	act := testhelpers.NewMockAction()
	act.ConfigLspURL = "http://lsp.example.com/descriptor.json"
	act.ConfigFarURL = "http://far.example.com"
	svc := registrysvc.New(act, mockHTTP, mockAWS)

	descriptor := buildLSPResponse(nil, nil, nil, []models.PlatformApplication{
		{Name: "mgr-tenants", Version: "1.0.0"},
		{Name: "folio-kong", Version: "3.0.0"},
	})
	stubLSP(mockHTTP, act.ConfigLspURL, descriptor)

	result, err := svc.GetModules(false)

	assert.NoError(t, err)
	assert.Empty(t, result.FolioModules)
	assert.Len(t, result.EurekaModules, 2)
	mockHTTP.AssertExpectations(t)
}

func TestGetModules_ModulePartitioning(t *testing.T) {
	mockHTTP := &testhelpers.MockHTTPClient{}
	mockAWS := &MockAWSSvc{}
	act := testhelpers.NewMockAction()
	act.ConfigLspURL = "http://lsp.example.com/descriptor.json"
	act.ConfigFarURL = "http://far.example.com"
	svc := registrysvc.New(act, mockHTTP, mockAWS)

	descriptor := buildLSPResponse(
		[]models.PlatformApplication{{Name: "app-mixed", Version: "1.0.0"}},
		nil, nil, nil,
	)
	stubLSP(mockHTTP, act.ConfigLspURL, descriptor)
	stubFAR(mockHTTP, act.ConfigFarURL, "app-mixed", "1.0.0", []any{
		map[string]any{"id": "mod-inventory-1.0.0", "name": "mod-inventory", "version": "1.0.0"},
		map[string]any{"id": "mgr-tenants-1.0.0", "name": "mgr-tenants", "version": "1.0.0"},
		map[string]any{"id": "mod-login-keycloak-1.0.0", "name": "mod-login-keycloak", "version": "1.0.0"},
		map[string]any{"id": "folio-kong-3.0.0", "name": "folio-kong", "version": "3.0.0"},
		map[string]any{"id": "folio-module-sidecar-1.0.0", "name": "folio-module-sidecar", "version": "1.0.0"},
		map[string]any{"id": "mod-scheduler-2.0.0", "name": "mod-scheduler", "version": "2.0.0"},
	})

	result, err := svc.GetModules(false)

	assert.NoError(t, err)
	assert.Len(t, result.FolioModules, 1)
	assert.Len(t, result.EurekaModules, 5)
	assert.Equal(t, "mod-inventory-1.0.0", result.FolioModules[0].ID)
	mockHTTP.AssertExpectations(t)
}

func TestGetModules_RequiredAndOptionalMerged(t *testing.T) {
	mockHTTP := &testhelpers.MockHTTPClient{}
	mockAWS := &MockAWSSvc{}
	act := testhelpers.NewMockAction()
	act.ConfigLspURL = "http://lsp.example.com/descriptor.json"
	act.ConfigFarURL = "http://far.example.com"
	svc := registrysvc.New(act, mockHTTP, mockAWS)

	descriptor := buildLSPResponse(
		[]models.PlatformApplication{{Name: "app-required", Version: "1.0.0"}},
		[]models.PlatformApplication{{Name: "app-optional", Version: "2.0.0"}},
		nil, nil,
	)
	stubLSP(mockHTTP, act.ConfigLspURL, descriptor)
	stubFAR(mockHTTP, act.ConfigFarURL, "app-required", "1.0.0", []any{
		map[string]any{"id": "mod-alpha-1.0.0", "name": "mod-alpha", "version": "1.0.0"},
	})
	stubFAR(mockHTTP, act.ConfigFarURL, "app-optional", "2.0.0", []any{
		map[string]any{"id": "mod-beta-2.0.0", "name": "mod-beta", "version": "2.0.0"},
	})

	result, err := svc.GetModules(false)

	assert.NoError(t, err)
	assert.Len(t, result.FolioModules, 2)
	mockHTTP.AssertExpectations(t)
}

func TestGetModules_ExperimentalAppsIncluded(t *testing.T) {
	mockHTTP := &testhelpers.MockHTTPClient{}
	mockAWS := &MockAWSSvc{}
	act := testhelpers.NewMockAction()
	act.ConfigLspURL = "http://lsp.example.com/descriptor.json"
	act.ConfigFarURL = "http://far.example.com"
	svc := registrysvc.New(act, mockHTTP, mockAWS)

	descriptor := buildLSPResponse(
		[]models.PlatformApplication{{Name: "app-required", Version: "1.0.0"}},
		nil,
		[]models.PlatformApplication{{Name: "app-experimental", Version: "3.0.0"}},
		nil,
	)
	stubLSP(mockHTTP, act.ConfigLspURL, descriptor)
	stubFAR(mockHTTP, act.ConfigFarURL, "app-required", "1.0.0", []any{
		map[string]any{"id": "mod-inventory-1.0.0", "name": "mod-inventory", "version": "1.0.0"},
	})
	stubFAR(mockHTTP, act.ConfigFarURL, "app-experimental", "3.0.0", []any{
		map[string]any{"id": "mod-linked-data-3.0.0", "name": "mod-linked-data", "version": "3.0.0"},
		map[string]any{"id": "mod-inn-reach-3.0.0", "name": "mod-inn-reach", "version": "3.0.0"},
	})

	result, err := svc.GetModules(false)

	assert.NoError(t, err)
	assert.Len(t, result.FolioModules, 3)
	assert.Empty(t, result.EurekaModules)
	mockHTTP.AssertExpectations(t)
}

func TestGetModules_UIModulesIncluded(t *testing.T) {
	mockHTTP := &testhelpers.MockHTTPClient{}
	mockAWS := &MockAWSSvc{}
	act := testhelpers.NewMockAction()
	act.ConfigLspURL = "http://lsp.example.com/descriptor.json"
	act.ConfigFarURL = "http://far.example.com"
	svc := registrysvc.New(act, mockHTTP, mockAWS)

	descriptor := buildLSPResponse(
		[]models.PlatformApplication{{Name: "app-ui", Version: "1.0.0"}},
		nil, nil, nil,
	)
	stubLSP(mockHTTP, act.ConfigLspURL, descriptor)
	stubFARWithUI(mockHTTP, act.ConfigFarURL, "app-ui", "1.0.0",
		[]any{
			map[string]any{"id": "mod-inventory-1.0.0", "name": "mod-inventory", "version": "1.0.0"},
		},
		[]any{
			map[string]any{"id": "folio_inventory-1.0.0", "name": "folio_inventory", "version": "1.0.0"},
			map[string]any{"id": "folio_users-2.0.0", "name": "folio_users", "version": "2.0.0"},
		},
	)

	result, err := svc.GetModules(false)

	assert.NoError(t, err)
	assert.Len(t, result.FolioModules, 3)
	assert.Empty(t, result.EurekaModules)
	mockHTTP.AssertExpectations(t)
}

func TestGetModules_EmptyApplications(t *testing.T) {
	mockHTTP := &testhelpers.MockHTTPClient{}
	mockAWS := &MockAWSSvc{}
	act := testhelpers.NewMockAction()
	act.ConfigLspURL = "http://lsp.example.com/descriptor.json"
	act.ConfigFarURL = "http://far.example.com"
	svc := registrysvc.New(act, mockHTTP, mockAWS)

	descriptor := buildLSPResponse(nil, nil, nil, []models.PlatformApplication{
		{Name: "folio-module-sidecar", Version: "1.0.0"},
	})
	stubLSP(mockHTTP, act.ConfigLspURL, descriptor)

	result, err := svc.GetModules(false)

	assert.NoError(t, err)
	assert.Empty(t, result.FolioModules)
	assert.Len(t, result.EurekaModules, 1)
	mockHTTP.AssertExpectations(t)
}

// ==================== isEurekaModule Tests ====================

func TestIsEurekaModule_KeycloakSuffix(t *testing.T) {
	mockHTTP := &testhelpers.MockHTTPClient{}
	mockAWS := &MockAWSSvc{}
	act := testhelpers.NewMockAction()
	act.ConfigLspURL = "http://lsp.example.com/descriptor.json"
	act.ConfigFarURL = "http://far.example.com"
	svc := registrysvc.New(act, mockHTTP, mockAWS)

	descriptor := buildLSPResponse(
		[]models.PlatformApplication{{Name: "app-kc", Version: "1.0.0"}},
		nil, nil, nil,
	)
	stubLSP(mockHTTP, act.ConfigLspURL, descriptor)
	stubFAR(mockHTTP, act.ConfigFarURL, "app-kc", "1.0.0", []any{
		map[string]any{"id": "mod-auth-keycloak-1.0.0", "name": "mod-auth-keycloak", "version": "1.0.0"},
	})

	result, err := svc.GetModules(false)
	assert.NoError(t, err)
	assert.Empty(t, result.FolioModules)
	assert.Len(t, result.EurekaModules, 1)
}

func TestIsEurekaModule_MgrPrefix(t *testing.T) {
	mockHTTP := &testhelpers.MockHTTPClient{}
	mockAWS := &MockAWSSvc{}
	act := testhelpers.NewMockAction()
	act.ConfigLspURL = "http://lsp.example.com/descriptor.json"
	act.ConfigFarURL = "http://far.example.com"
	svc := registrysvc.New(act, mockHTTP, mockAWS)

	descriptor := buildLSPResponse(
		[]models.PlatformApplication{{Name: "app-mgr", Version: "1.0.0"}},
		nil, nil, nil,
	)
	stubLSP(mockHTTP, act.ConfigLspURL, descriptor)
	stubFAR(mockHTTP, act.ConfigFarURL, "app-mgr", "1.0.0", []any{
		map[string]any{"id": "mgr-applications-1.0.0", "name": "mgr-applications", "version": "1.0.0"},
		map[string]any{"id": "mgr-tenants-1.0.0", "name": "mgr-tenants", "version": "1.0.0"},
	})

	result, err := svc.GetModules(false)
	assert.NoError(t, err)
	assert.Empty(t, result.FolioModules)
	assert.Len(t, result.EurekaModules, 2)
}

func TestIsEurekaModule_ExactMatches(t *testing.T) {
	mockHTTP := &testhelpers.MockHTTPClient{}
	mockAWS := &MockAWSSvc{}
	act := testhelpers.NewMockAction()
	act.ConfigLspURL = "http://lsp.example.com/descriptor.json"
	act.ConfigFarURL = "http://far.example.com"
	svc := registrysvc.New(act, mockHTTP, mockAWS)

	descriptor := buildLSPResponse(
		[]models.PlatformApplication{{Name: "app-exact", Version: "1.0.0"}},
		nil, nil, nil,
	)
	stubLSP(mockHTTP, act.ConfigLspURL, descriptor)
	stubFAR(mockHTTP, act.ConfigFarURL, "app-exact", "1.0.0", []any{
		map[string]any{"id": "folio-kong-3.0.0", "name": "folio-kong", "version": "3.0.0"},
		map[string]any{"id": "folio-module-sidecar-1.0.0", "name": "folio-module-sidecar", "version": "1.0.0"},
		map[string]any{"id": "mod-scheduler-2.0.0", "name": "mod-scheduler", "version": "2.0.0"},
	})

	result, err := svc.GetModules(false)
	assert.NoError(t, err)
	assert.Empty(t, result.FolioModules)
	assert.Len(t, result.EurekaModules, 3)
}

func TestIsEurekaModule_RegularFolioModuleNotEureka(t *testing.T) {
	mockHTTP := &testhelpers.MockHTTPClient{}
	mockAWS := &MockAWSSvc{}
	act := testhelpers.NewMockAction()
	act.ConfigLspURL = "http://lsp.example.com/descriptor.json"
	act.ConfigFarURL = "http://far.example.com"
	svc := registrysvc.New(act, mockHTTP, mockAWS)

	descriptor := buildLSPResponse(
		[]models.PlatformApplication{{Name: "app-folio", Version: "1.0.0"}},
		nil, nil, nil,
	)
	stubLSP(mockHTTP, act.ConfigLspURL, descriptor)
	stubFAR(mockHTTP, act.ConfigFarURL, "app-folio", "1.0.0", []any{
		map[string]any{"id": "mod-inventory-1.0.0", "name": "mod-inventory", "version": "1.0.0"},
		map[string]any{"id": "mod-users-2.0.0", "name": "mod-users", "version": "2.0.0"},
		map[string]any{"id": "edge-patron-1.0.0", "name": "edge-patron", "version": "1.0.0"},
	})

	result, err := svc.GetModules(false)
	assert.NoError(t, err)
	assert.Len(t, result.FolioModules, 3)
	assert.Empty(t, result.EurekaModules)
}

func TestExtractModuleMetadata_Success(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	mockAWS := &MockAWSSvc{}
	action := testhelpers.NewMockAction()
	svc := registrysvc.New(action, mockHTTP, mockAWS)

	version1 := "1.0.0"
	version2 := "2.0.0"

	modules := &models.ProxyModulesByRegistry{
		FolioModules: []*models.ProxyModule{
			{ID: "mod-inventory-1.0.0"},
			{ID: "mod-users-2.0.0"},
		},
		EurekaModules: []*models.ProxyModule{
			{ID: "mod-custom-1.5.0"},
		},
	}

	// Act
	svc.ExtractModuleMetadata(modules)

	// Assert
	// Check FOLIO modules
	assert.Equal(t, "mod-inventory", modules.FolioModules[0].Metadata.Name)
	assert.Equal(t, &version1, modules.FolioModules[0].Metadata.Version)
	assert.Equal(t, "mod-inventory-sc", modules.FolioModules[0].Metadata.SidecarName)

	assert.Equal(t, "mod-users", modules.FolioModules[1].Metadata.Name)
	assert.Equal(t, &version2, modules.FolioModules[1].Metadata.Version)
	assert.Equal(t, "mod-users-sc", modules.FolioModules[1].Metadata.SidecarName)

	// Check Eureka modules
	version3 := "1.5.0"
	assert.Equal(t, "mod-custom", modules.EurekaModules[0].Metadata.Name)
	assert.Equal(t, &version3, modules.EurekaModules[0].Metadata.Version)
	assert.Equal(t, "mod-custom-sc", modules.EurekaModules[0].Metadata.SidecarName)
}

func TestExtractModuleMetadata_EdgeModule(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	mockAWS := &MockAWSSvc{}
	action := testhelpers.NewMockAction()
	svc := registrysvc.New(action, mockHTTP, mockAWS)

	modules := &models.ProxyModulesByRegistry{
		FolioModules: []*models.ProxyModule{
			{ID: "edge-patron-1.0.0"},
		},
		EurekaModules: nil,
	}

	// Act
	svc.ExtractModuleMetadata(modules)

	// Assert
	// Edge modules should have SidecarName equal to Name
	assert.Equal(t, "edge-patron", modules.FolioModules[0].Metadata.Name)
	assert.Equal(t, "edge-patron", modules.FolioModules[0].Metadata.SidecarName)
}

func TestExtractModuleMetadata_SkipsOkapi(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	mockAWS := &MockAWSSvc{}
	action := testhelpers.NewMockAction()
	svc := registrysvc.New(action, mockHTTP, mockAWS)

	modules := &models.ProxyModulesByRegistry{
		FolioModules: []*models.ProxyModule{
			{ID: "okapi"},
			{ID: "mod-users-1.0.0"},
		},
		EurekaModules: nil,
	}

	// Act
	svc.ExtractModuleMetadata(modules)

	// Assert
	// Okapi should remain unchanged
	assert.Empty(t, modules.FolioModules[0].Metadata.Name)
	assert.Nil(t, modules.FolioModules[0].Metadata.Version)
	assert.Empty(t, modules.FolioModules[0].Metadata.SidecarName)

	// Other module should be processed
	assert.Equal(t, "mod-users", modules.FolioModules[1].Metadata.Name)
}

func TestExtractModuleMetadata_EmptyModules(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	mockAWS := &MockAWSSvc{}
	action := testhelpers.NewMockAction()
	svc := registrysvc.New(action, mockHTTP, mockAWS)

	modules := &models.ProxyModulesByRegistry{FolioModules: nil, EurekaModules: nil}

	// Act & Assert - should not panic
	svc.ExtractModuleMetadata(modules)
}

func TestExtractModuleMetadata_ModuleWithoutVersion(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	mockAWS := &MockAWSSvc{}
	action := testhelpers.NewMockAction()
	svc := registrysvc.New(action, mockHTTP, mockAWS)

	modules := &models.ProxyModulesByRegistry{
		FolioModules: []*models.ProxyModule{
			{ID: "mod-users"},
		},
		EurekaModules: nil,
	}

	// Act
	svc.ExtractModuleMetadata(modules)

	// Assert
	// When module ID doesn't have a proper version, the regex parsing behavior
	// extracts what it can - "mod" as name and "-users" as the "version"
	// This tests the actual behavior of GetModuleNameFromID and GetOptionalModuleVersion
	assert.Equal(t, "mod", modules.FolioModules[0].Metadata.Name)
	assert.NotNil(t, modules.FolioModules[0].Metadata.Version)
	assert.Equal(t, "-users", *modules.FolioModules[0].Metadata.Version)
	assert.Equal(t, "mod-sc", modules.FolioModules[0].Metadata.SidecarName)
}

// Tests for getSidecarName

func TestGetSidecarName_StandardModule(t *testing.T) {
	t.Run("TestGetSidecarName_StandardModule", func(t *testing.T) {
		// Arrange
		mockHTTP := &testhelpers.MockHTTPClient{}
		mockAWS := &MockAWSSvc{}
		action := testhelpers.NewMockAction()
		svc := registrysvc.New(action, mockHTTP, mockAWS)

		module := &models.ProxyModule{
			ID: "mod-inventory-1.0.0",
		}

		// Act
		svc.ExtractModuleMetadata(&models.ProxyModulesByRegistry{
			FolioModules: []*models.ProxyModule{module},
		})

		// Assert - verify sidecar name follows standard pattern
		assert.Equal(t, "mod-inventory-sc", module.Metadata.SidecarName)
		assert.Equal(t, "mod-inventory", module.Metadata.Name)
	})
}

func TestGetSidecarName_EdgeModule(t *testing.T) {
	t.Run("TestGetSidecarName_EdgeModule", func(t *testing.T) {
		// Arrange
		mockHTTP := &testhelpers.MockHTTPClient{}
		mockAWS := &MockAWSSvc{}
		action := testhelpers.NewMockAction()
		svc := registrysvc.New(action, mockHTTP, mockAWS)

		module := &models.ProxyModule{
			ID: "edge-patron-1.0.0",
		}

		// Act
		svc.ExtractModuleMetadata(&models.ProxyModulesByRegistry{
			FolioModules: []*models.ProxyModule{module},
		})

		// Assert - edge modules should use name as sidecar name (no -sc suffix)
		assert.Equal(t, "edge-patron", module.Metadata.SidecarName)
		assert.Equal(t, "edge-patron", module.Metadata.Name)
	})
}

func TestGetSidecarName_EdgeOaiPmhModule(t *testing.T) {
	t.Run("TestGetSidecarName_EdgeOaiPmhModule", func(t *testing.T) {
		// Arrange
		mockHTTP := &testhelpers.MockHTTPClient{}
		mockAWS := &MockAWSSvc{}
		action := testhelpers.NewMockAction()
		svc := registrysvc.New(action, mockHTTP, mockAWS)

		module := &models.ProxyModule{
			ID: "edge-oai-pmh-2.0.0",
		}

		// Act
		svc.ExtractModuleMetadata(&models.ProxyModulesByRegistry{
			EurekaModules: []*models.ProxyModule{module},
		})

		// Assert - edge module should use name without -sc suffix
		assert.Equal(t, "edge-oai-pmh", module.Metadata.SidecarName)
		assert.Equal(t, "edge-oai-pmh", module.Metadata.Name)
	})
}

func TestGetSidecarName_EdgeRtacModule(t *testing.T) {
	t.Run("TestGetSidecarName_EdgeRtacModule", func(t *testing.T) {
		// Arrange
		mockHTTP := &testhelpers.MockHTTPClient{}
		mockAWS := &MockAWSSvc{}
		action := testhelpers.NewMockAction()
		svc := registrysvc.New(action, mockHTTP, mockAWS)

		module := &models.ProxyModule{
			ID: "edge-rtac-1.5.0",
		}

		// Act
		svc.ExtractModuleMetadata(&models.ProxyModulesByRegistry{
			FolioModules: []*models.ProxyModule{module},
		})

		// Assert - edge module should use name without -sc suffix
		assert.Equal(t, "edge-rtac", module.Metadata.SidecarName)
		assert.Equal(t, "edge-rtac", module.Metadata.Name)
	})
}

func TestGetSidecarName_NonEdgeModuleContainingEdge(t *testing.T) {
	t.Run("TestGetSidecarName_NonEdgeModuleContainingEdge", func(t *testing.T) {
		// Arrange
		mockHTTP := &testhelpers.MockHTTPClient{}
		mockAWS := &MockAWSSvc{}
		action := testhelpers.NewMockAction()
		svc := registrysvc.New(action, mockHTTP, mockAWS)

		module := &models.ProxyModule{
			ID: "mod-knowledge-1.0.0",
		}

		// Act
		svc.ExtractModuleMetadata(&models.ProxyModulesByRegistry{
			FolioModules: []*models.ProxyModule{module},
		})

		// Assert - should use standard sidecar naming (contains 'edge' but doesn't start with 'edge')
		assert.Equal(t, "mod-knowledge-sc", module.Metadata.SidecarName)
		assert.Equal(t, "mod-knowledge", module.Metadata.Name)
	})
}

func TestGetSidecarName_MultipleEdgeModules(t *testing.T) {
	t.Run("TestGetSidecarName_MultipleEdgeModules", func(t *testing.T) {
		// Arrange
		mockHTTP := &testhelpers.MockHTTPClient{}
		mockAWS := &MockAWSSvc{}
		action := testhelpers.NewMockAction()
		svc := registrysvc.New(action, mockHTTP, mockAWS)

		modules := &models.ProxyModulesByRegistry{
			FolioModules: []*models.ProxyModule{
				{ID: "edge-patron-1.0.0"},
				{ID: "edge-orders-2.0.0"},
				{ID: "mod-users-3.0.0"},
			},
		}

		// Act
		svc.ExtractModuleMetadata(modules)

		// Assert - verify all edge modules use name without -sc, non-edge use -sc
		assert.Equal(t, "edge-patron", modules.FolioModules[0].Metadata.SidecarName)
		assert.Equal(t, "edge-orders", modules.FolioModules[1].Metadata.SidecarName)
		assert.Equal(t, "mod-users-sc", modules.FolioModules[2].Metadata.SidecarName)
	})
}
