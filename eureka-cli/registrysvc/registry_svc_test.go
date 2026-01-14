package registrysvc_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/j011195/eureka-setup/eureka-cli/constant"
	"github.com/j011195/eureka-setup/eureka-cli/field"
	"github.com/j011195/eureka-setup/eureka-cli/internal/testhelpers"
	"github.com/j011195/eureka-setup/eureka-cli/models"
	"github.com/j011195/eureka-setup/eureka-cli/registrysvc"
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
	result, err := svc.GetModules(installJsonURLs, true, false)

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
	result, err := svc.GetModules(installJsonURLs, true, false)

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
	result, err := svc.GetModules(installJsonURLs, true, false)

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
	result, err := svc.GetModules(installJsonURLs, true, false)

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
	result, err := svc.GetModules(installJsonURLs, true, true)

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

func TestGetModules_ModuleSorting(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	mockAWS := &MockAWSSvc{}
	action := testhelpers.NewMockAction()
	svc := registrysvc.New(action, mockHTTP, mockAWS)

	installJsonURLs := map[string]string{
		constant.FolioRegistry: "http://folio.example.com/install.json",
	}

	// Unsorted modules
	folioModules := []*models.ProxyModule{
		{ID: "mod-users-1.0.0", Action: "enable"},
		{ID: "mod-inventory-1.0.0", Action: "enable"},
		{ID: "mod-agreements-2.0.0", Action: "enable"},
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
	result, err := svc.GetModules(installJsonURLs, true, false)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.FolioModules, 3)
	// Verify sorting - should be alphabetically sorted
	assert.Equal(t, "mod-agreements-2.0.0", result.FolioModules[0].ID)
	assert.Equal(t, "mod-inventory-1.0.0", result.FolioModules[1].ID)
	assert.Equal(t, "mod-users-1.0.0", result.FolioModules[2].ID)
	mockHTTP.AssertExpectations(t)
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

func TestGetModules_BothRegistries(t *testing.T) {
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
		{ID: "mod-z-1.0.0", Action: "enable"},
		{ID: "mod-a-1.0.0", Action: "enable"},
	}

	eurekaModules := []*models.ProxyModule{
		{ID: "mod-eureka-2.0.0", Action: "enable"},
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
	result, err := svc.GetModules(installJsonURLs, true, false)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.FolioModules, 2)
	assert.Len(t, result.EurekaModules, 1)
	// Verify sorting: mod-a should come before mod-z
	assert.Equal(t, "mod-a-1.0.0", result.FolioModules[0].ID)
	assert.Equal(t, "mod-z-1.0.0", result.FolioModules[1].ID)
	mockHTTP.AssertExpectations(t)
}

// Tests for useRemote parameter

func TestGetModules_UseRemoteFalse_Success(t *testing.T) {
	t.Run("TestGetModules_UseRemoteFalse_Success", func(t *testing.T) {
		// Arrange
		mockHTTP := &testhelpers.MockHTTPClient{}
		mockAWS := &MockAWSSvc{}
		action := testhelpers.NewMockAction()
		svc := registrysvc.New(action, mockHTTP, mockAWS)

		// Create temp directory structure matching .eureka
		tmpDir := t.TempDir()
		eurekaDir := filepath.Join(tmpDir, ".eureka")
		err := os.MkdirAll(eurekaDir, 0755)
		assert.NoError(t, err)

		folioModules := []*models.ProxyModule{
			{ID: "mod-inventory-1.0.0", Action: "enable"},
			{ID: "mod-users-1.0.0", Action: "enable"},
		}

		eurekaModules := []*models.ProxyModule{
			{ID: "mod-custom-1.0.0", Action: "enable"},
		}

		// Create local install files in .eureka directory (not misc subdirectory)
		testhelpers.CreateJSONFileInDir(t, eurekaDir, "install_folio.json", folioModules)
		testhelpers.CreateJSONFileInDir(t, eurekaDir, "install_eureka.json", eurekaModules)

		// Override GetHomeMiscDir by setting environment variable
		t.Setenv("HOME", tmpDir)
		t.Setenv("USERPROFILE", tmpDir)

		installJsonURLs := map[string]string{
			constant.FolioRegistry:  "http://folio.example.com/install.json",
			constant.EurekaRegistry: "http://eureka.example.com/install.json",
		}

		// Act - useRemote=false should read local files
		result, err := svc.GetModules(installJsonURLs, false, false)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.FolioModules, 2)
		assert.Len(t, result.EurekaModules, 1)
		// HTTP client should not be called when useRemote=false
		mockHTTP.AssertNotCalled(t, "GetRetryReturnStruct")
	})
}

func TestGetModules_UseRemoteFalse_LocalFileNotFound(t *testing.T) {
	t.Run("TestGetModules_UseRemoteFalse_LocalFileNotFound", func(t *testing.T) {
		// Arrange
		mockHTTP := &testhelpers.MockHTTPClient{}
		mockAWS := &MockAWSSvc{}
		action := testhelpers.NewMockAction()
		svc := registrysvc.New(action, mockHTTP, mockAWS)

		// Set to empty temp directory (no install files)
		tmpDir := t.TempDir()
		t.Setenv("HOME", tmpDir)
		t.Setenv("USERPROFILE", tmpDir)

		installJsonURLs := map[string]string{
			constant.FolioRegistry: "http://folio.example.com/install.json",
		}

		// Act - useRemote=false but local file doesn't exist
		result, err := svc.GetModules(installJsonURLs, false, false)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		// HTTP client should not be called
		mockHTTP.AssertNotCalled(t, "GetRetryReturnStruct")
	})
}

func TestGetModules_SkipRegistry_ReadsLocalFile(t *testing.T) {
	t.Run("TestGetModules_SkipRegistry_ReadsLocalFile", func(t *testing.T) {
		// Arrange
		mockHTTP := &testhelpers.MockHTTPClient{}
		mockAWS := &MockAWSSvc{}
		action := testhelpers.NewMockAction()
		action.Param.SkipRegistry = true // SkipRegistry forces local file read
		svc := registrysvc.New(action, mockHTTP, mockAWS)

		// Create temp directory structure
		tmpDir := t.TempDir()
		eurekaDir := filepath.Join(tmpDir, ".eureka")
		err := os.MkdirAll(eurekaDir, 0755)
		assert.NoError(t, err)

		folioModules := []*models.ProxyModule{
			{ID: "mod-local-1.0.0", Action: "enable"},
		}

		testhelpers.CreateJSONFileInDir(t, eurekaDir, "install_folio.json", folioModules)

		// Override home directory
		t.Setenv("HOME", tmpDir)
		t.Setenv("USERPROFILE", tmpDir)

		installJsonURLs := map[string]string{
			constant.FolioRegistry: "http://folio.example.com/install.json",
		}

		// Act - useRemote=true but SkipRegistry=true should read local file
		result, err := svc.GetModules(installJsonURLs, true, false)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.FolioModules, 1)
		assert.Equal(t, "mod-local-1.0.0", result.FolioModules[0].ID)
		// HTTP client should not be called when SkipRegistry=true
		mockHTTP.AssertNotCalled(t, "GetRetryReturnStruct")
	})
}

func TestGetModules_UseRemoteFalse_WithCustomFrontendModules(t *testing.T) {
	t.Run("TestGetModules_UseRemoteFalse_WithCustomFrontendModules", func(t *testing.T) {
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
		}

		// Create temp directory structure
		tmpDir := t.TempDir()
		eurekaDir := filepath.Join(tmpDir, ".eureka")
		err := os.MkdirAll(eurekaDir, 0755)
		assert.NoError(t, err)

		folioModules := []*models.ProxyModule{
			{ID: "mod-inventory-1.0.0", Action: "enable"},
		}

		testhelpers.CreateJSONFileInDir(t, eurekaDir, "install_folio.json", folioModules)

		t.Setenv("HOME", tmpDir)
		t.Setenv("USERPROFILE", tmpDir)

		installJsonURLs := map[string]string{
			constant.FolioRegistry: "http://folio.example.com/install.json",
		}

		// Act - useRemote=false with custom modules
		result, err := svc.GetModules(installJsonURLs, false, false)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		// Should have 1 local module + 1 custom module
		assert.Len(t, result.FolioModules, 2)

		// Verify custom module was added
		hasCustomUI := false
		for _, mod := range result.FolioModules {
			if mod.ID == "custom-ui-2.0.0" {
				hasCustomUI = true
			}
		}
		assert.True(t, hasCustomUI, "custom-ui module should be present")
		mockHTTP.AssertNotCalled(t, "GetRetryReturnStruct")
	})
}

func TestGetModules_UseRemoteFalse_EmptyLocalFile(t *testing.T) {
	t.Run("TestGetModules_UseRemoteFalse_EmptyLocalFile", func(t *testing.T) {
		// Arrange
		mockHTTP := &testhelpers.MockHTTPClient{}
		mockAWS := &MockAWSSvc{}
		action := testhelpers.NewMockAction()
		svc := registrysvc.New(action, mockHTTP, mockAWS)

		// Create temp directory structure with empty install file
		tmpDir := t.TempDir()
		eurekaDir := filepath.Join(tmpDir, ".eureka")
		err := os.MkdirAll(eurekaDir, 0755)
		assert.NoError(t, err)

		emptyModules := []*models.ProxyModule{}
		testhelpers.CreateJSONFileInDir(t, eurekaDir, "install_folio.json", emptyModules)

		t.Setenv("HOME", tmpDir)
		t.Setenv("USERPROFILE", tmpDir)

		installJsonURLs := map[string]string{
			constant.FolioRegistry: "http://folio.example.com/install.json",
		}

		// Act - useRemote=false with empty local file
		result, err := svc.GetModules(installJsonURLs, false, false)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		// Error message contains reference to local install file
		assert.Contains(t, err.Error(), "local install file")
		mockHTTP.AssertNotCalled(t, "GetRetryReturnStruct")
	})
}

func TestGetModules_UseRemoteFalse_MultipleRegistries(t *testing.T) {
	t.Run("TestGetModules_UseRemoteFalse_MultipleRegistries", func(t *testing.T) {
		// Arrange
		mockHTTP := &testhelpers.MockHTTPClient{}
		mockAWS := &MockAWSSvc{}
		action := testhelpers.NewMockAction()
		svc := registrysvc.New(action, mockHTTP, mockAWS)

		// Create temp directory structure with install files for both registries
		tmpDir := t.TempDir()
		eurekaDir := filepath.Join(tmpDir, ".eureka")
		err := os.MkdirAll(eurekaDir, 0755)
		assert.NoError(t, err)

		folioModules := []*models.ProxyModule{
			{ID: "mod-users-1.0.0", Action: "enable"},
			{ID: "mod-inventory-1.0.0", Action: "enable"},
		}

		eurekaModules := []*models.ProxyModule{
			{ID: "mod-eureka-custom-1.0.0", Action: "enable"},
		}

		testhelpers.CreateJSONFileInDir(t, eurekaDir, "install_folio.json", folioModules)
		testhelpers.CreateJSONFileInDir(t, eurekaDir, "install_eureka.json", eurekaModules)

		t.Setenv("HOME", tmpDir)
		t.Setenv("USERPROFILE", tmpDir)

		installJsonURLs := map[string]string{
			constant.FolioRegistry:  "http://folio.example.com/install.json",
			constant.EurekaRegistry: "http://eureka.example.com/install.json",
		}

		// Act - useRemote=false reads both local files
		result, err := svc.GetModules(installJsonURLs, false, false)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.FolioModules, 2)
		assert.Len(t, result.EurekaModules, 1)
		// Verify sorting
		assert.Equal(t, "mod-inventory-1.0.0", result.FolioModules[0].ID)
		assert.Equal(t, "mod-users-1.0.0", result.FolioModules[1].ID)
		mockHTTP.AssertNotCalled(t, "GetRetryReturnStruct")
	})
}

func TestGetModules_UseRemoteTrue_CreatesLocalFile(t *testing.T) {
	t.Run("TestGetModules_UseRemoteTrue_CreatesLocalFile", func(t *testing.T) {
		// Arrange
		mockHTTP := &testhelpers.MockHTTPClient{}
		mockAWS := &MockAWSSvc{}
		action := testhelpers.NewMockAction()
		svc := registrysvc.New(action, mockHTTP, mockAWS)

		tmpDir := t.TempDir()
		// Create the .eureka directory structure that the code expects
		eurekaDir := filepath.Join(tmpDir, ".eureka")
		err := os.MkdirAll(eurekaDir, 0755)
		assert.NoError(t, err)

		t.Setenv("HOME", tmpDir)
		t.Setenv("USERPROFILE", tmpDir)

		installJsonURLs := map[string]string{
			constant.FolioRegistry: "http://folio.example.com/install.json",
		}

		folioModules := []*models.ProxyModule{
			{ID: "mod-remote-1.0.0", Action: "enable"},
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

		// Act - useRemote=true fetches from remote and creates local file
		result, err := svc.GetModules(installJsonURLs, true, false)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.FolioModules, 1)
		assert.Equal(t, "mod-remote-1.0.0", result.FolioModules[0].ID)
		mockHTTP.AssertExpectations(t)

		// Verify local file was created in the .eureka directory
		expectedPath := filepath.Join(tmpDir, ".eureka", "install_folio.json")
		fileInfo, err := os.Stat(expectedPath)
		assert.NoError(t, err, "local file should be created")
		if err == nil {
			assert.False(t, fileInfo.IsDir(), "should be a file not a directory")
		}
	})
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

func TestGetModules_UseRemoteFalse_InvalidJSON(t *testing.T) {
	t.Run("TestGetModules_UseRemoteFalse_InvalidJSON", func(t *testing.T) {
		// Arrange
		mockHTTP := &testhelpers.MockHTTPClient{}
		mockAWS := &MockAWSSvc{}
		action := testhelpers.NewMockAction()
		svc := registrysvc.New(action, mockHTTP, mockAWS)

		// Create temp directory with invalid JSON file
		tmpDir := t.TempDir()
		eurekaDir := filepath.Join(tmpDir, ".eureka")
		err := os.MkdirAll(eurekaDir, 0755)
		assert.NoError(t, err)

		// Write invalid JSON to file
		invalidJSON := []byte(`{"invalid": json}`)
		err = os.WriteFile(filepath.Join(eurekaDir, "install_folio.json"), invalidJSON, 0600)
		assert.NoError(t, err)

		t.Setenv("HOME", tmpDir)
		t.Setenv("USERPROFILE", tmpDir)

		installJsonURLs := map[string]string{
			constant.FolioRegistry: "http://folio.example.com/install.json",
		}

		// Act - useRemote=false with invalid JSON file
		result, err := svc.GetModules(installJsonURLs, false, false)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		mockHTTP.AssertNotCalled(t, "GetRetryReturnStruct")
	})
}

func TestGetModules_UseRemoteTrue_WriteFileError(t *testing.T) {
	t.Run("TestGetModules_UseRemoteTrue_WriteFileError", func(t *testing.T) {
		// Arrange
		mockHTTP := &testhelpers.MockHTTPClient{}
		mockAWS := &MockAWSSvc{}
		action := testhelpers.NewMockAction()
		svc := registrysvc.New(action, mockHTTP, mockAWS)

		// Set to a directory path where the install file path is actually a directory
		tmpDir := t.TempDir()
		eurekaDir := filepath.Join(tmpDir, ".eureka")
		err := os.MkdirAll(eurekaDir, 0755)
		assert.NoError(t, err)

		// Create a directory with the name of the file we want to write
		// This will cause WriteJSONToFile to fail
		installFileDir := filepath.Join(eurekaDir, "install_folio.json")
		err = os.MkdirAll(installFileDir, 0755)
		assert.NoError(t, err)

		t.Setenv("HOME", tmpDir)
		t.Setenv("USERPROFILE", tmpDir)

		installJsonURLs := map[string]string{
			constant.FolioRegistry: "http://folio.example.com/install.json",
		}

		folioModules := []*models.ProxyModule{
			{ID: "mod-test-1.0.0", Action: "enable"},
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

		// Act - useRemote=true but can't write file (path is a directory)
		result, err := svc.GetModules(installJsonURLs, true, false)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		mockHTTP.AssertExpectations(t)
	})
}
