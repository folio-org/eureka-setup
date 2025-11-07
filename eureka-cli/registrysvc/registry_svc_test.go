package registrysvc_test

import (
	"errors"
	"testing"

	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/field"
	"github.com/folio-org/eureka-cli/internal/testhelpers"
	"github.com/folio-org/eureka-cli/models"
	"github.com/folio-org/eureka-cli/registrysvc"
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
	assert.Equal(t, constant.SnapshotRegistry, namespace)
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
	assert.Equal(t, constant.ReleaseRegistry, namespace)
	mockAWS.AssertExpectations(t)
}

func TestGetModules_Success(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	mockAWS := &MockAWSSvc{}
	action := testhelpers.NewMockAction()
	svc := registrysvc.New(action, mockHTTP, mockAWS)

	installJsonURLs := map[string]string{
		constant.FolioRegistry:  "http://folio.example.com/install.json",
		constant.EurekaRegistry: "http://eureka.example.com/install.json",
	}

	folioModules := []*models.ProxyModule{
		{ID: "mod-inventory-1.0.0", Action: "enable"},
		{ID: "mod-users-1.0.0", Action: "enable"},
	}

	eurekaModules := []*models.ProxyModule{
		{ID: "mod-custom-1.0.0", Action: "enable"},
	}

	mockHTTP.On("GetRetryReturnStruct",
		"http://folio.example.com/install.json",
		mock.Anything,
		mock.AnythingOfType("*[]*models.ProxyModule")).
		Run(func(args mock.Arguments) {
			arg := args.Get(2).(*[]*models.ProxyModule)
			*arg = folioModules
		}).
		Return(nil)

	mockHTTP.On("GetRetryReturnStruct",
		"http://eureka.example.com/install.json",
		mock.Anything,
		mock.AnythingOfType("*[]*models.ProxyModule")).
		Run(func(args mock.Arguments) {
			arg := args.Get(2).(*[]*models.ProxyModule)
			*arg = eurekaModules
		}).
		Return(nil)

	// Act
	result, err := svc.GetModules(installJsonURLs, false)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.FolioModules, 2)
	assert.Len(t, result.EurekaModules, 1)
	mockHTTP.AssertExpectations(t)
}

func TestGetModules_WithCustomFrontendModules(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	mockAWS := &MockAWSSvc{}
	action := testhelpers.NewMockAction()
	svc := registrysvc.New(action, mockHTTP, mockAWS)

	// Setup custom frontend modules
	action.ConfigCustomFrontendModules = map[string]any{
		"custom-ui": map[string]any{
			field.ModuleVersionEntry: "2.0.0",
		},
		"invalid-module": nil,
		"no-version": map[string]any{
			"other-field": "value",
		},
	}

	installJsonURLs := map[string]string{
		constant.FolioRegistry: "http://folio.example.com/install.json",
	}

	folioModules := []*models.ProxyModule{
		{ID: "mod-inventory-1.0.0", Action: "enable"},
	}

	mockHTTP.On("GetRetryReturnStruct",
		"http://folio.example.com/install.json",
		mock.Anything,
		mock.AnythingOfType("*[]*models.ProxyModule")).
		Run(func(args mock.Arguments) {
			arg := args.Get(2).(*[]*models.ProxyModule)
			*arg = folioModules
		}).
		Return(nil)

	// Act
	result, err := svc.GetModules(installJsonURLs, false)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	// Should have original module + 1 custom module (invalid-module and no-version are skipped)
	assert.Len(t, result.FolioModules, 2)
	// Verify custom module was added
	hasCustomUI := false
	for _, mod := range result.FolioModules {
		if mod.ID == "custom-ui-2.0.0" {
			hasCustomUI = true
			assert.Equal(t, "enable", mod.Action)
		}
	}
	assert.True(t, hasCustomUI, "custom-ui module should be present")
	mockHTTP.AssertExpectations(t)
}

func TestGetModules_HTTPError(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	mockAWS := &MockAWSSvc{}
	action := testhelpers.NewMockAction()
	svc := registrysvc.New(action, mockHTTP, mockAWS)

	installJsonURLs := map[string]string{
		constant.FolioRegistry: "http://folio.example.com/install.json",
	}

	mockHTTP.On("GetRetryReturnStruct",
		"http://folio.example.com/install.json",
		mock.Anything,
		mock.AnythingOfType("*[]*models.ProxyModule")).
		Return(errors.New("HTTP 500"))

	// Act
	result, err := svc.GetModules(installJsonURLs, false)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "HTTP 500")
	mockHTTP.AssertExpectations(t)
}

func TestGetModules_EmptyResults(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	mockAWS := &MockAWSSvc{}
	action := testhelpers.NewMockAction()
	svc := registrysvc.New(action, mockHTTP, mockAWS)

	installJsonURLs := map[string]string{
		constant.FolioRegistry: "http://folio.example.com/install.json",
	}

	emptyModules := []*models.ProxyModule{}

	mockHTTP.On("GetRetryReturnStruct",
		"http://folio.example.com/install.json",
		mock.Anything,
		mock.AnythingOfType("*[]*models.ProxyModule")).
		Run(func(args mock.Arguments) {
			arg := args.Get(2).(*[]*models.ProxyModule)
			*arg = emptyModules
		}).
		Return(nil)

	// Act
	result, err := svc.GetModules(installJsonURLs, false)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result.FolioModules)
	mockHTTP.AssertExpectations(t)
}

func TestGetModules_WithVerbose(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	mockAWS := &MockAWSSvc{}
	action := testhelpers.NewMockAction()
	svc := registrysvc.New(action, mockHTTP, mockAWS)

	installJsonURLs := map[string]string{
		constant.FolioRegistry: "http://folio.example.com/install.json",
	}

	folioModules := []*models.ProxyModule{
		{ID: "mod-inventory-1.0.0", Action: "enable"},
	}

	mockHTTP.On("GetRetryReturnStruct",
		"http://folio.example.com/install.json",
		mock.Anything,
		mock.AnythingOfType("*[]*models.ProxyModule")).
		Run(func(args mock.Arguments) {
			arg := args.Get(2).(*[]*models.ProxyModule)
			*arg = folioModules
		}).
		Return(nil)

	// Act
	result, err := svc.GetModules(installJsonURLs, true)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.FolioModules, 1)
	mockHTTP.AssertExpectations(t)
}

func TestExtractModuleMetadata_Success(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	mockAWS := &MockAWSSvc{}
	action := testhelpers.NewMockAction()
	svc := registrysvc.New(action, mockHTTP, mockAWS)

	version1 := "1.0.0"
	version2 := "2.0.0"

	modules := models.NewProxyModulesByRegistry(
		[]*models.ProxyModule{
			{ID: "mod-inventory-1.0.0"},
			{ID: "mod-users-2.0.0"},
		},
		[]*models.ProxyModule{
			{ID: "mod-custom-1.5.0"},
		},
	)

	// Act
	svc.ExtractModuleMetadata(modules)

	// Assert
	// Check FOLIO modules
	assert.Equal(t, "mod-inventory", modules.FolioModules[0].Name)
	assert.Equal(t, &version1, modules.FolioModules[0].Version)
	assert.Equal(t, "mod-inventory-sc", modules.FolioModules[0].SidecarName)

	assert.Equal(t, "mod-users", modules.FolioModules[1].Name)
	assert.Equal(t, &version2, modules.FolioModules[1].Version)
	assert.Equal(t, "mod-users-sc", modules.FolioModules[1].SidecarName)

	// Check Eureka modules
	version3 := "1.5.0"
	assert.Equal(t, "mod-custom", modules.EurekaModules[0].Name)
	assert.Equal(t, &version3, modules.EurekaModules[0].Version)
	assert.Equal(t, "mod-custom-sc", modules.EurekaModules[0].SidecarName)
}

func TestExtractModuleMetadata_EdgeModule(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	mockAWS := &MockAWSSvc{}
	action := testhelpers.NewMockAction()
	svc := registrysvc.New(action, mockHTTP, mockAWS)

	modules := models.NewProxyModulesByRegistry(
		[]*models.ProxyModule{
			{ID: "edge-patron-1.0.0"},
		},
		nil,
	)

	// Act
	svc.ExtractModuleMetadata(modules)

	// Assert
	// Edge modules should have SidecarName equal to Name
	assert.Equal(t, "edge-patron", modules.FolioModules[0].Name)
	assert.Equal(t, "edge-patron", modules.FolioModules[0].SidecarName)
}

func TestExtractModuleMetadata_SkipsOkapi(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	mockAWS := &MockAWSSvc{}
	action := testhelpers.NewMockAction()
	svc := registrysvc.New(action, mockHTTP, mockAWS)

	modules := models.NewProxyModulesByRegistry(
		[]*models.ProxyModule{
			{ID: "okapi"},
			{ID: "mod-users-1.0.0"},
		},
		nil,
	)

	// Act
	svc.ExtractModuleMetadata(modules)

	// Assert
	// Okapi should remain unchanged
	assert.Empty(t, modules.FolioModules[0].Name)
	assert.Nil(t, modules.FolioModules[0].Version)
	assert.Empty(t, modules.FolioModules[0].SidecarName)

	// Other module should be processed
	assert.Equal(t, "mod-users", modules.FolioModules[1].Name)
}

func TestExtractModuleMetadata_EmptyModules(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	mockAWS := &MockAWSSvc{}
	action := testhelpers.NewMockAction()
	svc := registrysvc.New(action, mockHTTP, mockAWS)

	modules := models.NewProxyModulesByRegistry(nil, nil)

	// Act & Assert - should not panic
	svc.ExtractModuleMetadata(modules)
}
