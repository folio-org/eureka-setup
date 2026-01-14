package modulesvc

import (
	"errors"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/j011195/eureka-setup/eureka-cli/constant"
	"github.com/j011195/eureka-setup/eureka-cli/field"
	"github.com/j011195/eureka-setup/eureka-cli/internal/testhelpers"
	"github.com/j011195/eureka-setup/eureka-cli/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNew(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	mockHTTP := new(testhelpers.MockHTTPClient)
	mockRegistry := new(testhelpers.MockRegistrySvc)
	mockModuleEnv := new(testhelpers.MockModuleEnv)

	// Act
	svc := New(action, mockHTTP, nil, mockRegistry, mockModuleEnv)

	// Assert
	assert.NotNil(t, svc)
	assert.Equal(t, action, svc.Action)
	assert.Equal(t, mockHTTP, svc.HTTPClient)
	assert.Equal(t, mockRegistry, svc.RegistrySvc)
	assert.Equal(t, mockModuleEnv, svc.ModuleEnv)
}

func TestGetBackendModule_Found(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	svc := New(action, nil, nil, nil, nil)

	backendModule := models.BackendModule{DeployModule: true}
	version := "1.0.0"
	module := &models.ProxyModule{
		ID: "mod-test-1.0.0",
		Metadata: models.ProxyModuleMetadata{
			Name:    "mod-test",
			Version: &version,
		},
	}

	containers := &models.Containers{
		Modules: &models.ProxyModulesByRegistry{
			FolioModules: []*models.ProxyModule{module},
		},
		BackendModules: map[string]models.BackendModule{
			"mod-test": backendModule,
		},
	}

	// Act
	foundBackend, foundModule := svc.GetBackendModule(containers, "mod-test")

	// Assert
	assert.NotNil(t, foundBackend)
	assert.NotNil(t, foundModule)
	assert.Equal(t, backendModule.DeployModule, foundBackend.DeployModule)
	assert.Equal(t, module.Metadata.Name, foundModule.Metadata.Name)
}

func TestGetBackendModule_NotFound(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	svc := New(action, nil, nil, nil, nil)

	containers := &models.Containers{
		Modules: &models.ProxyModulesByRegistry{
			FolioModules: []*models.ProxyModule{},
		},
		BackendModules: map[string]models.BackendModule{},
	}

	// Act
	foundBackend, foundModule := svc.GetBackendModule(containers, "mod-nonexistent")

	// Assert
	assert.Nil(t, foundBackend)
	assert.Nil(t, foundModule)
}

func TestGetBackendModule_DeployModuleFalse(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	svc := New(action, nil, nil, nil, nil)

	backendModule := models.BackendModule{DeployModule: false}
	version := "1.0.0"
	module := &models.ProxyModule{
		ID: "mod-test-1.0.0",
		Metadata: models.ProxyModuleMetadata{
			Name:    "mod-test",
			Version: &version,
		},
	}

	containers := &models.Containers{
		Modules: &models.ProxyModulesByRegistry{
			FolioModules: []*models.ProxyModule{module},
		},
		BackendModules: map[string]models.BackendModule{
			"mod-test": backendModule,
		},
	}

	// Act
	foundBackend, foundModule := svc.GetBackendModule(containers, "mod-test")

	// Assert - should not be found because DeployModule is false
	assert.Nil(t, foundBackend)
	assert.Nil(t, foundModule)
}

func TestGetBackendModule_ChecksEurekaModules(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	svc := New(action, nil, nil, nil, nil)

	backendModule := models.BackendModule{DeployModule: true}
	version := "2.0.0"
	eurekaModule := &models.ProxyModule{
		ID: "mod-eureka-2.0.0",
		Metadata: models.ProxyModuleMetadata{
			Name:    "mod-eureka",
			Version: &version,
		},
	}

	containers := &models.Containers{
		Modules: &models.ProxyModulesByRegistry{
			FolioModules:  []*models.ProxyModule{},
			EurekaModules: []*models.ProxyModule{eurekaModule},
		},
		BackendModules: map[string]models.BackendModule{
			"mod-eureka": backendModule,
		},
	}

	// Act
	foundBackend, foundModule := svc.GetBackendModule(containers, "mod-eureka")

	// Assert
	assert.NotNil(t, foundBackend)
	assert.NotNil(t, foundModule)
	assert.Equal(t, "mod-eureka", foundModule.Metadata.Name)
}

func TestGetModuleImageVersion_UseBackendModuleVersion(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	svc := New(action, nil, nil, nil, nil)

	moduleVersion := "1.0.0"
	backendModuleVersion := "2.0.0-custom"
	backendModule := models.BackendModule{ModuleVersion: &backendModuleVersion}
	module := &models.ProxyModule{
		Metadata: models.ProxyModuleMetadata{
			Version: &moduleVersion,
		},
	}

	// Act
	version := svc.GetModuleImageVersion(backendModule, module)

	// Assert - should use backend module version
	assert.Equal(t, "2.0.0-custom", version)
}

func TestGetModuleImageVersion_UseProxyModuleVersion(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	svc := New(action, nil, nil, nil, nil)

	moduleVersion := "1.0.0"
	backendModule := models.BackendModule{ModuleVersion: nil}
	module := &models.ProxyModule{
		Metadata: models.ProxyModuleMetadata{
			Version: &moduleVersion,
		},
	}

	// Act
	version := svc.GetModuleImageVersion(backendModule, module)

	// Assert - should use proxy module version
	assert.Equal(t, "1.0.0", version)
}

func TestGetSidecarImage_CustomNamespace(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	action.ConfigSidecarModule = map[string]any{
		field.SidecarModuleVersionEntry:         "3.0.0",
		field.SidecarModuleCustomNamespaceEntry: true,
		field.SidecarModuleImageEntry:           "my-custom-sidecar",
	}
	svc := New(action, nil, nil, nil, nil)

	sidecarVersion := "3.0.0"
	modules := []*models.ProxyModule{
		{
			Metadata: models.ProxyModuleMetadata{
				Name:    constant.SidecarProjectName,
				Version: &sidecarVersion,
			},
		},
	}

	// Act
	image, shouldPull, err := svc.GetSidecarImage(modules)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "my-custom-sidecar:3.0.0", image)
	assert.True(t, shouldPull)
}

func TestGetSidecarImage_RegistryImage(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	action.ConfigSidecarModule = map[string]any{
		field.SidecarModuleVersionEntry:         "3.0.0",
		field.SidecarModuleCustomNamespaceEntry: false,
		field.SidecarModuleImageEntry:           "mgr-tenant-entitlement",
	}

	mockRegistry := new(testhelpers.MockRegistrySvc)
	mockRegistry.On("GetNamespace", "3.0.0").Return("docker.io/folioorg")

	svc := New(action, nil, nil, mockRegistry, nil)

	sidecarVersion := "3.0.0"
	modules := []*models.ProxyModule{
		{
			Metadata: models.ProxyModuleMetadata{
				Name:    constant.SidecarProjectName,
				Version: &sidecarVersion,
			},
		},
	}

	// Act
	image, shouldPull, err := svc.GetSidecarImage(modules)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "docker.io/folioorg/mgr-tenant-entitlement:3.0.0", image)
	assert.True(t, shouldPull)
	mockRegistry.AssertExpectations(t)
}

func TestGetSidecarImage_NoSidecarVersionFound(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	action.ConfigSidecarModule = map[string]any{
		field.SidecarModuleVersionEntry: nil,
		field.SidecarModuleImageEntry:   "mgr-tenant-entitlement",
	}

	svc := New(action, nil, nil, nil, nil)

	otherVersion := "1.0.0"
	modules := []*models.ProxyModule{
		{
			Metadata: models.ProxyModuleMetadata{
				Name:    "some-other-module",
				Version: &otherVersion,
			},
		},
	}

	// Act
	image, shouldPull, err := svc.GetSidecarImage(modules)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "resource not found")
	assert.Empty(t, image)
	assert.False(t, shouldPull)
}

func TestGetSidecarImage_BlankImageName(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	action.ConfigSidecarModule = map[string]any{
		field.SidecarModuleVersionEntry: "3.0.0",
		field.SidecarModuleImageEntry:   "",
	}

	svc := New(action, nil, nil, nil, nil)

	sidecarVersion := "3.0.0"
	modules := []*models.ProxyModule{
		{
			Metadata: models.ProxyModuleMetadata{
				Name:    constant.SidecarProjectName,
				Version: &sidecarVersion,
			},
		},
	}

	// Act
	image, shouldPull, err := svc.GetSidecarImage(modules)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sidecar image is blank")
	assert.Empty(t, image)
	assert.False(t, shouldPull)
}

func TestGetSidecarImage_EmptyVersionString(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	action.ConfigSidecarModule = map[string]any{
		field.SidecarModuleVersionEntry: nil,
		field.SidecarModuleImageEntry:   "mgr-tenant-entitlement",
	}

	svc := New(action, nil, nil, nil, nil)

	emptyVersion := ""
	modules := []*models.ProxyModule{
		{
			Metadata: models.ProxyModuleMetadata{
				Name:    constant.SidecarProjectName,
				Version: &emptyVersion, // Empty string version
			},
		},
	}

	// Act
	image, shouldPull, err := svc.GetSidecarImage(modules)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "resource not found")
	assert.Empty(t, image)
	assert.False(t, shouldPull)
}

func TestGetSidecarImage_InvalidVersionType(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	action.ConfigSidecarModule = map[string]any{
		field.SidecarModuleVersionEntry: 123, // Invalid type (int instead of string)
		field.SidecarModuleImageEntry:   "mgr-tenant-entitlement",
	}

	mockRegistry := new(testhelpers.MockRegistrySvc)
	mockRegistry.On("GetNamespace", "3.0.0").Return("docker.io/folioorg")

	svc := New(action, nil, nil, mockRegistry, nil)

	sidecarVersion := "3.0.0"
	modules := []*models.ProxyModule{
		{
			Metadata: models.ProxyModuleMetadata{
				Name:    constant.SidecarProjectName,
				Version: &sidecarVersion,
			},
		},
	}

	// Act
	image, shouldPull, err := svc.GetSidecarImage(modules)

	// Assert - Falls back to registry version when config version is invalid type
	assert.NoError(t, err)
	assert.Equal(t, "docker.io/folioorg/mgr-tenant-entitlement:3.0.0", image)
	assert.True(t, shouldPull)
	mockRegistry.AssertExpectations(t)
}

func TestGetModuleImage(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	mockRegistry := new(testhelpers.MockRegistrySvc)
	mockRegistry.On("GetNamespace", "1.5.0").Return("ghcr.io/folio-org")

	svc := New(action, nil, nil, mockRegistry, nil)

	module := &models.ProxyModule{
		Metadata: models.ProxyModuleMetadata{
			Name: "mod-users",
		},
	}

	// Act
	image := svc.GetModuleImage(module, "1.5.0")

	// Assert
	assert.Equal(t, "ghcr.io/folio-org/mod-users:1.5.0", image)
	mockRegistry.AssertExpectations(t)
}

func TestGetModuleEnv_AllFeatures(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	mockModuleEnv := new(testhelpers.MockModuleEnv)

	mockModuleEnv.On("VaultEnv", []string(nil), mock.Anything).Return([]string{"VAULT_ENABLED=true"})
	mockModuleEnv.On("OkapiEnv", mock.Anything, "mod-test-sidecar", 8081).Return([]string{"OKAPI_URL=http://mod-test-sidecar:8081"})
	mockModuleEnv.On("DisabledSystemUserEnv", mock.Anything, "mod-test").Return([]string{"SYSTEM_USER_ENABLED=false"})
	mockModuleEnv.On("ModuleEnv", mock.Anything, map[string]any{"CUSTOM_VAR": "value"}).Return([]string{"CUSTOM_VAR=value"})

	svc := New(action, nil, nil, nil, mockModuleEnv)

	container := &models.Containers{}

	module := &models.ProxyModule{
		Metadata: models.ProxyModuleMetadata{
			Name:        "mod-test",
			SidecarName: "mod-test-sidecar",
		},
	}

	backendModule := models.BackendModule{
		UseVault:          true,
		UseOkapiURL:       true,
		DisableSystemUser: true,
		PrivatePort:       8081,
		ModuleEnv:         map[string]any{"CUSTOM_VAR": "value"},
	}

	// Act
	env := svc.GetModuleEnv(container, module, backendModule)

	// Assert
	assert.NotEmpty(t, env)
	mockModuleEnv.AssertExpectations(t)
}

func TestGetModuleEnv_OnlyGlobalEnv(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	mockModuleEnv := new(testhelpers.MockModuleEnv)

	mockModuleEnv.On("ModuleEnv", []string(nil), map[string]any(nil)).Return([]string{"GLOBAL_VAR=global"})

	svc := New(action, nil, nil, nil, mockModuleEnv)

	container := &models.Containers{}

	module := &models.ProxyModule{
		Metadata: models.ProxyModuleMetadata{
			Name: "mod-simple",
		},
	}

	backendModule := models.BackendModule{
		UseVault:          false,
		UseOkapiURL:       false,
		DisableSystemUser: false,
		ModuleEnv:         nil,
	}

	// Act
	env := svc.GetModuleEnv(container, module, backendModule)

	// Assert
	assert.NotEmpty(t, env)
	mockModuleEnv.AssertExpectations(t)
}

func TestGetSidecarEnv(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	mockModuleEnv := new(testhelpers.MockModuleEnv)

	mockModuleEnv.On("VaultEnv", []string(nil), mock.Anything).Return([]string{"VAULT_ENV=set"})
	mockModuleEnv.On("KeycloakEnv", mock.Anything).Return([]string{"KEYCLOAK_ENV=set"})
	mockModuleEnv.On("SidecarEnv", mock.Anything, mock.Anything, 8081, "http://mod-test:8081", "http://sidecar:8081").
		Return([]string{"SIDECAR_ENV=set"})

	svc := New(action, nil, nil, nil, mockModuleEnv)

	containers := &models.Containers{}

	module := &models.ProxyModule{
		Metadata: models.ProxyModuleMetadata{
			Name: "mod-test",
		},
	}

	backendModule := models.BackendModule{
		PrivatePort: 8081,
	}

	// Act
	env := svc.GetSidecarEnv(containers, module, backendModule, "http://mod-test:8081", "http://sidecar:8081")

	// Assert
	assert.NotEmpty(t, env)
	mockModuleEnv.AssertExpectations(t)
}

// ModuleReadinessChecker Tests

func TestCheckModuleReadiness_Success(t *testing.T) {
	// Arrange
	mockHTTP := new(testhelpers.MockHTTPClient)
	action := testhelpers.NewMockAction()
	svc := New(action, mockHTTP, nil, nil, nil)
	svc.ReadinessMaxRetries = 3
	svc.ReadinessWait = 1 * time.Millisecond

	mockHTTP.On("Ping",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/admin/health")
		}),
		mock.Anything).
		Return(http.StatusOK, nil)

	wg := &sync.WaitGroup{}
	errCh := make(chan error, 1)
	wg.Add(1)

	// Act
	go svc.CheckModuleReadiness(wg, errCh, "test-module", 8080)
	wg.Wait()
	close(errCh)

	// Assert
	select {
	case err := <-errCh:
		assert.NoError(t, err)
	default:
		// Success - no error sent
	}
	mockHTTP.AssertExpectations(t)
}

func TestCheckModuleReadiness_HTTPError(t *testing.T) {
	// Arrange
	mockHTTP := new(testhelpers.MockHTTPClient)
	action := testhelpers.NewMockAction()
	svc := New(action, mockHTTP, nil, nil, nil)
	svc.ReadinessMaxRetries = 3
	svc.ReadinessWait = 1 * time.Millisecond

	mockHTTP.On("Ping",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/admin/health")
		}),
		mock.Anything).
		Return(0, errors.New("connection error"))

	wg := &sync.WaitGroup{}
	errCh := make(chan error, 1)
	wg.Add(1)

	// Act
	go svc.CheckModuleReadiness(wg, errCh, "test-module", 8080)
	wg.Wait()
	close(errCh)

	// Assert
	err := <-errCh
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "module test-module")
	mockHTTP.AssertExpectations(t)
}

func TestCheckModuleReadiness_NilResponse(t *testing.T) {
	// Arrange
	mockHTTP := new(testhelpers.MockHTTPClient)
	action := testhelpers.NewMockAction()
	svc := New(action, mockHTTP, nil, nil, nil)
	svc.ReadinessMaxRetries = 3
	svc.ReadinessWait = 1 * time.Millisecond

	mockHTTP.On("Ping",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/admin/health")
		}),
		mock.Anything).
		Return(0, errors.New("nil response"))

	wg := &sync.WaitGroup{}
	errCh := make(chan error, 1)
	wg.Add(1)

	// Act
	go svc.CheckModuleReadiness(wg, errCh, "test-module", 8080)
	wg.Wait()
	close(errCh)

	// Assert
	err := <-errCh
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "module test-module")
	mockHTTP.AssertExpectations(t)
}

func TestCheckModuleReadiness_NonOKStatusCode(t *testing.T) {
	// Arrange
	mockHTTP := new(testhelpers.MockHTTPClient)
	action := testhelpers.NewMockAction()
	svc := New(action, mockHTTP, nil, nil, nil)
	svc.ReadinessMaxRetries = 3
	svc.ReadinessWait = 1 * time.Millisecond

	mockHTTP.On("Ping",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/admin/health")
		}),
		mock.Anything).
		Return(http.StatusServiceUnavailable, nil)

	wg := &sync.WaitGroup{}
	errCh := make(chan error, 1)
	wg.Add(1)

	// Act
	go svc.CheckModuleReadiness(wg, errCh, "test-module", 8080)
	wg.Wait()
	close(errCh)

	// Assert
	err := <-errCh
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "module test-module")
	mockHTTP.AssertExpectations(t)
}

func TestCheckModuleReadiness_EventualSuccess(t *testing.T) {
	// Arrange
	mockHTTP := new(testhelpers.MockHTTPClient)
	action := testhelpers.NewMockAction()
	svc := New(action, mockHTTP, nil, nil, nil)
	svc.ReadinessMaxRetries = 3
	svc.ReadinessWait = 1 * time.Millisecond

	// First 2 calls fail, third succeeds
	mockHTTP.On("Ping",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/admin/health")
		}),
		mock.Anything).
		Return(http.StatusServiceUnavailable, nil).Times(2)

	mockHTTP.On("Ping",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/admin/health")
		}),
		mock.Anything).
		Return(http.StatusOK, nil).Once()

	wg := &sync.WaitGroup{}
	errCh := make(chan error, 1)
	wg.Add(1)

	// Act
	go svc.CheckModuleReadiness(wg, errCh, "test-module", 8080)
	wg.Wait()
	close(errCh)

	// Assert
	select {
	case err := <-errCh:
		assert.NoError(t, err)
	default:
		// Success - no error sent
	}
	mockHTTP.AssertExpectations(t)
}

func TestCheckModuleReadiness_MultipleModulesConcurrent(t *testing.T) {
	// Arrange
	mockHTTP := new(testhelpers.MockHTTPClient)
	action := testhelpers.NewMockAction()
	svc := New(action, mockHTTP, nil, nil, nil)
	svc.ReadinessMaxRetries = 3
	svc.ReadinessWait = 1 * time.Millisecond

	mockHTTP.On("Ping",
		mock.Anything,
		mock.Anything).
		Return(http.StatusOK, nil)

	wg := &sync.WaitGroup{}
	errCh := make(chan error, 3)

	modules := []struct {
		name string
		port int
	}{
		{"module-1", 8081},
		{"module-2", 8082},
		{"module-3", 8083},
	}

	// Act
	for _, mod := range modules {
		wg.Add(1)
		go svc.CheckModuleReadiness(wg, errCh, mod.name, mod.port)
	}
	wg.Wait()
	close(errCh)

	// Assert
	errorCount := 0
	for err := range errCh {
		if err != nil {
			errorCount++
		}
	}
	assert.Equal(t, 0, errorCount, "Expected no errors from concurrent module checks")
	mockHTTP.AssertExpectations(t)
}

func TestCheckModuleReadiness_ErrorChannelFull(t *testing.T) {
	// Arrange
	mockHTTP := new(testhelpers.MockHTTPClient)
	action := testhelpers.NewMockAction()
	svc := New(action, mockHTTP, nil, nil, nil)
	svc.ReadinessMaxRetries = 3
	svc.ReadinessWait = 1 * time.Millisecond

	mockHTTP.On("Ping",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/admin/health")
		}),
		mock.Anything).
		Return(0, errors.New("test error"))

	wg := &sync.WaitGroup{}
	errCh := make(chan error) // Unbuffered channel to test default case
	wg.Add(1)

	// Act
	go svc.CheckModuleReadiness(wg, errCh, "test-module", 8080)
	wg.Wait()

	// Assert
	// The goroutine should complete without blocking even if error channel is not read
	// This tests the default case in the select statement
	mockHTTP.AssertExpectations(t)
}

func TestCheckModuleReadiness_VerifyRetryLogic(t *testing.T) {
	// Arrange
	mockHTTP := new(testhelpers.MockHTTPClient)
	action := testhelpers.NewMockAction()
	svc := New(action, mockHTTP, nil, nil, nil)
	svc.ReadinessMaxRetries = 5
	svc.ReadinessWait = 1 * time.Millisecond

	// Will retry until max retries
	mockHTTP.On("Ping",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/admin/health")
		}),
		mock.Anything).
		Return(http.StatusServiceUnavailable, nil)

	wg := &sync.WaitGroup{}
	errCh := make(chan error, 1)
	wg.Add(1)

	// Act
	go svc.CheckModuleReadiness(wg, errCh, "test-module", 8080)
	wg.Wait()
	close(errCh)

	// Assert
	err := <-errCh
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "module test-module")
	// Should have been called exactly maxRetries times
	mockHTTP.AssertNumberOfCalls(t, "Ping", 5)
}

func TestCheckModuleReadiness_PortInURL(t *testing.T) {
	// Arrange
	mockHTTP := new(testhelpers.MockHTTPClient)
	action := testhelpers.NewMockAction()
	svc := New(action, mockHTTP, nil, nil, nil)
	svc.ReadinessMaxRetries = 3
	svc.ReadinessWait = 1 * time.Millisecond

	var capturedURL string
	mockHTTP.On("Ping",
		mock.MatchedBy(func(urlStr string) bool {
			capturedURL = urlStr
			return true
		}),
		mock.Anything).
		Return(http.StatusOK, nil)

	wg := &sync.WaitGroup{}
	errCh := make(chan error, 1)
	wg.Add(1)

	// Act
	go svc.CheckModuleReadiness(wg, errCh, "test-module", 9999)
	wg.Wait()
	close(errCh)

	// Assert
	assert.Contains(t, capturedURL, ":9999")
	assert.Contains(t, capturedURL, "/admin/health")
	mockHTTP.AssertExpectations(t)
}

func TestCheckModuleReadiness_DefaultMaxRetries(t *testing.T) {
	// Arrange
	mockHTTP := new(testhelpers.MockHTTPClient)
	action := testhelpers.NewMockAction()
	svc := New(action, mockHTTP, nil, nil, nil)
	// Don't set ReadinessMaxRetries - should default to constant value
	svc.ReadinessWait = 1 * time.Millisecond

	mockHTTP.On("Ping",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/admin/health")
		}),
		mock.Anything).
		Return(http.StatusOK, nil)

	wg := &sync.WaitGroup{}
	errCh := make(chan error, 1)
	wg.Add(1)

	// Act
	go svc.CheckModuleReadiness(wg, errCh, "test-module", 8080)
	wg.Wait()
	close(errCh)

	// Assert
	select {
	case err := <-errCh:
		assert.NoError(t, err)
	default:
		// Success - no error sent, defaults to constant.ModuleReadinessMaxRetries
	}
	mockHTTP.AssertExpectations(t)
}

// ==================== GetLocalModuleImage Tests ====================

func TestGetLocalModuleImage_Success(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	svc := New(action, nil, nil, nil, nil)

	namespace := "ghcr.io/folio-org"
	moduleName := "mod-inventory"
	moduleVersion := "1.2.3"

	// Act
	result := svc.GetLocalModuleImage(namespace, moduleName, moduleVersion)

	// Assert
	assert.Equal(t, "ghcr.io/folio-org/mod-inventory:1.2.3", result)
}

func TestGetLocalModuleImage_WithSnapshot(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	svc := New(action, nil, nil, nil, nil)

	namespace := "docker.dev.folio.org"
	moduleName := "mod-circulation"
	moduleVersion := "2.0.0-SNAPSHOT"

	// Act
	result := svc.GetLocalModuleImage(namespace, moduleName, moduleVersion)

	// Assert
	assert.Equal(t, "docker.dev.folio.org/mod-circulation:2.0.0-SNAPSHOT", result)
}

func TestGetLocalModuleImage_CustomNamespace(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	svc := New(action, nil, nil, nil, nil)

	namespace := "localhost:5000"
	moduleName := "mod-custom"
	moduleVersion := "1.0.0"

	// Act
	result := svc.GetLocalModuleImage(namespace, moduleName, moduleVersion)

	// Assert
	assert.Equal(t, "localhost:5000/mod-custom:1.0.0", result)
}

func TestGetLocalModuleImage_EmptyNamespace(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	svc := New(action, nil, nil, nil, nil)

	namespace := ""
	moduleName := "mod-test"
	moduleVersion := "1.0.0"

	// Act
	result := svc.GetLocalModuleImage(namespace, moduleName, moduleVersion)

	// Assert
	assert.Equal(t, "/mod-test:1.0.0", result)
}

func TestGetLocalModuleImage_EmptyModuleName(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	svc := New(action, nil, nil, nil, nil)

	namespace := "docker.io/library"
	moduleName := ""
	moduleVersion := "1.0.0"

	// Act
	result := svc.GetLocalModuleImage(namespace, moduleName, moduleVersion)

	// Assert
	assert.Equal(t, "docker.io/library/:1.0.0", result)
}

func TestGetLocalModuleImage_EmptyVersion(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	svc := New(action, nil, nil, nil, nil)

	namespace := "docker.io/library"
	moduleName := "mod-test"
	moduleVersion := ""

	// Act
	result := svc.GetLocalModuleImage(namespace, moduleName, moduleVersion)

	// Assert
	assert.Equal(t, "docker.io/library/mod-test:", result)
}

func TestGetLocalModuleImage_AllEmpty(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	svc := New(action, nil, nil, nil, nil)

	namespace := ""
	moduleName := ""
	moduleVersion := ""

	// Act
	result := svc.GetLocalModuleImage(namespace, moduleName, moduleVersion)

	// Assert
	assert.Equal(t, "/:", result)
}

func TestGetLocalModuleImage_WithHyphenatedModuleName(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	svc := New(action, nil, nil, nil, nil)

	namespace := "registry.example.com"
	moduleName := "mod-inventory-storage"
	moduleVersion := "3.4.5"

	// Act
	result := svc.GetLocalModuleImage(namespace, moduleName, moduleVersion)

	// Assert
	assert.Equal(t, "registry.example.com/mod-inventory-storage:3.4.5", result)
}

func TestGetLocalModuleImage_WithSpecialCharacters(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	svc := New(action, nil, nil, nil, nil)

	namespace := "registry.example.com/project/team"
	moduleName := "mod-test_special"
	moduleVersion := "1.0.0-beta.1"

	// Act
	result := svc.GetLocalModuleImage(namespace, moduleName, moduleVersion)

	// Assert
	assert.Equal(t, "registry.example.com/project/team/mod-test_special:1.0.0-beta.1", result)
}

func TestGetLocalModuleImage_ShortVersion(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	svc := New(action, nil, nil, nil, nil)

	namespace := "ghcr.io/folio"
	moduleName := "mod-simple"
	moduleVersion := "1"

	// Act
	result := svc.GetLocalModuleImage(namespace, moduleName, moduleVersion)

	// Assert
	assert.Equal(t, "ghcr.io/folio/mod-simple:1", result)
}

func TestGetLocalModuleImage_LatestTag(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	svc := New(action, nil, nil, nil, nil)

	namespace := "docker.io/folioorg"
	moduleName := "mod-latest"
	moduleVersion := "latest"

	// Act
	result := svc.GetLocalModuleImage(namespace, moduleName, moduleVersion)

	// Assert
	assert.Equal(t, "docker.io/folioorg/mod-latest:latest", result)
}
