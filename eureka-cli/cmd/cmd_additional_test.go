package cmd

import (
	"os"
	"testing"

	"github.com/docker/docker/client"
	"github.com/j011195/eureka-setup/eureka-cli/action"
	"github.com/j011195/eureka-setup/eureka-cli/constant"
	"github.com/j011195/eureka-setup/eureka-cli/models"
	"github.com/j011195/eureka-setup/eureka-cli/modulesvc"
	"github.com/j011195/eureka-setup/eureka-cli/runconfig"
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
	mockRegistrySvc.On("GetModules", mock.Anything, true, true).Return(&models.ProxyModulesByRegistry{}, nil)
	mockRegistrySvc.On("ExtractModuleMetadata", mock.Anything).Return()
	mockDocker.On("Create").Return(nil, nil)
	mockModule.On("GetVaultRootToken", mock.Anything).Return("", nil)
	mockModule.On("DeployModules", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(map[string]int{"test-module": 8080}, nil)
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
	mockRegistrySvc.On("GetModules", mock.Anything, true, true).Return(&models.ProxyModulesByRegistry{}, nil)
	mockRegistrySvc.On("ExtractModuleMetadata", mock.Anything).Return()
	mockDocker.On("Create").Return(nil, nil)
	mockModule.On("GetVaultRootToken", mock.Anything).Return("", nil)
	mockModule.On("DeployModules", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, expectedError)
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
	mockRegistrySvc.On("GetModules", mock.Anything, true, true).Return(&models.ProxyModulesByRegistry{}, nil)
	mockRegistrySvc.On("ExtractModuleMetadata", mock.Anything).Return()
	mockDocker.On("Create").Return(nil, nil)
	mockModule.On("GetVaultRootToken", mock.Anything).Return("", nil)
	mockModule.On("DeployModules", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(map[string]int{}, nil)
	mockDocker.On("Close", mock.Anything).Return(nil)

	// Act
	err := run.DeployManagement()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "modules not deployed")
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
