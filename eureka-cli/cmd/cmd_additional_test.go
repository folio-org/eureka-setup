package cmd

import (
	"errors"
	"net"
	"os"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/folio-org/eureka-setup/eureka-cli/action"
	"github.com/folio-org/eureka-setup/eureka-cli/constant"
	"github.com/folio-org/eureka-setup/eureka-cli/models"
	"github.com/folio-org/eureka-setup/eureka-cli/modulesvc"
	"github.com/folio-org/eureka-setup/eureka-cli/runconfig"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUpgradeModuleSvc is a mock for upgrademodulesvc.UpgradeModuleProcessor
type MockUpgradeModuleSvc struct {
	mock.Mock
}

func (m *MockUpgradeModuleSvc) SetNewModuleVersionAndIDIntoContext() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockUpgradeModuleSvc) SetDefaultNamespaceIntoContext() {
	m.Called()
}

func (m *MockUpgradeModuleSvc) BuildModuleArtifact(moduleName, moduleVersion, modulePath string) error {
	args := m.Called(moduleName, moduleVersion, modulePath)
	return args.Error(0)
}

func (m *MockUpgradeModuleSvc) CleanModuleArtifact(moduleName, modulePath string) error {
	args := m.Called(moduleName, modulePath)
	return args.Error(0)
}

func (m *MockUpgradeModuleSvc) BuildModuleImage(namespace, moduleName, moduleVersion, modulePath string) error {
	args := m.Called(namespace, moduleName, moduleVersion, modulePath)
	return args.Error(0)
}

func (m *MockUpgradeModuleSvc) ReadModuleDescriptor(moduleName, moduleVersion, modulePath string) (map[string]any, error) {
	args := m.Called(moduleName, moduleVersion, modulePath)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]any), args.Error(1)
}

func (m *MockUpgradeModuleSvc) ResolveModuleIdentity(modulePath string) (string, string, error) {
	args := m.Called(modulePath)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockUpgradeModuleSvc) GetModuleDescriptorPath(modulePath string) (string, error) {
	args := m.Called(modulePath)
	return args.String(0), args.Error(1)
}

func (m *MockUpgradeModuleSvc) UpdateBackendModules(moduleName, newModuleVersion string, shouldBuild bool, oldBackendModules []any) ([]map[string]any, []map[string]string, string, error) {
	args := m.Called(moduleName, newModuleVersion, shouldBuild, oldBackendModules)
	if args.Get(0) == nil {
		return nil, nil, "", args.Error(3)
	}
	return args.Get(0).([]map[string]any), args.Get(1).([]map[string]string), args.String(2), args.Error(3)
}

func (m *MockUpgradeModuleSvc) UpdateFrontendModules(shouldBuild bool, oldFrontendModules []any) []map[string]any {
	args := m.Called(shouldBuild, oldFrontendModules)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]map[string]any)
}

func (m *MockUpgradeModuleSvc) UpdateBackendModuleDescriptors(moduleName, oldModuleID string, newModuleDescriptor map[string]any, oldBackendModuleDescriptors []any) []any {
	args := m.Called(moduleName, oldModuleID, newModuleDescriptor, oldBackendModuleDescriptors)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]any)
}

func (m *MockUpgradeModuleSvc) DeployModuleAndSidecarPair(client *client.Client, pair *modulesvc.ModulePair) error {
	args := m.Called(client, pair)
	return args.Error(0)
}

// MockKongSvc is a mock for kongsvc.KongProcessor
type MockKongSvc struct {
	mock.Mock
}

func (m *MockKongSvc) CheckRouteReadiness() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockKongSvc) CheckRouteExists(routeID string) (bool, *models.KongRoute, error) {
	args := m.Called(routeID)
	return args.Bool(0), args.Get(1).(*models.KongRoute), args.Error(2)
}

func (m *MockKongSvc) FindRouteByExpressions(expressions []string) ([]*models.KongRoute, error) {
	args := m.Called(expressions)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.KongRoute), args.Error(1)
}

func (m *MockKongSvc) ListAllRoutes() ([]models.KongRoute, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.KongRoute), args.Error(1)
}

// ==================== UpgradeModule Tests ====================

func TestValidateModulePath_EmptyPath(t *testing.T) {
	// Arrange
	run := &Run{
		Config: &runconfig.RunConfig{
			Infrastructure: &runconfig.Infrastructure{
				Action: &action.Action{},
			},
		},
	}

	// Act
	err := run.validateModulePath("")

	// Assert
	assert.NoError(t, err)
}

func TestValidateModulePath_ValidDirectory(t *testing.T) {
	// Arrange
	run := &Run{
		Config: &runconfig.RunConfig{
			Infrastructure: &runconfig.Infrastructure{
				Action: &action.Action{},
			},
		},
	}
	tempDir := t.TempDir()

	// Act
	err := run.validateModulePath(tempDir)

	// Assert
	assert.NoError(t, err)
}

func TestValidateModulePath_PathDoesNotExist(t *testing.T) {
	// Arrange
	run := &Run{
		Config: &runconfig.RunConfig{
			Infrastructure: &runconfig.Infrastructure{
				Action: &action.Action{},
			},
		},
	}
	nonExistentPath := "/path/that/does/not/exist/at/all"

	// Act
	err := run.validateModulePath(nonExistentPath)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "module path does not exist")
	assert.Contains(t, err.Error(), nonExistentPath)
}

func TestValidateModulePath_PathIsFile(t *testing.T) {
	// Arrange
	run := &Run{
		Config: &runconfig.RunConfig{
			Infrastructure: &runconfig.Infrastructure{
				Action: &action.Action{},
			},
		},
	}
	tempDir := t.TempDir()
	tempFile := tempDir + "/testfile.txt"

	// Create a temporary file
	file, err := os.Create(tempFile)
	assert.NoError(t, err)
	assert.NoError(t, file.Close())

	// Act
	err = run.validateModulePath(tempFile)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "module path is not a directory")
	assert.Contains(t, err.Error(), tempFile)
}

func TestUpgradeModule_Success(t *testing.T) {
	t.Skip("UpgradeModule requires extensive mocking of build, deployment, and application management flow; tested via integration tests")
}

func TestUpgradeModule_GetLatestApplicationError(t *testing.T) {
	t.Skip("UpgradeModule requires extensive mocking; tested via integration tests")
}

// ==================== RunLocalModule Rollback Tests ====================

func setupRunLocalModuleRollbackTest(t *testing.T) (*Run, *MockManagementSvc, *MockUpgradeModuleSvc, func()) {
	t.Helper()
	run, mockManagement, mockKeycloak, _, _, _ := newTestRun(action.RunLocalModule)
	mockUpgrade := &MockUpgradeModuleSvc{}
	run.Config.UpgradeModuleSvc = mockUpgrade

	originalParams := params
	params = action.Param{
		ModuleName:           "mod-x",
		ModuleVersion:        "1.0.0",
		ModulePath:           "",
		Namespace:            constant.SnapshotNamespace, // folioci -> shouldBuild=false
		ApplicationName:      "app-local",
		SkipModuleDeployment: true,
	}

	mockKeycloak.On("GetMasterAccessToken", mock.AnythingOfType("constant.KeycloakGrantType")).Return("master-token", nil)
	mockUpgrade.On("SetDefaultNamespaceIntoContext").Return()
	mockManagement.On("GetLatestApplication").Return(map[string]any{
		"name":    "app-combined",
		"version": "1.0.0",
		"modules": []any{map[string]any{"name": "mod-users"}},
	}, nil)

	return run, mockManagement, mockUpgrade, func() { params = originalParams }
}

func existingLocalApp() map[string]any {
	return map[string]any{
		"id":                  "app-local-1.0.1",
		"version":             "1.0.1",
		"dependencies":        []any{map[string]any{"name": "app-combined", "version": "1.0.0"}},
		"modules":             []any{map[string]any{"id": "mod-x-0.9.0", "name": "mod-x", "version": "0.9.0"}},
		"moduleDescriptors":   []any{},
		"uiModules":           []any{},
		"uiModuleDescriptors": []any{},
	}
}


func TestRunLocalModule_DiscoveryFailureKeepsPreviousVersion(t *testing.T) {
	// Arrange
	run, mockManagement, mockUpgrade, cleanup := setupRunLocalModuleRollbackTest(t)
	defer cleanup()

	mockManagement.On("GetLatestApplicationByName", "app-local").Return(existingLocalApp(), nil)
	mockManagement.On("CreateNewApplication", mock.Anything).Return(nil)
	mockManagement.On("CreateNewModuleDiscovery", mock.Anything).Return(assert.AnError)
	mockManagement.On("RemoveApplications", "app-local", "app-local-1.0.1").Return(nil)

	// Act
	err := run.RunLocalModule()

	// Assert
	assert.Error(t, err)
	assert.Equal(t, assert.AnError, err)
	mockManagement.AssertCalled(t, "RemoveApplications", "app-local", "app-local-1.0.1")
	mockManagement.AssertNotCalled(t, "UpgradeTenantEntitlement", mock.Anything, mock.Anything, mock.Anything)
	mockManagement.AssertNotCalled(t, "CreateTenantEntitlementForApplication", mock.Anything, mock.Anything, mock.Anything)
	mockManagement.AssertExpectations(t)
	mockUpgrade.AssertExpectations(t)
}


func TestRunLocalModule_DiscoveryFailureRemovesNewVersionWhenNoPrevious(t *testing.T) {
	// Arrange
	run, mockManagement, mockUpgrade, cleanup := setupRunLocalModuleRollbackTest(t)
	defer cleanup()

	mockManagement.On("GetLatestApplicationByName", "app-local").Return(nil, nil)
	mockManagement.On("CreateNewApplication", mock.Anything).Return(nil)
	mockManagement.On("CreateNewModuleDiscovery", mock.Anything).Return(assert.AnError)
	mockManagement.On("RemoveApplications", "app-local", "").Return(nil)

	// Act
	err := run.RunLocalModule()

	// Assert
	assert.Error(t, err)
	assert.Equal(t, assert.AnError, err)
	mockManagement.AssertCalled(t, "RemoveApplications", "app-local", "")
	mockManagement.AssertExpectations(t)
	mockUpgrade.AssertExpectations(t)
}


func TestRunLocalModule_EntitlementFailureKeepsPreviousVersion(t *testing.T) {
	// Arrange
	run, mockManagement, mockUpgrade, cleanup := setupRunLocalModuleRollbackTest(t)
	defer cleanup()

	mockManagement.On("GetLatestApplicationByName", "app-local").Return(existingLocalApp(), nil)
	mockManagement.On("CreateNewApplication", mock.Anything).Return(nil)
	mockManagement.On("CreateNewModuleDiscovery", mock.Anything).Return(nil)
	mockManagement.On("UpgradeTenantEntitlement", mock.Anything, mock.Anything, "app-local-1.0.2").Return(assert.AnError)
	mockManagement.On("RemoveApplications", "app-local", "app-local-1.0.1").Return(nil)

	// Act
	err := run.RunLocalModule()

	// Assert
	assert.Error(t, err)
	assert.Equal(t, assert.AnError, err)
	mockManagement.AssertCalled(t, "RemoveApplications", "app-local", "app-local-1.0.1")
	mockManagement.AssertExpectations(t)
	mockUpgrade.AssertExpectations(t)
}


func TestReserveUsedHostPorts_SeedsReservedPortsFromRunningContainers(t *testing.T) {
	// Arrange
	run, _, _, _, mockDocker, mockModule := newTestRun(action.RunLocalModule)

	mockDocker.On("Create").Return(nil, nil)
	mockDocker.On("Close", mock.Anything).Return()
	mockModule.On("GetDeployedModules", mock.Anything, mock.Anything).Return([]container.Summary{
		{Ports: []container.Port{{PublicPort: 30112}, {PublicPort: 30113}, {PublicPort: 0}}},
		{Ports: []container.Port{{PublicPort: 30114}}},
	}, nil)

	// Act
	err := run.reserveUsedHostPorts()

	// Assert
	assert.NoError(t, err)
	assert.Contains(t, run.Config.Action.ReservedPorts, 30112)
	assert.Contains(t, run.Config.Action.ReservedPorts, 30113)
	assert.Contains(t, run.Config.Action.ReservedPorts, 30114)
	assert.NotContains(t, run.Config.Action.ReservedPorts, 0) // unpublished ports (PublicPort 0) are skipped
	mockModule.AssertExpectations(t)
	mockDocker.AssertExpectations(t)
}

func TestRunLocalModule_FirstRunCreatesInitialVersion(t *testing.T) {
	// Arrange
	run, mockManagement, mockUpgrade, cleanup := setupRunLocalModuleRollbackTest(t)
	defer cleanup()

	mockManagement.On("GetLatestApplicationByName", "app-local").Return(nil, nil)
	mockManagement.On("CreateNewApplication", mock.MatchedBy(func(r *models.ApplicationUpgradeRequest) bool {
		return r.NewApplicationID == "app-local-1.0.0" && r.NewApplicationVersion == "1.0.0"
	})).Return(nil)
	mockManagement.On("CreateNewModuleDiscovery", mock.Anything).Return(nil)
	mockManagement.On("CreateTenantEntitlementForApplication", mock.Anything, mock.Anything, "app-local-1.0.0").Return(nil)
	mockManagement.On("RemoveApplications", "app-local", "app-local-1.0.0").Return(nil)

	// Act
	err := run.RunLocalModule()

	// Assert
	assert.NoError(t, err)
	mockManagement.AssertCalled(t, "CreateTenantEntitlementForApplication", mock.Anything, mock.Anything, "app-local-1.0.0")
	mockManagement.AssertExpectations(t)
	mockUpgrade.AssertExpectations(t)
}


func TestCleanupLocalAppOnFailure_OnlyRemovesAppVersions(t *testing.T) {
	// Arrange
	run, mockManagement, _, _, mockDocker, mockModule := newTestRun(action.RunLocalModule)

	originalParams := params
	defer func() { params = originalParams }()
	params = action.Param{ModuleName: "mod-x", ID: "mod-x-1.0.0"}

	mockManagement.On("RemoveApplications", "app-local", "app-local-1.0.1").Return(nil)

	// Act
	err := run.cleanupLocalAppOnFailure("app-local", "app-local-1.0.1")

	// Assert
	assert.NoError(t, err)
	mockManagement.AssertCalled(t, "RemoveApplications", "app-local", "app-local-1.0.1")
	mockManagement.AssertNotCalled(t, "RemoveModuleDiscovery", mock.Anything)
	mockDocker.AssertNotCalled(t, "Create")
	mockModule.AssertNotCalled(t, "UndeployModuleByNamePattern", mock.Anything, mock.Anything)
	mockManagement.AssertExpectations(t)
}

// ==================== DeployManagement Tests ====================

func TestDeployManagement_Success(t *testing.T) {
	// Arrange
	run, _, mockKeycloak, _, mockDocker, mockModule := newTestRun(action.DeployManagement)
	mockModuleProps := &MockModuleProps{}
	mockRegistrySvc := &MockRegistrySvc{}
	mockKongSvc := &MockKongSvc{}
	run.Config.ModuleProps = mockModuleProps
	run.Config.RegistrySvc = mockRegistrySvc
	run.Config.KongSvc = mockKongSvc

	mockModuleProps.On("ReadBackendModules", true, true).Return(map[string]models.BackendModule{}, nil)
	mockRegistrySvc.On("GetModules", true, true).Return(&models.ProxyModulesByRegistry{}, nil)
	mockRegistrySvc.On("ResolveModuleMetadata", mock.Anything).Return()
	mockDocker.On("Create").Return(nil, nil)
	mockModule.On("GetVaultRootToken", mock.Anything).Return("", nil)
	mockModule.On("DeployModules", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(map[string]int{"test-module": 8080}, 1, nil)
	mockModule.On("CheckModuleReadiness", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
	mockKongSvc.On("CheckRouteReadiness").Return(nil)
	mockKeycloak.On("GetMasterAccessToken", mock.Anything).Return("access-token", nil)
	mockKeycloak.On("UpdateRealmAccessTokenSettings", constant.KeycloakMasterRealm, mock.Anything).Return(nil)
	mockDocker.On("Close", mock.Anything).Return(nil)

	// Act
	err := run.DeployManagement()

	// Assert
	assert.NoError(t, err)
	mockKongSvc.AssertExpectations(t)
}

func TestDeployManagement_DeployModulesError(t *testing.T) {
	// Arrange
	run, _, _, _, mockDocker, mockModule := newTestRun(action.DeployManagement)
	mockModuleProps := &MockModuleProps{}
	mockRegistrySvc := &MockRegistrySvc{}
	run.Config.ModuleProps = mockModuleProps
	run.Config.RegistrySvc = mockRegistrySvc

	expectedError := assert.AnError
	mockModuleProps.On("ReadBackendModules", true, true).Return(map[string]models.BackendModule{}, nil)
	mockRegistrySvc.On("GetModules", true, true).Return(&models.ProxyModulesByRegistry{}, nil)
	mockRegistrySvc.On("ResolveModuleMetadata", mock.Anything).Return()
	mockDocker.On("Create").Return(nil, nil)
	mockModule.On("GetVaultRootToken", mock.Anything).Return("", nil)
	mockModule.On("DeployModules", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, 0, expectedError)
	mockDocker.On("Close", mock.Anything).Return(nil)

	// Act
	err := run.DeployManagement()

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
}

func TestDeployManagement_NoModulesDeployed(t *testing.T) {
	// Arrange
	run, _, _, _, mockDocker, mockModule := newTestRun(action.DeployManagement)
	mockModuleProps := &MockModuleProps{}
	mockRegistrySvc := &MockRegistrySvc{}
	run.Config.ModuleProps = mockModuleProps
	run.Config.RegistrySvc = mockRegistrySvc

	mockModuleProps.On("ReadBackendModules", true, true).Return(map[string]models.BackendModule{}, nil)
	mockRegistrySvc.On("GetModules", true, true).Return(&models.ProxyModulesByRegistry{}, nil)
	mockRegistrySvc.On("ResolveModuleMetadata", mock.Anything).Return()
	mockDocker.On("Create").Return(nil, nil)
	mockModule.On("GetVaultRootToken", mock.Anything).Return("", nil)
	mockModule.On("DeployModules", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(map[string]int{}, 0, nil)
	mockDocker.On("Close", mock.Anything).Return(nil)

	// Act
	err := run.DeployManagement()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "modules not deployed")
}

func TestDeployManagement_AllAlreadyDeployed_SkipsHealthcheck(t *testing.T) {
	// Arrange
	run, _, mockKeycloak, _, mockDocker, mockModule := newTestRun(action.DeployManagement)
	mockModuleProps := &MockModuleProps{}
	mockRegistrySvc := &MockRegistrySvc{}
	run.Config.ModuleProps = mockModuleProps
	run.Config.RegistrySvc = mockRegistrySvc

	mockModuleProps.On("ReadBackendModules", true, true).Return(map[string]models.BackendModule{}, nil)
	mockRegistrySvc.On("GetModules", true, true).Return(&models.ProxyModulesByRegistry{}, nil)
	mockRegistrySvc.On("ResolveModuleMetadata", mock.Anything).Return()
	mockDocker.On("Create").Return(nil, nil)
	mockModule.On("GetVaultRootToken", mock.Anything).Return("", nil)
	// newlyDeployed=empty (all already existed), totalMatched=1
	mockModule.On("DeployModules", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(map[string]int{}, 1, nil)
	mockKeycloak.On("GetMasterAccessToken", mock.Anything).Return("access-token", nil)
	mockKeycloak.On("UpdateRealmAccessTokenSettings", constant.KeycloakMasterRealm, mock.Anything).Return(nil)
	mockDocker.On("Close", mock.Anything).Return(nil)

	// Act
	err := run.DeployManagement()

	// Assert — no healthcheck, no kong check
	assert.NoError(t, err)
	mockModule.AssertNotCalled(t, "CheckModuleReadiness")
}

// ==================== UndeployManagement Tests ====================

func TestUndeployManagement_Success(t *testing.T) {
	// Arrange
	run, _, _, _, mockDocker, mockModule := newTestRun(action.UndeployManagement)

	mockDocker.On("Create").Return(nil, nil)
	mockModule.On("UndeployModuleByNamePattern", mock.Anything, constant.ManagementContainerPattern).Return(nil)
	mockDocker.On("Close", mock.Anything).Return(nil)

	// Act
	err := run.UndeployManagement()

	// Assert
	assert.NoError(t, err)
	mockModule.AssertExpectations(t)
}

func TestUndeployManagement_UndeployError(t *testing.T) {
	// Arrange
	run, _, _, _, mockDocker, mockModule := newTestRun(action.UndeployManagement)

	expectedError := assert.AnError
	mockDocker.On("Create").Return(nil, nil)
	mockModule.On("UndeployModuleByNamePattern", mock.Anything, constant.ManagementContainerPattern).Return(expectedError)
	mockDocker.On("Close", mock.Anything).Return(nil)

	// Act
	err := run.UndeployManagement()

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
}

// ==================== DeployApplication Tests ====================

func TestDeployApplication_Success(t *testing.T) {
	t.Skip("DeployApplication requires extensive mocking of multiple sub-commands; tested via integration tests")
}

func TestUndeployApplication_Success(t *testing.T) {
	t.Skip("UndeployApplication requires extensive mocking of multiple sub-commands; tested via integration tests")
}

// Note: DeployApplication and UndeployApplication are complex orchestration functions
// that call multiple other commands (DeploySystem, DeployManagement, DeployModules, etc.).
// These are better suited for integration tests rather than unit tests since they would
// require mocking every dependency of all sub-commands. The individual commands
// (DeploySystem, DeployManagement, etc.) have their own comprehensive unit tests.

// ==================== DeployAdditionalSystem Tests ====================

func TestDeployAdditionalSystem_NoContainers(t *testing.T) {
	// Arrange
	run, _, _, _, _, _ := newTestRun(action.DeployAdditionalSystem)
	run.Config.Action.ConfigBackendModules = nil // No additional containers

	// Act
	err := run.DeployAdditionalSystem()

	// Assert
	assert.NoError(t, err)
}

func TestUndeployAdditionalSystem_NoContainers(t *testing.T) {
	// Arrange
	run, _, _, _, _, _ := newTestRun(action.UndeployAdditionalSystem)
	run.Config.Action.ConfigBackendModules = nil // No additional containers

	// Act
	err := run.UndeployAdditionalSystem()

	// Assert
	assert.NoError(t, err)
}

// Note: Full testing of DeployAdditionalSystem and UndeployAdditionalSystem with containers
// requires mocking ExecSvc.ExecFromDir and ExecSvc.Exec, which execute docker compose commands.
// The core logic is minimal (building docker compose command), so integration tests are more appropriate.

// ==================== CreateConsortium Tests ====================

func TestCreateConsortium_NotSet(t *testing.T) {
	// Arrange
	run, _, _, _, _, _ := newTestRun(action.CreateConsortiums)
	run.Config.Action.ConfigConsortiums = nil // No consortiums configured

	// Act
	err := run.CreateConsortium()

	// Assert
	assert.NoError(t, err)
}

// ==================== ConsortiumPartition Tests ====================

func TestConsortiumPartition_NoConsortiums(t *testing.T) {
	// Arrange
	run, _, _, _, _, _ := newTestRun(action.DeployApplication)
	run.Config.Action.ConfigConsortiums = nil
	called := false
	fn := func(consortiumName string, tenantType constant.TenantType) error {
		called = true
		return nil
	}

	// Act
	err := run.ConsortiumPartition(fn)

	// Assert
	assert.NoError(t, err)
	assert.True(t, called) // Should be called once for NoneConsortium
}

func TestConsortiumPartition_WithConsortiums(t *testing.T) {
	// Arrange
	run, _, _, _, _, _ := newTestRun(action.DeployApplication)
	run.Config.Action.ConfigConsortiums = map[string]any{
		"consortium1": map[string]any{},
	}
	// Note: action.IsSet(field.Consortiums) will return false in tests since viper isn't set up,
	// so ConsortiumPartition will call fn once with NoneConsortium and Default
	callCount := 0
	var calledWith []string
	fn := func(consortiumName string, tenantType constant.TenantType) error {
		callCount++
		calledWith = append(calledWith, string(consortiumName)+":"+string(tenantType))
		return nil
	}

	// Act
	err := run.ConsortiumPartition(fn)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 1, callCount)
	assert.Contains(t, calledWith, constant.NoneConsortium+":"+string(constant.Default))
}

func TestConsortiumPartition_FunctionError(t *testing.T) {
	// Arrange
	run, _, _, _, _, _ := newTestRun(action.DeployApplication)
	run.Config.Action.ConfigConsortiums = nil
	expectedError := assert.AnError
	fn := func(consortiumName string, tenantType constant.TenantType) error {
		return expectedError
	}

	// Act
	err := run.ConsortiumPartition(fn)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
}

// ==================== CheckDeployedModuleReadiness Tests ====================

func TestCheckDeployedModuleReadiness_NoModules(t *testing.T) {
	// Arrange
	run, _, _, _, _, _ := newTestRun(action.DeployModules)
	modules := map[string]int{}

	// Act
	err := run.CheckDeployedModuleReadiness("backend", modules)

	// Assert
	assert.NoError(t, err)
}

func TestCheckDeployedModuleReadiness_WithModules(t *testing.T) {
	// Arrange
	run, _, _, _, _, mockModule := newTestRun(action.DeployModules)
	modules := map[string]int{
		"mod-test-1": 8081,
		"mod-test-2": 8082,
	}

	// CheckModuleReadiness is called once per module in a goroutine with WaitGroup
	mockModule.On("CheckModuleReadiness", mock.Anything, mock.Anything, "mod-test-1", 8081).Return()
	mockModule.On("CheckModuleReadiness", mock.Anything, mock.Anything, "mod-test-2", 8082).Return()

	// Act
	err := run.CheckDeployedModuleReadiness("backend", modules)

	// Assert
	assert.NoError(t, err)
	mockModule.AssertExpectations(t)
}

// ==================== ValidateParentApplications Tests ====================

func TestValidateParentApplications_AllPresent(t *testing.T) {
	// Arrange
	run, mockManagement, mockKeycloak, _, _, _ := newTestRun(action.DeployApplication)
	run.Config.Action.ConfigApplicationDependencies = map[string]any{
		"name":    "app-combined",
		"version": "1.0.0",
	}

	mockKeycloak.On("GetMasterAccessToken", mock.AnythingOfType("constant.KeycloakGrantType")).Return("token", nil)
	mockManagement.On("GetApplications").Return(models.ApplicationsResponse{
		ApplicationDescriptors: []map[string]any{
			{"id": "app-combined-1.0.0", "name": "app-combined"},
		},
		TotalRecords: 1,
	}, nil)

	// Act
	err := run.ValidateParentApplications()

	// Assert
	assert.NoError(t, err)
	mockKeycloak.AssertExpectations(t)
	mockManagement.AssertExpectations(t)
}

func TestValidateParentApplications_ParentMissing(t *testing.T) {
	// Arrange
	run, mockManagement, mockKeycloak, _, _, _ := newTestRun(action.DeployApplication)
	run.Config.Action.ConfigApplicationDependencies = map[string]any{
		"name":    "app-combined",
		"version": "1.0.0",
	}

	mockKeycloak.On("GetMasterAccessToken", mock.AnythingOfType("constant.KeycloakGrantType")).Return("token", nil)
	mockManagement.On("GetApplications").Return(models.ApplicationsResponse{
		ApplicationDescriptors: []map[string]any{},
		TotalRecords:           0,
	}, nil)

	// Act
	err := run.ValidateParentApplications()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "app-combined-1.0.0")
	mockKeycloak.AssertExpectations(t)
	mockManagement.AssertExpectations(t)
}

func TestValidateParentApplications_MultipleParentsMissing(t *testing.T) {
	// Arrange
	run, mockManagement, mockKeycloak, _, _, _ := newTestRun(action.DeployApplication)
	run.Config.Action.ConfigApplicationDependencies = map[string]any{
		"dep1": map[string]any{"name": "app-combined", "version": "1.0.0"},
		"dep2": map[string]any{"name": "app-platform", "version": "2.0.0"},
	}

	mockKeycloak.On("GetMasterAccessToken", mock.AnythingOfType("constant.KeycloakGrantType")).Return("token", nil)
	mockManagement.On("GetApplications").Return(models.ApplicationsResponse{
		ApplicationDescriptors: []map[string]any{},
		TotalRecords:           0,
	}, nil)

	// Act
	err := run.ValidateParentApplications()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parent application(s) not registered in mgr-applications")
	mockKeycloak.AssertExpectations(t)
	mockManagement.AssertExpectations(t)
}

func TestValidateParentApplications_MultipleParentsOnePresent(t *testing.T) {
	// Arrange
	run, mockManagement, mockKeycloak, _, _, _ := newTestRun(action.DeployApplication)
	run.Config.Action.ConfigApplicationDependencies = map[string]any{
		"dep1": map[string]any{"name": "app-combined", "version": "1.0.0"},
		"dep2": map[string]any{"name": "app-platform", "version": "2.0.0"},
	}

	mockKeycloak.On("GetMasterAccessToken", mock.AnythingOfType("constant.KeycloakGrantType")).Return("token", nil)
	mockManagement.On("GetApplications").Return(models.ApplicationsResponse{
		ApplicationDescriptors: []map[string]any{
			{"id": "app-combined-1.0.0", "name": "app-combined"},
		},
		TotalRecords: 1,
	}, nil)

	// Act
	err := run.ValidateParentApplications()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "app-platform-2.0.0")
	assert.NotContains(t, err.Error(), "app-combined-1.0.0")
	mockKeycloak.AssertExpectations(t)
	mockManagement.AssertExpectations(t)
}

func TestValidateParentApplications_GetApplicationsError(t *testing.T) {
	t.Run("NetworkError_WrapsWithNotReachable", func(t *testing.T) {
		// Arrange
		run, mockManagement, mockKeycloak, _, _, _ := newTestRun(action.DeployApplication)
		run.Config.Action.ConfigApplicationDependencies = map[string]any{
			"name": "app-combined", "version": "1.0.0",
		}
		netErr := &net.OpError{Op: "dial", Net: "tcp", Err: errors.New("connection refused")}

		mockKeycloak.On("GetMasterAccessToken", mock.AnythingOfType("constant.KeycloakGrantType")).Return("token", nil)
		mockManagement.On("GetApplications").Return(models.ApplicationsResponse{}, netErr)

		// Act
		err := run.ValidateParentApplications()

		// Assert
		assert.Error(t, err)
		assert.True(t, errors.Is(err, netErr))
		assert.Contains(t, err.Error(), "parent application services unreachable")
		mockKeycloak.AssertExpectations(t)
		mockManagement.AssertExpectations(t)
	})

	t.Run("NonNetworkError_PassesThrough", func(t *testing.T) {
		// Arrange
		run, mockManagement, mockKeycloak, _, _, _ := newTestRun(action.DeployApplication)
		run.Config.Action.ConfigApplicationDependencies = map[string]any{
			"name": "app-combined", "version": "1.0.0",
		}
		expectedError := errors.New("mgr-applications returned 500")

		mockKeycloak.On("GetMasterAccessToken", mock.AnythingOfType("constant.KeycloakGrantType")).Return("token", nil)
		mockManagement.On("GetApplications").Return(models.ApplicationsResponse{}, expectedError)

		// Act
		err := run.ValidateParentApplications()

		// Assert
		assert.Equal(t, expectedError, err)
		mockKeycloak.AssertExpectations(t)
		mockManagement.AssertExpectations(t)
	})
}

func TestValidateParentApplications_TokenError(t *testing.T) {
	t.Run("NetworkError_WrapsWithNotReachable", func(t *testing.T) {
		// Arrange
		run, _, mockKeycloak, _, _, _ := newTestRun(action.DeployApplication)
		run.Config.Action.ConfigApplicationDependencies = map[string]any{
			"name": "app-combined", "version": "1.0.0",
		}
		netErr := &net.OpError{Op: "dial", Net: "tcp", Err: errors.New("connection refused")}

		mockKeycloak.On("GetMasterAccessToken", mock.AnythingOfType("constant.KeycloakGrantType")).Return("", netErr)

		// Act
		err := run.ValidateParentApplications()

		// Assert
		assert.Error(t, err)
		assert.True(t, errors.Is(err, netErr))
		assert.Contains(t, err.Error(), "parent application services unreachable")
		mockKeycloak.AssertExpectations(t)
	})

	t.Run("NonNetworkError_PassesThrough", func(t *testing.T) {
		// Arrange
		run, _, mockKeycloak, _, _, _ := newTestRun(action.DeployApplication)
		run.Config.Action.ConfigApplicationDependencies = map[string]any{
			"name": "app-combined", "version": "1.0.0",
		}
		expectedError := errors.New("keycloak returned 401")

		mockKeycloak.On("GetMasterAccessToken", mock.AnythingOfType("constant.KeycloakGrantType")).Return("", expectedError)

		// Act
		err := run.ValidateParentApplications()

		// Assert
		assert.Equal(t, expectedError, err)
		mockKeycloak.AssertExpectations(t)
	})
}
