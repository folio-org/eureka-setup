package cmd

import (
	"bytes"
	"context"
	"os/exec"
	"sync"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/folio-org/eureka-setup/eureka-cli/action"
	"github.com/folio-org/eureka-setup/eureka-cli/constant"
	"github.com/folio-org/eureka-setup/eureka-cli/field"
	"github.com/folio-org/eureka-setup/eureka-cli/gitrepository"
	"github.com/folio-org/eureka-setup/eureka-cli/internal/testhelpers"
	"github.com/folio-org/eureka-setup/eureka-cli/models"
	"github.com/folio-org/eureka-setup/eureka-cli/modulesvc"
	"github.com/folio-org/eureka-setup/eureka-cli/runconfig"
	"github.com/go-git/go-git/v5/plumbing"
	vault "github.com/hashicorp/vault-client-go"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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

func (m *MockManagementSvc) RemoveApplications(applicationName, ignoreApplicationID string) error {
	args := m.Called(applicationName, ignoreApplicationID)
	return args.Error(0)
}

func (m *MockManagementSvc) GetModuleDiscovery(name string) (models.ModuleDiscoveryResponse, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return models.ModuleDiscoveryResponse{}, args.Error(1)
	}
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
	if args.Get(0) == nil {
		return models.TenantEntitlementResponse{}, args.Error(1)
	}
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

// MockKeycloakSvc is a mock for keycloaksvc.KeycloakProcessor
type MockKeycloakSvc struct {
	mock.Mock
}

func (m *MockKeycloakSvc) GetAccessToken(tenantName string) (string, error) {
	args := m.Called(tenantName)
	return args.String(0), args.Error(1)
}

func (m *MockKeycloakSvc) GetMasterAccessToken(grantType constant.KeycloakGrantType) (string, error) {
	args := m.Called(grantType)
	return args.String(0), args.Error(1)
}

func (m *MockKeycloakSvc) UpdateRealmAccessTokenSettings(tenantName string, lifespan int) error {
	args := m.Called(tenantName, lifespan)
	return args.Error(0)
}

func (m *MockKeycloakSvc) UpdatePublicClientSettings(tenantName string, url string) error {
	args := m.Called(tenantName, url)
	return args.Error(0)
}

func (m *MockKeycloakSvc) GetUsers(tenantName string) ([]any, error) {
	args := m.Called(tenantName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]any), args.Error(1)
}

func (m *MockKeycloakSvc) CreateUsers(configTenant string) error {
	args := m.Called(configTenant)
	return args.Error(0)
}

func (m *MockKeycloakSvc) RemoveUsers(tenantName string) error {
	args := m.Called(tenantName)
	return args.Error(0)
}

func (m *MockKeycloakSvc) GetRoles(headers map[string]string) ([]any, error) {
	args := m.Called(headers)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]any), args.Error(1)
}

func (m *MockKeycloakSvc) GetRoleByName(roleName string, headers map[string]string) (map[string]any, error) {
	args := m.Called(roleName, headers)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]any), args.Error(1)
}

func (m *MockKeycloakSvc) CreateRoles(configTenant string) error {
	args := m.Called(configTenant)
	return args.Error(0)
}

func (m *MockKeycloakSvc) RemoveRoles(tenantName string) error {
	args := m.Called(tenantName)
	return args.Error(0)
}

func (m *MockKeycloakSvc) GetCapabilitySets(headers map[string]string) ([]any, error) {
	args := m.Called(headers)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]any), args.Error(1)
}

func (m *MockKeycloakSvc) GetCapabilitySetsByName(headers map[string]string, capabilityName string) ([]any, error) {
	args := m.Called(headers, capabilityName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]any), args.Error(1)
}

func (m *MockKeycloakSvc) AttachCapabilitySets(tenantName string) error {
	args := m.Called(tenantName)
	return args.Error(0)
}

func (m *MockKeycloakSvc) DetachCapabilitySets(tenantName string) error {
	args := m.Called(tenantName)
	return args.Error(0)
}

func (m *MockKeycloakSvc) AttachCapabilitySetsToRoles(tenantName string) error {
	args := m.Called(tenantName)
	return args.Error(0)
}

func (m *MockKeycloakSvc) DetachCapabilitySetsFromRoles(tenantName string) error {
	args := m.Called(tenantName)
	return args.Error(0)
}

func (m *MockKeycloakSvc) UpdateKeycloakPublicClients(tenantName string) error {
	args := m.Called(tenantName)
	return args.Error(0)
}

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

// MockDockerClient is a mock for dockerclient.DockerClientRunner
type MockDockerClient struct {
	mock.Mock
}

func (m *MockDockerClient) Create() (*client.Client, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.Client), args.Error(1)
}

func (m *MockDockerClient) Close(cli *client.Client) {
	m.Called(cli)
}

func (m *MockDockerClient) PushImage(namespace string, imageName string) error {
	args := m.Called(namespace, imageName)
	return args.Error(0)
}

func (m *MockDockerClient) ForcePullImage(imageName string) (string, error) {
	args := m.Called(imageName)
	return args.String(0), args.Error(1)
}

// MockModuleSvc is a mock for modulesvc.ModuleProcessor
type MockModuleSvc struct {
	mock.Mock
}

func (m *MockModuleSvc) GetVaultRootToken(client *client.Client) (string, error) {
	args := m.Called(client)
	return args.String(0), args.Error(1)
}

func (m *MockModuleSvc) CheckModuleReadiness(wg *sync.WaitGroup, errCh chan<- error, moduleName string, port int) {
	defer wg.Done()
	m.Called(wg, errCh, moduleName, port)
}

func (m *MockModuleSvc) GetBackendModule(containers *models.Containers, moduleName string) (*models.BackendModule, *models.ProxyModule) {
	args := m.Called(containers, moduleName)
	if args.Get(0) == nil {
		return nil, nil
	}
	return args.Get(0).(*models.BackendModule), args.Get(1).(*models.ProxyModule)
}

func (m *MockModuleSvc) GetModuleImageVersion(backendModule models.BackendModule, module *models.ProxyModule) string {
	args := m.Called(backendModule, module)
	return args.String(0)
}

func (m *MockModuleSvc) GetSidecarImage(modules []*models.ProxyModule) (string, bool, error) {
	args := m.Called(modules)
	return args.String(0), args.Bool(1), args.Error(2)
}

func (m *MockModuleSvc) GetModuleImage(module *models.ProxyModule, moduleVersion string) string {
	args := m.Called(module, moduleVersion)
	return args.String(0)
}

func (m *MockModuleSvc) GetLocalModuleImage(namespace, moduleName, moduleVersion string) string {
	args := m.Called(namespace, moduleName, moduleVersion)
	return args.String(0)
}

func (m *MockModuleSvc) GetModuleEnv(container *models.Containers, module *models.ProxyModule, backendModule models.BackendModule) []string {
	args := m.Called(container, module, backendModule)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]string)
}

func (m *MockModuleSvc) GetSidecarEnv(containers *models.Containers, module *models.ProxyModule, backendModule models.BackendModule, moduleURL, sidecarURL string) []string {
	args := m.Called(containers, module, backendModule, moduleURL, sidecarURL)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]string)
}

func (m *MockModuleSvc) GetDeployedModules(cli *client.Client, f filters.Args) ([]container.Summary, error) {
	args := m.Called(cli, f)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]container.Summary), args.Error(1)
}

func (m *MockModuleSvc) PullModule(cli *client.Client, imageName string) error {
	args := m.Called(cli, imageName)
	return args.Error(0)
}

func (m *MockModuleSvc) DeployModules(cli *client.Client, containers *models.Containers, sidecarImage string, sidecarResources *container.Resources) (map[string]int, error) {
	args := m.Called(cli, containers, sidecarImage, sidecarResources)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]int), args.Error(1)
}

func (m *MockModuleSvc) DeployModule(cli *client.Client, container *models.Container) error {
	args := m.Called(cli, container)
	return args.Error(0)
}

func (m *MockModuleSvc) UndeployModuleByNamePattern(cli *client.Client, pattern string) error {
	args := m.Called(cli, pattern)
	return args.Error(0)
}

func (m *MockModuleSvc) UndeployModuleAndSidecarPair(cli *client.Client, pair *modulesvc.ModulePair) error {
	args := m.Called(cli, pair)
	return args.Error(0)
}

func (m *MockModuleSvc) DeployCustomModule(cli *client.Client, pair *modulesvc.ModulePair) error {
	args := m.Called(cli, pair)
	return args.Error(0)
}

func (m *MockModuleSvc) DeployCustomSidecar(cli *client.Client, pair *modulesvc.ModulePair) error {
	args := m.Called(cli, pair)
	return args.Error(0)
}

func (m *MockModuleSvc) CheckModuleAndSidecarReadiness(pair *modulesvc.ModulePair) error {
	args := m.Called(pair)
	return args.Error(0)
}

// MockInterceptModuleSvc is a mock for interceptmodulesvc.InterceptModuleProcessor
type MockInterceptModuleSvc struct {
	mock.Mock
}

func (m *MockInterceptModuleSvc) DeployDefaultModuleAndSidecarPair(cli *client.Client, pair *modulesvc.ModulePair) error {
	args := m.Called(cli, pair)
	return args.Error(0)
}

func (m *MockInterceptModuleSvc) DeployCustomSidecarForInterception(cli *client.Client, pair *modulesvc.ModulePair) error {
	args := m.Called(cli, pair)
	return args.Error(0)
}

// Helper function to create a test Run instance with mocks
func newTestRun(actionName string) (*Run, *MockManagementSvc, *MockKeycloakSvc, *MockVaultClient, *MockDockerClient, *MockModuleSvc) {
	mockAction := testhelpers.NewMockAction()
	mockAction.Name = actionName
	mockAction.KeycloakMasterAccessToken = "master-token"
	mockAction.ConfigTenants = map[string]any{
		"test-tenant": map[string]any{
			"name": "test-tenant",
		},
	}

	mockHTTP := &testhelpers.MockHTTPClient{}
	mockManagement := &MockManagementSvc{}
	mockKeycloak := &MockKeycloakSvc{}
	mockVault := &MockVaultClient{}
	mockDocker := &MockDockerClient{}
	mockModule := &MockModuleSvc{}

	config := &runconfig.RunConfig{
		Infrastructure: &runconfig.Infrastructure{
			Action:       mockAction,
			HTTPClient:   mockHTTP,
			VaultClient:  mockVault,
			DockerClient: mockDocker,
		},
		Services: &runconfig.Services{
			ManagementSvc: mockManagement,
			KeycloakSvc:   mockKeycloak,
			ModuleSvc:     mockModule,
		},
	}

	run := &Run{Config: config}
	return run, mockManagement, mockKeycloak, mockVault, mockDocker, mockModule
}

// MockExecSvc is a mock for execsvc.CommandRunner
type MockExecSvc struct {
	mock.Mock
}

func (m *MockExecSvc) Exec(cmd *exec.Cmd) error {
	args := m.Called(cmd)
	return args.Error(0)
}

func (m *MockExecSvc) ExecReturnOutput(cmd *exec.Cmd) (bytes.Buffer, bytes.Buffer, error) {
	args := m.Called(cmd)
	return args.Get(0).(bytes.Buffer), args.Get(1).(bytes.Buffer), args.Error(2)
}

func (m *MockExecSvc) ExecFromDir(cmd *exec.Cmd, dir string) error {
	args := m.Called(cmd, dir)
	return args.Error(0)
}

// MockUISvc is a mock for uisvc.UIProcessor
type MockUISvc struct {
	mock.Mock
}

func (m *MockUISvc) CloneAndUpdateRepository(updateCloned bool) (string, error) {
	args := m.Called(updateCloned)
	return args.String(0), args.Error(1)
}

func (m *MockUISvc) BuildImage(tenantName string, outputDir string) (string, error) {
	args := m.Called(tenantName, outputDir)
	return args.String(0), args.Error(1)
}

func (m *MockUISvc) DeployContainer(tenantName string, imageName string, externalPort int) error {
	args := m.Called(tenantName, imageName, externalPort)
	return args.Error(0)
}

func (m *MockUISvc) PrepareImage(tenantName string) (string, error) {
	args := m.Called(tenantName)
	return args.String(0), args.Error(1)
}

func (m *MockUISvc) PreparePackageJSON(configPath string) error {
	args := m.Called(configPath)
	return args.Error(0)
}

func (m *MockUISvc) GetStripesBranch() plumbing.ReferenceName {
	args := m.Called()
	return args.Get(0).(plumbing.ReferenceName)
}

func (m *MockUISvc) PrepareStripesConfigJS(tenantName string, configPath string) error {
	args := m.Called(tenantName, configPath)
	return args.Error(0)
}

// MockSearchSvc is a mock for searchsvc.SearchProcessor
type MockSearchSvc struct {
	mock.Mock
}

func (m *MockSearchSvc) ReindexInventoryRecords(tenantName string) error {
	args := m.Called(tenantName)
	return args.Error(0)
}

func (m *MockSearchSvc) ReindexInstanceRecords(tenantName string) error {
	args := m.Called(tenantName)
	return args.Error(0)
}

// MockKafkaSvc is a mock for kafkasvc.KafkaProcessor
type MockKafkaSvc struct {
	mock.Mock
}

func (m *MockKafkaSvc) CheckBrokerReadiness() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockKafkaSvc) PollConsumerGroup(tenantName string) error {
	args := m.Called(tenantName)
	return args.Error(0)
}

type MockModuleProps struct {
	mock.Mock
}

func (m *MockModuleProps) ReadBackendModules(checkIntegrity, skipFrontendCompatibility bool) (map[string]models.BackendModule, error) {
	args := m.Called(checkIntegrity, skipFrontendCompatibility)
	return args.Get(0).(map[string]models.BackendModule), args.Error(1)
}

func (m *MockModuleProps) ReadFrontendModules(checkIntegrity bool) (map[string]models.FrontendModule, error) {
	args := m.Called(checkIntegrity)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]models.FrontendModule), args.Error(1)
}

type MockRegistrySvc struct {
	mock.Mock
}

func (m *MockRegistrySvc) GetNamespace(version string) string {
	args := m.Called(version)
	return args.String(0)
}

func (m *MockRegistrySvc) GetModules(installJSONURLs map[string]string, useRemote, verbose bool) (*models.ProxyModulesByRegistry, error) {
	args := m.Called(installJSONURLs, useRemote, verbose)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ProxyModulesByRegistry), args.Error(1)
}

func (m *MockRegistrySvc) ExtractModuleMetadata(modules *models.ProxyModulesByRegistry) {
	m.Called(modules)
}

func (m *MockRegistrySvc) GetAuthorizationToken() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

// ==================== CreateTenants Tests ====================

func TestCreateTenants_Success(t *testing.T) {
	// Arrange
	run, mockManagement, mockKeycloak, _, _, _ := newTestRun(action.CreateTenants)

	mockKeycloak.On("GetMasterAccessToken", mock.AnythingOfType("constant.KeycloakGrantType")).Return("", nil)
	mockManagement.On("CreateTenants").Return(nil)

	// Act
	err := run.CreateTenants()

	// Assert
	assert.NoError(t, err)
	mockKeycloak.AssertExpectations(t)
	mockManagement.AssertExpectations(t)
}

func TestCreateTenants_GetTokenError(t *testing.T) {
	// Arrange
	run, mockManagement, mockKeycloak, _, _, _ := newTestRun(action.CreateTenants)

	expectedError := assert.AnError
	mockKeycloak.On("GetMasterAccessToken", mock.AnythingOfType("constant.KeycloakGrantType")).Return("", expectedError)

	// Act
	err := run.CreateTenants()

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockKeycloak.AssertExpectations(t)
	mockManagement.AssertNotCalled(t, "CreateTenants")
}

func TestCreateTenants_CreateTenantsError(t *testing.T) {
	// Arrange
	run, mockManagement, mockKeycloak, _, _, _ := newTestRun(action.CreateTenants)

	expectedError := assert.AnError
	mockKeycloak.On("GetMasterAccessToken", mock.AnythingOfType("constant.KeycloakGrantType")).Return("", nil)
	mockManagement.On("CreateTenants").Return(expectedError)

	// Act
	err := run.CreateTenants()

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockKeycloak.AssertExpectations(t)
	mockManagement.AssertExpectations(t)
}

// ==================== RemoveTenants Tests ====================

func TestRemoveTenants_Success(t *testing.T) {
	// Arrange
	run, mockManagement, mockKeycloak, _, _, _ := newTestRun(action.RemoveTenants)

	mockKeycloak.On("GetMasterAccessToken", mock.AnythingOfType("constant.KeycloakGrantType")).Return("", nil)
	mockManagement.On("RemoveTenants", mock.Anything, mock.Anything).Return(nil)

	// Act
	err := run.RemoveTenants(constant.NoneConsortium, constant.Default)

	// Assert
	assert.NoError(t, err)
	mockKeycloak.AssertExpectations(t)
	mockManagement.AssertExpectations(t)
}

func TestRemoveTenants_GetTokenError(t *testing.T) {
	// Arrange
	run, mockManagement, mockKeycloak, _, _, _ := newTestRun(action.RemoveTenants)

	expectedError := assert.AnError
	mockKeycloak.On("GetMasterAccessToken", mock.AnythingOfType("constant.KeycloakGrantType")).Return("", expectedError)

	// Act
	err := run.RemoveTenants(constant.NoneConsortium, constant.Default)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockKeycloak.AssertExpectations(t)
	mockManagement.AssertNotCalled(t, "RemoveTenants")
}

func TestRemoveTenants_RemoveTenantsError(t *testing.T) {
	// Arrange
	run, mockManagement, mockKeycloak, _, _, _ := newTestRun(action.RemoveTenants)

	expectedError := assert.AnError
	mockKeycloak.On("GetMasterAccessToken", mock.AnythingOfType("constant.KeycloakGrantType")).Return("", nil)
	mockManagement.On("RemoveTenants", mock.Anything, mock.Anything).Return(expectedError)

	// Act
	err := run.RemoveTenants(constant.NoneConsortium, constant.Default)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockKeycloak.AssertExpectations(t)
	mockManagement.AssertExpectations(t)
}

// ==================== CreateUsers Tests ====================

func TestCreateUsers_Success(t *testing.T) {
	// Arrange
	run, mockManagement, mockKeycloak, _, mockDocker, mockModule := newTestRun(action.CreateUsers)

	expectedTenant := map[string]any{"name": "test-tenant", "description": "nop-default"}

	mockDocker.On("Create").Return(nil, nil)
	mockModule.On("GetVaultRootToken", mock.Anything).Return("", nil)
	mockKeycloak.On("GetMasterAccessToken", mock.AnythingOfType("constant.KeycloakGrantType")).Return("", nil)
	mockManagement.On("GetTenants", mock.Anything, mock.Anything).
		Return([]any{expectedTenant}, nil)
	mockKeycloak.On("GetAccessToken", "test-tenant").Return("", nil)
	mockKeycloak.On("CreateUsers", "test-tenant").Return(nil)

	// Act
	err := run.CreateUsers(constant.NoneConsortium, constant.Default)

	// Assert
	assert.NoError(t, err)
	mockDocker.AssertExpectations(t)
	mockModule.AssertExpectations(t)
	mockKeycloak.AssertExpectations(t)
	mockManagement.AssertExpectations(t)
}

func TestCreateUsers_GetTenantsError(t *testing.T) {
	// Arrange
	run, mockManagement, mockKeycloak, _, mockDocker, mockModule := newTestRun(action.CreateUsers)

	expectedError := assert.AnError

	mockDocker.On("Create").Return(nil, nil)
	mockModule.On("GetVaultRootToken", mock.Anything).Return("", nil)
	mockKeycloak.On("GetMasterAccessToken", mock.AnythingOfType("constant.KeycloakGrantType")).Return("", nil)
	mockManagement.On("GetTenants", mock.Anything, mock.Anything).
		Return(nil, expectedError)

	// Act
	err := run.CreateUsers(constant.NoneConsortium, constant.Default)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockKeycloak.AssertNotCalled(t, "CreateUsers")
	mockKeycloak.AssertNotCalled(t, "GetAccessToken")
}

func TestCreateUsers_CreateUsersError(t *testing.T) {
	// Arrange
	run, mockManagement, mockKeycloak, _, mockDocker, mockModule := newTestRun(action.CreateUsers)

	expectedError := assert.AnError
	expectedTenant := map[string]any{"name": "test-tenant", "description": "nop-default"}

	mockDocker.On("Create").Return(nil, nil)
	mockModule.On("GetVaultRootToken", mock.Anything).Return("", nil)
	mockKeycloak.On("GetMasterAccessToken", mock.AnythingOfType("constant.KeycloakGrantType")).Return("", nil)
	mockManagement.On("GetTenants", mock.Anything, mock.Anything).
		Return([]any{expectedTenant}, nil)
	mockKeycloak.On("GetAccessToken", "test-tenant").Return("", nil)
	mockKeycloak.On("CreateUsers", "test-tenant").Return(expectedError)

	// Act
	err := run.CreateUsers(constant.NoneConsortium, constant.Default)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockKeycloak.AssertExpectations(t)
	mockManagement.AssertExpectations(t)
}

func TestCreateUsers_GetAccessTokenError(t *testing.T) {
	// Arrange
	run, mockManagement, mockKeycloak, _, mockDocker, mockModule := newTestRun(action.CreateUsers)

	expectedError := assert.AnError
	expectedTenant := map[string]any{"name": "test-tenant", "description": "nop-default"}

	mockDocker.On("Create").Return(nil, nil)
	mockModule.On("GetVaultRootToken", mock.Anything).Return("", nil)
	mockKeycloak.On("GetMasterAccessToken", mock.AnythingOfType("constant.KeycloakGrantType")).Return("", nil)
	mockManagement.On("GetTenants", mock.Anything, mock.Anything).
		Return([]any{expectedTenant}, nil)
	mockKeycloak.On("GetAccessToken", "test-tenant").Return("", expectedError)

	// Act
	err := run.CreateUsers(constant.NoneConsortium, constant.Default)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockKeycloak.AssertNotCalled(t, "CreateUsers")
}

// ==================== RemoveUsers Tests ====================

func TestRemoveUsers_Success(t *testing.T) {
	// Arrange
	run, mockManagement, mockKeycloak, _, mockDocker, mockModule := newTestRun(action.RemoveUsers)

	mockDocker.On("Create").Return(nil, nil)
	mockModule.On("GetVaultRootToken", mock.Anything).Return("", nil)
	mockKeycloak.On("GetMasterAccessToken", mock.AnythingOfType("constant.KeycloakGrantType")).Return("", nil)
	mockManagement.On("GetTenants", mock.Anything, mock.Anything).
		Return([]any{map[string]any{"name": "test-tenant", "description": "nop-default"}}, nil)
	mockKeycloak.On("GetAccessToken", mock.Anything).Return("", nil)
	mockKeycloak.On("RemoveUsers", "test-tenant").Return(nil)

	// Act
	err := run.RemoveUsers(constant.NoneConsortium, constant.Default)

	// Assert
	assert.NoError(t, err)
	mockDocker.AssertExpectations(t)
	// mockVault not used in these tests
	mockKeycloak.AssertExpectations(t)
	mockManagement.AssertExpectations(t)
}

func TestRemoveUsers_RemoveUsersError(t *testing.T) {
	// Arrange
	run, mockManagement, mockKeycloak, _, mockDocker, mockModule := newTestRun(action.RemoveUsers)

	expectedError := assert.AnError
	mockDocker.On("Create").Return(nil, nil)
	mockModule.On("GetVaultRootToken", mock.Anything).Return("", nil)
	mockKeycloak.On("GetMasterAccessToken", mock.AnythingOfType("constant.KeycloakGrantType")).Return("", nil)
	mockManagement.On("GetTenants", mock.Anything, mock.Anything).
		Return([]any{map[string]any{"name": "test-tenant", "description": "nop-default"}}, nil)
	mockKeycloak.On("GetAccessToken", mock.Anything).Return("", nil)
	mockKeycloak.On("RemoveUsers", "test-tenant").Return(expectedError)

	// Act
	err := run.RemoveUsers(constant.NoneConsortium, constant.Default)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockKeycloak.AssertExpectations(t)
}

// ==================== CreateRoles Tests ====================

func TestCreateRoles_Success(t *testing.T) {
	// Arrange
	run, mockManagement, mockKeycloak, _, mockDocker, mockModule := newTestRun(action.CreateRoles)

	mockDocker.On("Create").Return(nil, nil)
	mockModule.On("GetVaultRootToken", mock.Anything).Return("", nil)
	mockKeycloak.On("GetMasterAccessToken", mock.AnythingOfType("constant.KeycloakGrantType")).Return("", nil)
	mockManagement.On("GetTenants", mock.Anything, mock.Anything).
		Return([]any{map[string]any{"name": "test-tenant", "description": "nop-default"}}, nil)
	mockKeycloak.On("GetAccessToken", mock.Anything).Return("", nil)
	mockKeycloak.On("CreateRoles", "test-tenant").Return(nil)

	// Act
	err := run.CreateRoles(constant.NoneConsortium, constant.Default)

	// Assert
	assert.NoError(t, err)
	mockDocker.AssertExpectations(t)
	// mockVault not used in these tests
	mockKeycloak.AssertExpectations(t)
	mockManagement.AssertExpectations(t)
}

func TestCreateRoles_CreateRolesError(t *testing.T) {
	// Arrange
	run, mockManagement, mockKeycloak, _, mockDocker, mockModule := newTestRun(action.CreateRoles)

	expectedError := assert.AnError
	mockDocker.On("Create").Return(nil, nil)
	mockModule.On("GetVaultRootToken", mock.Anything).Return("", nil)
	mockKeycloak.On("GetMasterAccessToken", mock.AnythingOfType("constant.KeycloakGrantType")).Return("", nil)
	mockManagement.On("GetTenants", mock.Anything, mock.Anything).
		Return([]any{map[string]any{"name": "test-tenant", "description": "nop-default"}}, nil)
	mockKeycloak.On("GetAccessToken", mock.Anything).Return("", nil)
	mockKeycloak.On("CreateRoles", "test-tenant").Return(expectedError)

	// Act
	err := run.CreateRoles(constant.NoneConsortium, constant.Default)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockKeycloak.AssertExpectations(t)
}

// ==================== RemoveRoles Tests ====================

func TestRemoveRoles_Success(t *testing.T) {
	// Arrange
	run, mockManagement, mockKeycloak, _, mockDocker, mockModule := newTestRun(action.RemoveRoles)

	mockDocker.On("Create").Return(nil, nil)
	mockModule.On("GetVaultRootToken", mock.Anything).Return("", nil)
	mockKeycloak.On("GetMasterAccessToken", mock.AnythingOfType("constant.KeycloakGrantType")).Return("", nil)
	mockManagement.On("GetTenants", mock.Anything, mock.Anything).
		Return([]any{map[string]any{"name": "test-tenant", "description": "nop-default"}}, nil)
	mockKeycloak.On("GetAccessToken", mock.Anything).Return("", nil)
	mockKeycloak.On("RemoveRoles", "test-tenant").Return(nil)

	// Act
	err := run.RemoveRoles(constant.NoneConsortium, constant.Default)

	// Assert
	assert.NoError(t, err)
	mockDocker.AssertExpectations(t)
	// mockVault not used in these tests
	mockKeycloak.AssertExpectations(t)
	mockManagement.AssertExpectations(t)
}

func TestRemoveRoles_RemoveRolesError(t *testing.T) {
	// Arrange
	run, mockManagement, mockKeycloak, _, mockDocker, mockModule := newTestRun(action.RemoveRoles)

	expectedError := assert.AnError
	mockDocker.On("Create").Return(nil, nil)
	mockModule.On("GetVaultRootToken", mock.Anything).Return("", nil)
	mockKeycloak.On("GetMasterAccessToken", mock.AnythingOfType("constant.KeycloakGrantType")).Return("", nil)
	mockManagement.On("GetTenants", mock.Anything, mock.Anything).
		Return([]any{map[string]any{"name": "test-tenant", "description": "nop-default"}}, nil)
	mockKeycloak.On("GetAccessToken", mock.Anything).Return("", nil)
	mockKeycloak.On("RemoveRoles", "test-tenant").Return(expectedError)

	// Act
	err := run.RemoveRoles(constant.NoneConsortium, constant.Default)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockKeycloak.AssertExpectations(t)
}

// ==================== BuildSystem Tests ====================

func TestCloneUpdateRepositories_Success(t *testing.T) {
	// Arrange
	run, _, _, _, _, _ := newTestRun(action.BuildSystem)
	mockGitClient := &testhelpers.MockGitClient{}
	run.Config.GitClient = mockGitClient

	mockGitClient.On("KongRepository").Return(&gitrepository.GitRepository{}, nil)
	mockGitClient.On("KeycloakRepository").Return(&gitrepository.GitRepository{}, nil)
	mockGitClient.On("Clone", mock.Anything).Return(nil).Times(2)

	// Act
	err := run.CloneUpdateRepositories()

	// Assert
	assert.NoError(t, err)
	mockGitClient.AssertExpectations(t)
}

func TestCloneUpdateRepositories_KongRepoError(t *testing.T) {
	// Arrange
	run, _, _, _, _, _ := newTestRun(action.BuildSystem)
	mockGitClient := &testhelpers.MockGitClient{}
	run.Config.GitClient = mockGitClient

	expectedError := assert.AnError
	mockGitClient.On("KongRepository").Return(nil, expectedError)

	// Act
	err := run.CloneUpdateRepositories()

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockGitClient.AssertExpectations(t)
}

func TestCloneUpdateRepositories_KeycloakRepoError(t *testing.T) {
	// Arrange
	run, _, _, _, _, _ := newTestRun(action.BuildSystem)
	mockGitClient := &testhelpers.MockGitClient{}
	run.Config.GitClient = mockGitClient

	expectedError := assert.AnError
	mockGitClient.On("KongRepository").Return(&gitrepository.GitRepository{}, nil)
	mockGitClient.On("KeycloakRepository").Return(nil, expectedError)

	// Act
	err := run.CloneUpdateRepositories()

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockGitClient.AssertExpectations(t)
}

func TestBuildSystem_Success(t *testing.T) {
	// Arrange
	run, _, _, _, _, _ := newTestRun(action.BuildSystem)
	mockExecSvc := &MockExecSvc{}
	run.Config.ExecSvc = mockExecSvc

	mockExecSvc.On("ExecFromDir", mock.Anything, mock.Anything).Return(nil)

	// Act
	err := run.BuildSystem()

	// Assert
	assert.NoError(t, err)
	mockExecSvc.AssertExpectations(t)
}

func TestBuildSystem_Error(t *testing.T) {
	// Arrange
	run, _, _, _, _, _ := newTestRun(action.BuildSystem)
	mockExecSvc := &MockExecSvc{}
	run.Config.ExecSvc = mockExecSvc

	expectedError := assert.AnError
	mockExecSvc.On("ExecFromDir", mock.Anything, mock.Anything).Return(expectedError)

	// Act
	err := run.BuildSystem()

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockExecSvc.AssertExpectations(t)
}

// ==================== BuildAndPushUi Tests ====================

func TestBuildAndPushUi_Success(t *testing.T) {
	// Arrange
	run, _, _, _, mockDocker, _ := newTestRun(action.BuildAndPushUi)
	mockTenantSvc := &testhelpers.MockTenantSvc{}
	mockUISvc := &MockUISvc{}
	run.Config.TenantSvc = mockTenantSvc
	run.Config.UISvc = mockUISvc

	mockTenantSvc.On("SetConfigTenantParams", mock.Anything).Return(nil)
	mockUISvc.On("CloneAndUpdateRepository", mock.Anything).Return("/tmp/output", nil)
	mockUISvc.On("BuildImage", mock.Anything, mock.Anything).Return("test-image", nil)
	mockDocker.On("PushImage", mock.Anything, mock.Anything).Return(nil)

	// Act
	err := run.BuildAndPushUi()

	// Assert
	assert.NoError(t, err)
	mockTenantSvc.AssertExpectations(t)
	mockUISvc.AssertExpectations(t)
	mockDocker.AssertExpectations(t)
}

func TestBuildAndPushUi_SetConfigError(t *testing.T) {
	// Arrange
	run, _, _, _, _, _ := newTestRun(action.BuildAndPushUi)
	mockTenantSvc := &testhelpers.MockTenantSvc{}
	run.Config.TenantSvc = mockTenantSvc

	expectedError := assert.AnError
	mockTenantSvc.On("SetConfigTenantParams", mock.Anything).Return(expectedError)

	// Act
	err := run.BuildAndPushUi()

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockTenantSvc.AssertExpectations(t)
}

func TestBuildAndPushUi_CloneError(t *testing.T) {
	// Arrange
	run, _, _, _, _, _ := newTestRun(action.BuildAndPushUi)
	mockTenantSvc := &testhelpers.MockTenantSvc{}
	mockUISvc := &MockUISvc{}
	run.Config.TenantSvc = mockTenantSvc
	run.Config.UISvc = mockUISvc

	expectedError := assert.AnError
	mockTenantSvc.On("SetConfigTenantParams", mock.Anything).Return(nil)
	mockUISvc.On("CloneAndUpdateRepository", mock.Anything).Return("", expectedError)

	// Act
	err := run.BuildAndPushUi()

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockTenantSvc.AssertExpectations(t)
	mockUISvc.AssertExpectations(t)
}

func TestBuildAndPushUi_BuildImageError(t *testing.T) {
	// Arrange
	run, _, _, _, _, _ := newTestRun(action.BuildAndPushUi)
	mockTenantSvc := &testhelpers.MockTenantSvc{}
	mockUISvc := &MockUISvc{}
	run.Config.TenantSvc = mockTenantSvc
	run.Config.UISvc = mockUISvc

	expectedError := assert.AnError
	mockTenantSvc.On("SetConfigTenantParams", mock.Anything).Return(nil)
	mockUISvc.On("CloneAndUpdateRepository", mock.Anything).Return("/tmp/output", nil)
	mockUISvc.On("BuildImage", mock.Anything, mock.Anything).Return("", expectedError)

	// Act
	err := run.BuildAndPushUi()

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockTenantSvc.AssertExpectations(t)
	mockUISvc.AssertExpectations(t)
}

func TestBuildAndPushUi_PushImageError(t *testing.T) {
	// Arrange
	run, _, _, _, mockDocker, _ := newTestRun(action.BuildAndPushUi)
	mockTenantSvc := &testhelpers.MockTenantSvc{}
	mockUISvc := &MockUISvc{}
	run.Config.TenantSvc = mockTenantSvc
	run.Config.UISvc = mockUISvc

	expectedError := assert.AnError
	mockTenantSvc.On("SetConfigTenantParams", mock.Anything).Return(nil)
	mockUISvc.On("CloneAndUpdateRepository", mock.Anything).Return("/tmp/output", nil)
	mockUISvc.On("BuildImage", mock.Anything, mock.Anything).Return("test-image", nil)
	mockDocker.On("PushImage", mock.Anything, mock.Anything).Return(expectedError)

	// Act
	err := run.BuildAndPushUi()

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockTenantSvc.AssertExpectations(t)
	mockUISvc.AssertExpectations(t)
	mockDocker.AssertExpectations(t)
}

// ==================== ReindexIndices Tests ====================

func TestReindexIndices_Success(t *testing.T) {
	// Arrange
	viper.Set(field.Consortiums, map[string]any{"test-consortium": map[string]any{}})
	defer viper.Set(field.Consortiums, nil)

	run, mockManagement, mockKeycloak, _, mockDocker, mockModule := newTestRun(action.ReindexIndices)
	mockSearchSvc := &MockSearchSvc{}
	run.Config.SearchSvc = mockSearchSvc
	run.Config.Action.ConfigConsortiums = map[string]any{
		"test-consortium": map[string]any{
			"name": "test-consortium",
		},
	}
	run.Config.Action.ConfigTenants = map[string]any{
		"test-consortium-central": map[string]any{},
	}

	mockDocker.On("Create").Return(nil, nil)
	mockModule.On("GetVaultRootToken", mock.Anything).Return("", nil)
	mockKeycloak.On("GetMasterAccessToken", mock.AnythingOfType("constant.KeycloakGrantType")).Return("", nil)
	mockManagement.On("GetTenants", mock.Anything, mock.Anything).
		Return([]any{map[string]any{"name": "test-consortium-central", "description": "test-consortium-central"}}, nil)
	mockKeycloak.On("GetAccessToken", mock.Anything).Return("", nil)
	mockSearchSvc.On("ReindexInventoryRecords", "test-consortium-central").Return(nil)
	mockSearchSvc.On("ReindexInstanceRecords", "test-consortium-central").Return(nil)

	// Act
	err := run.ReindexIndices("test-consortium", constant.Central)

	// Assert
	assert.NoError(t, err)
	mockSearchSvc.AssertExpectations(t)
}

func TestReindexIndices_InventoryError(t *testing.T) {
	// Arrange
	viper.Set(field.Consortiums, map[string]any{"test-consortium": map[string]any{}})
	defer viper.Set(field.Consortiums, nil)

	run, mockManagement, mockKeycloak, _, mockDocker, mockModule := newTestRun(action.ReindexIndices)
	mockSearchSvc := &MockSearchSvc{}
	run.Config.SearchSvc = mockSearchSvc
	run.Config.Action.ConfigConsortiums = map[string]any{
		"test-consortium": map[string]any{
			"name": "test-consortium",
		},
	}
	run.Config.Action.ConfigTenants = map[string]any{
		"test-consortium-central": map[string]any{},
	}

	expectedError := assert.AnError
	mockDocker.On("Create").Return(nil, nil)
	mockModule.On("GetVaultRootToken", mock.Anything).Return("", nil)
	mockKeycloak.On("GetMasterAccessToken", mock.AnythingOfType("constant.KeycloakGrantType")).Return("", nil)
	mockManagement.On("GetTenants", mock.Anything, mock.Anything).
		Return([]any{map[string]any{"name": "test-consortium-central", "description": "test-consortium-central"}}, nil)
	mockKeycloak.On("GetAccessToken", mock.Anything).Return("", nil)
	mockSearchSvc.On("ReindexInventoryRecords", "test-consortium-central").Return(expectedError)

	// Act
	err := run.ReindexIndices("test-consortium", constant.Central)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockSearchSvc.AssertExpectations(t)
}

// ==================== DeployUi Tests ====================

func TestDeployUi_Success(t *testing.T) {
	// Arrange
	run, mockManagement, mockKeycloak, _, mockDocker, mockModule := newTestRun(action.DeployUi)
	mockTenantSvc := &testhelpers.MockTenantSvc{}
	mockUISvc := &MockUISvc{}
	run.Config.TenantSvc = mockTenantSvc
	run.Config.UISvc = mockUISvc
	run.Config.Action.ConfigTenants = map[string]any{
		"test-tenant": map[string]any{
			"deploy-ui": true,
		},
	}
	params.PlatformCompleteURL = "http://localhost:3000"

	mockDocker.On("Create").Return(nil, nil)
	mockModule.On("GetVaultRootToken", mock.Anything).Return("", nil)
	mockKeycloak.On("GetMasterAccessToken", mock.AnythingOfType("constant.KeycloakGrantType")).Return("", nil)
	mockManagement.On("GetTenants", mock.Anything, mock.Anything).
		Return([]any{map[string]any{"name": "test-tenant", "description": "nop-default"}}, nil)
	mockKeycloak.On("GetAccessToken", mock.Anything).Return("", nil)
	mockTenantSvc.On("SetConfigTenantParams", "test-tenant").Return(nil)
	mockUISvc.On("PrepareImage", "test-tenant").Return("test-image", nil)
	mockUISvc.On("DeployContainer", "test-tenant", "test-image", 3000).Return(nil)

	// Act
	err := run.DeployUi(constant.NoneConsortium, constant.Default)

	// Assert
	assert.NoError(t, err)
	mockTenantSvc.AssertExpectations(t)
	mockUISvc.AssertExpectations(t)
}

func TestDeployUi_SetConfigError(t *testing.T) {
	// Arrange
	run, mockManagement, mockKeycloak, _, mockDocker, mockModule := newTestRun(action.DeployUi)
	mockTenantSvc := &testhelpers.MockTenantSvc{}
	run.Config.TenantSvc = mockTenantSvc
	run.Config.Action.ConfigTenants = map[string]any{
		"test-tenant": map[string]any{
			"deploy-ui": true,
		},
	}

	expectedError := assert.AnError
	mockDocker.On("Create").Return(nil, nil)
	mockModule.On("GetVaultRootToken", mock.Anything).Return("", nil)
	mockKeycloak.On("GetMasterAccessToken", mock.AnythingOfType("constant.KeycloakGrantType")).Return("", nil)
	mockManagement.On("GetTenants", mock.Anything, mock.Anything).
		Return([]any{map[string]any{"name": "test-tenant", "description": "nop-default"}}, nil)
	mockKeycloak.On("GetAccessToken", mock.Anything).Return("", nil)
	mockTenantSvc.On("SetConfigTenantParams", "test-tenant").Return(expectedError)

	// Act
	err := run.DeployUi(constant.NoneConsortium, constant.Default)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockTenantSvc.AssertExpectations(t)
}

func TestDeployUi_PrepareImageError(t *testing.T) {
	// Arrange
	run, mockManagement, mockKeycloak, _, mockDocker, mockModule := newTestRun(action.DeployUi)
	mockTenantSvc := &testhelpers.MockTenantSvc{}
	mockUISvc := &MockUISvc{}
	run.Config.TenantSvc = mockTenantSvc
	run.Config.UISvc = mockUISvc
	run.Config.Action.ConfigTenants = map[string]any{
		"test-tenant": map[string]any{
			"deploy-ui": true,
		},
	}

	expectedError := assert.AnError
	mockDocker.On("Create").Return(nil, nil)
	mockModule.On("GetVaultRootToken", mock.Anything).Return("", nil)
	mockKeycloak.On("GetMasterAccessToken", mock.AnythingOfType("constant.KeycloakGrantType")).Return("", nil)
	mockManagement.On("GetTenants", mock.Anything, mock.Anything).
		Return([]any{map[string]any{"name": "test-tenant", "description": "nop-default"}}, nil)
	mockKeycloak.On("GetAccessToken", mock.Anything).Return("", nil)
	mockTenantSvc.On("SetConfigTenantParams", "test-tenant").Return(nil)
	mockUISvc.On("PrepareImage", "test-tenant").Return("", expectedError)

	// Act
	err := run.DeployUi(constant.NoneConsortium, constant.Default)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockUISvc.AssertExpectations(t)
}

// ==================== ListModules Tests ====================

func TestListModules_Success(t *testing.T) {
	// Arrange
	run, _, _, _, _, _ := newTestRun(action.ListModules)
	mockExecSvc := &MockExecSvc{}
	run.Config.ExecSvc = mockExecSvc
	params.ModuleName = "test-module"
	params.ModuleType = ""
	params.All = false

	mockExecSvc.On("Exec", mock.Anything).Return(nil)

	// Act
	err := run.ListModules()

	// Assert
	assert.NoError(t, err)
	mockExecSvc.AssertExpectations(t)
}

func TestListModules_ExecError(t *testing.T) {
	// Arrange
	run, _, _, _, _, _ := newTestRun(action.ListModules)
	mockExecSvc := &MockExecSvc{}
	run.Config.ExecSvc = mockExecSvc
	params.All = true

	expectedError := assert.AnError
	mockExecSvc.On("Exec", mock.Anything).Return(expectedError)

	// Act
	err := run.ListModules()

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockExecSvc.AssertExpectations(t)
}

func TestCreateFilter_All(t *testing.T) {
	// Arrange
	run, _, _, _, _, _ := newTestRun(action.ListModules)

	// Act
	result := run.createFilter("", "", true)

	// Assert
	assert.Equal(t, constant.AllContainerPattern, result)
}

func TestCreateFilter_SingleModule(t *testing.T) {
	// Arrange
	run, _, _, _, _, _ := newTestRun(action.ListModules)
	run.Config.Action.ConfigProfileName = "test-profile"

	// Act
	result := run.createFilter("test-module", "", false)

	// Assert
	assert.Contains(t, result, "test-profile")
	assert.Contains(t, result, "test-module")
}

// ==================== CheckPorts Tests ====================

func TestCheckPorts_Success(t *testing.T) {
	// Arrange
	run, _, _, _, mockDocker, mockModule := newTestRun(action.CheckPorts)
	mockExecSvc := &MockExecSvc{}
	run.Config.ExecSvc = mockExecSvc

	mockExecSvc.On("ExecFromDir", mock.Anything, mock.Anything).Return(nil)
	mockDocker.On("Create").Return(nil, nil)
	mockModule.On("GetDeployedModules", mock.Anything, mock.Anything).Return([]container.Summary{}, nil)
	mockDocker.On("Close", mock.Anything).Return(nil)

	// Act
	err := run.CheckPorts()

	// Assert
	assert.NoError(t, err)
	mockExecSvc.AssertExpectations(t)
}

func TestCheckPorts_ExecError(t *testing.T) {
	// Arrange
	run, _, _, _, _, _ := newTestRun(action.CheckPorts)
	mockExecSvc := &MockExecSvc{}
	run.Config.ExecSvc = mockExecSvc

	expectedError := assert.AnError
	mockExecSvc.On("ExecFromDir", mock.Anything, mock.Anything).Return(expectedError)

	// Act
	err := run.CheckPorts()

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockExecSvc.AssertExpectations(t)
}

// ==================== UpdateModuleDiscovery Tests ====================

func TestUpdateModuleDiscovery_Success(t *testing.T) {
	// Arrange
	run, mockManagement, mockKeycloak, _, _, _ := newTestRun(action.UpdateModuleDiscovery)
	params.ModuleName = "test-module"
	params.Restore = false
	params.PrivatePort = 8080

	mockKeycloak.On("GetMasterAccessToken", mock.AnythingOfType("constant.KeycloakGrantType")).Return("", nil)
	mockManagement.On("GetModuleDiscovery", "test-module").Return(models.ModuleDiscoveryResponse{
		Discovery: []models.ModuleDiscovery{
			{ID: "module-id-123", Name: "test-module"},
		},
	}, nil)
	mockManagement.On("UpdateModuleDiscovery", "module-id-123", false, 8080, mock.Anything).Return(nil)

	// Act
	err := run.UpdateModuleDiscovery()

	// Assert
	assert.NoError(t, err)
	mockManagement.AssertExpectations(t)
}

func TestUpdateModuleDiscovery_GetModuleError(t *testing.T) {
	// Arrange
	run, mockManagement, mockKeycloak, _, _, _ := newTestRun(action.UpdateModuleDiscovery)
	params.ModuleName = "test-module"

	expectedError := assert.AnError
	mockKeycloak.On("GetMasterAccessToken", mock.AnythingOfType("constant.KeycloakGrantType")).Return("", nil)
	mockManagement.On("GetModuleDiscovery", "test-module").Return(models.ModuleDiscoveryResponse{}, expectedError)

	// Act
	err := run.UpdateModuleDiscovery()

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockManagement.AssertExpectations(t)
}

func TestUpdateModuleDiscovery_UpdateError(t *testing.T) {
	// Arrange
	run, mockManagement, mockKeycloak, _, _, _ := newTestRun(action.UpdateModuleDiscovery)
	params.ModuleName = "test-module"
	params.Restore = true

	expectedError := assert.AnError
	mockKeycloak.On("GetMasterAccessToken", mock.AnythingOfType("constant.KeycloakGrantType")).Return("", nil)
	mockManagement.On("GetModuleDiscovery", "test-module").Return(models.ModuleDiscoveryResponse{
		Discovery: []models.ModuleDiscovery{
			{ID: "module-id-123", Name: "test-module"},
		},
	}, nil)
	mockManagement.On("UpdateModuleDiscovery", "module-id-123", true, mock.Anything, mock.Anything).Return(expectedError)

	// Act
	err := run.UpdateModuleDiscovery()

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockManagement.AssertExpectations(t)
}

// ==================== DeployModule Tests ====================

func TestDeployModule_Success(t *testing.T) {
	// Arrange
	run, _, _, _, _, _ := newTestRun(action.DeployModule)

	// Act
	err := run.DeployModule()

	// Assert
	assert.NoError(t, err)
}

// ==================== UndeployModule Tests ====================

func TestUndeployModule_Success(t *testing.T) {
	// Arrange
	run, _, _, _, mockDocker, mockModule := newTestRun(action.UndeployModule)
	params.ModuleName = "test-module"

	mockDocker.On("Create").Return(nil, nil)
	mockModule.On("UndeployModuleByNamePattern", mock.Anything, mock.Anything).Return(nil)
	mockDocker.On("Close", mock.Anything).Return(nil)

	// Act
	err := run.UndeployModule()

	// Assert
	assert.NoError(t, err)
	mockModule.AssertExpectations(t)
}

func TestUndeployModule_UndeployError(t *testing.T) {
	// Arrange
	run, _, _, _, mockDocker, mockModule := newTestRun(action.UndeployModule)
	params.ModuleName = "test-module"

	expectedError := assert.AnError
	mockDocker.On("Create").Return(nil, nil)
	mockModule.On("UndeployModuleByNamePattern", mock.Anything, mock.Anything).Return(expectedError)
	mockDocker.On("Close", mock.Anything).Return(nil)

	// Act
	err := run.UndeployModule()

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockModule.AssertExpectations(t)
}

// ==================== ListSystem Tests ====================

func TestListSystem_Success(t *testing.T) {
	// Arrange
	run, _, _, _, _, _ := newTestRun(action.ListSystem)
	mockExecSvc := &MockExecSvc{}
	run.Config.ExecSvc = mockExecSvc

	mockExecSvc.On("Exec", mock.Anything).Return(nil)

	// Act
	err := run.ListSystem()

	// Assert
	assert.NoError(t, err)
	mockExecSvc.AssertExpectations(t)
}

func TestListSystem_ExecError(t *testing.T) {
	// Arrange
	run, _, _, _, _, _ := newTestRun(action.ListSystem)
	mockExecSvc := &MockExecSvc{}
	run.Config.ExecSvc = mockExecSvc

	expectedError := assert.AnError
	mockExecSvc.On("Exec", mock.Anything).Return(expectedError)

	// Act
	err := run.ListSystem()

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockExecSvc.AssertExpectations(t)
}

// ==================== ListModuleVersions Tests ====================

func TestListModuleVersions_WithID(t *testing.T) {
	// Arrange
	run, _, _, _, _, _ := newTestRun(action.ListModuleVersions)
	mockHTTP := &testhelpers.MockHTTPClient{}
	run.Config.HTTPClient = mockHTTP
	params.ID = "test-module-1.0.0"

	mockHTTP.On("GetReturnRawBytes", mock.Anything, mock.Anything).Return([]byte(`{"id":"test-module-1.0.0"}`), nil)

	// Act
	err := run.ListModuleVersions()

	// Assert
	assert.NoError(t, err)
	mockHTTP.AssertExpectations(t)
}

func TestListModuleVersions_WithModuleName(t *testing.T) {
	// Arrange
	run, _, _, _, _, _ := newTestRun(action.ListModuleVersions)
	mockHTTP := &testhelpers.MockHTTPClient{}
	run.Config.HTTPClient = mockHTTP
	params.ID = ""
	params.ModuleName = "test-module"
	params.Versions = 10

	mockHTTP.On("GetRetryReturnStruct", mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			resp := args.Get(2).(*models.ProxyModulesResponse)
			*resp = models.ProxyModulesResponse{
				{ID: "test-module-1.0.0"},
				{ID: "test-module-0.9.0"},
			}
		}).Return(nil)

	// Act
	err := run.ListModuleVersions()

	// Assert
	assert.NoError(t, err)
	mockHTTP.AssertExpectations(t)
}

// ==================== AttachCapabilitySets Tests ====================

func TestAttachCapabilitySets_Success(t *testing.T) {
	// Arrange
	run, mockManagement, mockKeycloak, _, mockDocker, mockModule := newTestRun(action.AttachCapabilitySets)
	mockKafkaSvc := &MockKafkaSvc{}
	run.Config.KafkaSvc = mockKafkaSvc

	mockDocker.On("Create").Return(nil, nil)
	mockModule.On("GetVaultRootToken", mock.Anything).Return("", nil)
	mockKeycloak.On("GetMasterAccessToken", mock.AnythingOfType("constant.KeycloakGrantType")).Return("", nil)
	mockManagement.On("GetTenants", mock.Anything, mock.Anything).
		Return([]any{map[string]any{"name": "test-tenant", "description": "nop-default"}}, nil)
	mockKeycloak.On("UpdateRealmAccessTokenSettings", mock.Anything, mock.Anything).Return(nil)
	mockKeycloak.On("GetAccessToken", mock.Anything).Return("", nil)
	mockKafkaSvc.On("PollConsumerGroup", mock.Anything).Return(nil)
	mockKeycloak.On("AttachCapabilitySetsToRoles", "test-tenant").Return(nil)

	// Act
	err := run.AttachCapabilitySets(constant.NoneConsortium, constant.Default, 0)

	// Assert
	assert.NoError(t, err)
	mockKeycloak.AssertExpectations(t)
	mockKafkaSvc.AssertExpectations(t)
}

func TestAttachCapabilitySets_GetMasterTokenError(t *testing.T) {
	// Arrange
	run, _, mockKeycloak, _, _, _ := newTestRun(action.AttachCapabilitySets)

	expectedError := assert.AnError
	mockKeycloak.On("GetMasterAccessToken", mock.AnythingOfType("constant.KeycloakGrantType")).Return("", expectedError)

	// Act
	err := run.AttachCapabilitySets(constant.NoneConsortium, constant.Default, 0)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockKeycloak.AssertExpectations(t)
}

func TestAttachCapabilitySets_UpdateRealmError(t *testing.T) {
	// Arrange
	run, mockManagement, mockKeycloak, _, mockDocker, mockModule := newTestRun(action.AttachCapabilitySets)

	expectedError := assert.AnError
	mockDocker.On("Create").Return(nil, nil)
	mockModule.On("GetVaultRootToken", mock.Anything).Return("", nil)
	mockKeycloak.On("GetMasterAccessToken", mock.AnythingOfType("constant.KeycloakGrantType")).Return("", nil)
	mockManagement.On("GetTenants", mock.Anything, mock.Anything).
		Return([]any{map[string]any{"name": "test-tenant", "description": "nop-default"}}, nil)
	mockKeycloak.On("GetAccessToken", mock.Anything).Return("", nil)
	mockKeycloak.On("UpdateRealmAccessTokenSettings", mock.Anything, mock.Anything).Return(expectedError)

	// Act
	err := run.AttachCapabilitySets(constant.NoneConsortium, constant.Default, 0)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockKeycloak.AssertExpectations(t)
}

func TestAttachCapabilitySets_AttachError(t *testing.T) {
	// Arrange
	run, mockManagement, mockKeycloak, _, mockDocker, mockModule := newTestRun(action.AttachCapabilitySets)
	mockKafkaSvc := &MockKafkaSvc{}
	run.Config.KafkaSvc = mockKafkaSvc

	expectedError := assert.AnError
	mockDocker.On("Create").Return(nil, nil)
	mockModule.On("GetVaultRootToken", mock.Anything).Return("", nil)
	mockKeycloak.On("GetMasterAccessToken", mock.AnythingOfType("constant.KeycloakGrantType")).Return("", nil)
	mockManagement.On("GetTenants", mock.Anything, mock.Anything).
		Return([]any{map[string]any{"name": "test-tenant", "description": "nop-default"}}, nil)
	mockKeycloak.On("UpdateRealmAccessTokenSettings", mock.Anything, mock.Anything).Return(nil)
	mockKeycloak.On("GetAccessToken", mock.Anything).Return("", nil)
	mockKafkaSvc.On("PollConsumerGroup", mock.Anything).Return(nil)
	mockKeycloak.On("AttachCapabilitySetsToRoles", "test-tenant").Return(expectedError)

	// Act
	err := run.AttachCapabilitySets(constant.NoneConsortium, constant.Default, 0)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockKeycloak.AssertExpectations(t)
}

// ==================== DetachCapabilitySets Tests ====================

func TestDetachCapabilitySets_Success(t *testing.T) {
	// Arrange
	run, mockManagement, mockKeycloak, _, mockDocker, mockModule := newTestRun(action.DetachCapabilitySets)

	mockDocker.On("Create").Return(nil, nil)
	mockModule.On("GetVaultRootToken", mock.Anything).Return("", nil)
	mockKeycloak.On("GetMasterAccessToken", mock.AnythingOfType("constant.KeycloakGrantType")).Return("", nil)
	mockManagement.On("GetTenants", mock.Anything, mock.Anything).
		Return([]any{map[string]any{"name": "test-tenant", "description": "nop-default"}}, nil)
	mockKeycloak.On("GetAccessToken", mock.Anything).Return("", nil)
	mockKeycloak.On("DetachCapabilitySetsFromRoles", "test-tenant").Return(nil)

	// Act
	err := run.DetachCapabilitySets(constant.NoneConsortium, constant.Default)

	// Assert
	assert.NoError(t, err)
	mockKeycloak.AssertExpectations(t)
}

func TestDetachCapabilitySets_DetachError(t *testing.T) {
	// Arrange
	run, mockManagement, mockKeycloak, _, mockDocker, mockModule := newTestRun(action.DetachCapabilitySets)

	mockDocker.On("Create").Return(nil, nil)
	mockModule.On("GetVaultRootToken", mock.Anything).Return("", nil)
	mockKeycloak.On("GetMasterAccessToken", mock.AnythingOfType("constant.KeycloakGrantType")).Return("", nil)
	mockManagement.On("GetTenants", mock.Anything, mock.Anything).
		Return([]any{map[string]any{"name": "test-tenant", "description": "nop-default"}}, nil)
	mockKeycloak.On("GetAccessToken", mock.Anything).Return("", nil)
	mockKeycloak.On("DetachCapabilitySetsFromRoles", "test-tenant").Return(assert.AnError)

	// Act
	err := run.DetachCapabilitySets(constant.NoneConsortium, constant.Default)

	// Assert - function continues despite error
	assert.NoError(t, err)
	mockKeycloak.AssertExpectations(t)
}

// ==================== UpdateKeycloakPublicClients Tests ====================

func TestUpdateKeycloakPublicClients_Success(t *testing.T) {
	// Arrange
	run, mockManagement, mockKeycloak, _, mockDocker, mockModule := newTestRun(action.UpdateKeycloakPublicClients)
	mockTenantSvc := &testhelpers.MockTenantSvc{}
	run.Config.TenantSvc = mockTenantSvc
	run.Config.Action.ConfigTenants = map[string]any{
		"test-tenant": map[string]any{
			"deploy-ui": true,
		},
	}

	mockDocker.On("Create").Return(nil, nil)
	mockModule.On("GetVaultRootToken", mock.Anything).Return("", nil)
	mockKeycloak.On("GetMasterAccessToken", mock.AnythingOfType("constant.KeycloakGrantType")).Return("", nil)
	mockManagement.On("GetTenants", mock.Anything, mock.Anything).
		Return([]any{map[string]any{"name": "test-tenant", "description": "nop-default"}}, nil)
	mockKeycloak.On("GetAccessToken", mock.Anything).Return("", nil)
	mockTenantSvc.On("SetConfigTenantParams", "test-tenant").Return(nil)
	mockKeycloak.On("UpdatePublicClientSettings", "test-tenant", mock.Anything).Return(nil)

	// Act
	err := run.UpdateKeycloakPublicClients(constant.NoneConsortium, constant.Default)

	// Assert
	assert.NoError(t, err)
	mockKeycloak.AssertExpectations(t)
	mockTenantSvc.AssertExpectations(t)
}

func TestUpdateKeycloakPublicClients_SetConfigError(t *testing.T) {
	// Arrange
	run, mockManagement, mockKeycloak, _, mockDocker, mockModule := newTestRun(action.UpdateKeycloakPublicClients)
	mockTenantSvc := &testhelpers.MockTenantSvc{}
	run.Config.TenantSvc = mockTenantSvc
	run.Config.Action.ConfigTenants = map[string]any{
		"test-tenant": map[string]any{
			"deploy-ui": true,
		},
	}

	expectedError := assert.AnError
	mockDocker.On("Create").Return(nil, nil)
	mockModule.On("GetVaultRootToken", mock.Anything).Return("", nil)
	mockKeycloak.On("GetMasterAccessToken", mock.AnythingOfType("constant.KeycloakGrantType")).Return("", nil)
	mockManagement.On("GetTenants", mock.Anything, mock.Anything).
		Return([]any{map[string]any{"name": "test-tenant", "description": "nop-default"}}, nil)
	mockKeycloak.On("GetAccessToken", mock.Anything).Return("", nil)
	mockTenantSvc.On("SetConfigTenantParams", "test-tenant").Return(expectedError)

	// Act
	err := run.UpdateKeycloakPublicClients(constant.NoneConsortium, constant.Default)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockTenantSvc.AssertExpectations(t)
}

func TestUpdateKeycloakPublicClients_UpdateClientError(t *testing.T) {
	// Arrange
	run, mockManagement, mockKeycloak, _, mockDocker, mockModule := newTestRun(action.UpdateKeycloakPublicClients)
	mockTenantSvc := &testhelpers.MockTenantSvc{}
	run.Config.TenantSvc = mockTenantSvc
	run.Config.Action.ConfigTenants = map[string]any{
		"test-tenant": map[string]any{
			"deploy-ui": true,
		},
	}

	expectedError := assert.AnError
	mockDocker.On("Create").Return(nil, nil)
	mockModule.On("GetVaultRootToken", mock.Anything).Return("", nil)
	mockKeycloak.On("GetMasterAccessToken", mock.AnythingOfType("constant.KeycloakGrantType")).Return("", nil)
	mockManagement.On("GetTenants", mock.Anything, mock.Anything).
		Return([]any{map[string]any{"name": "test-tenant", "description": "nop-default"}}, nil)
	mockKeycloak.On("GetAccessToken", mock.Anything).Return("", nil)
	mockTenantSvc.On("SetConfigTenantParams", "test-tenant").Return(nil)
	mockKeycloak.On("UpdatePublicClientSettings", "test-tenant", mock.Anything).Return(expectedError)

	// Act
	err := run.UpdateKeycloakPublicClients(constant.NoneConsortium, constant.Default)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockKeycloak.AssertExpectations(t)
}

// ==================== GetVaultRootToken Tests ====================

func TestGetVaultRootToken_Success(t *testing.T) {
	// Arrange
	run, _, _, _, mockDocker, mockModule := newTestRun(action.GetVaultRootToken)

	expectedToken := "root-token-123"
	mockDocker.On("Create").Return(nil, nil)
	mockDocker.On("Close", mock.Anything).Return()
	mockModule.On("GetVaultRootToken", mock.Anything).Return(expectedToken, nil)

	// Act
	err := run.GetVaultRootToken()

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedToken, run.Config.Action.VaultRootToken)
	mockDocker.AssertExpectations(t)
	mockModule.AssertExpectations(t)
}

func TestGetVaultRootToken_CreateClientError(t *testing.T) {
	// Arrange
	run, _, _, _, mockDocker, _ := newTestRun(action.GetVaultRootToken)

	expectedError := assert.AnError
	mockDocker.On("Create").Return(nil, expectedError)

	// Act
	err := run.GetVaultRootToken()

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockDocker.AssertExpectations(t)
}

func TestGetVaultRootToken_GetTokenError(t *testing.T) {
	// Arrange
	run, _, _, _, mockDocker, mockModule := newTestRun(action.GetVaultRootToken)

	expectedError := assert.AnError
	mockDocker.On("Create").Return(nil, nil)
	mockDocker.On("Close", mock.Anything).Return()
	mockModule.On("GetVaultRootToken", mock.Anything).Return("", expectedError)

	// Act
	err := run.GetVaultRootToken()

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockDocker.AssertExpectations(t)
	mockModule.AssertExpectations(t)
}

// ==================== GetKeycloakAccessToken Tests ====================

func TestGetKeycloakAccessToken_MasterCustomToken(t *testing.T) {
	// Arrange
	run, _, mockKeycloak, _, _, _ := newTestRun(action.GetKeycloakAccessToken)

	expectedToken := "master-custom-token"
	mockKeycloak.On("GetMasterAccessToken", mock.AnythingOfType("constant.KeycloakGrantType")).Return(expectedToken, nil)

	// Act
	token, err := run.GetKeycloakAccessToken(constant.MasterCustomToken, "")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedToken, token)
	mockKeycloak.AssertExpectations(t)
}

func TestGetKeycloakAccessToken_MasterAdminCLIToken(t *testing.T) {
	// Arrange
	run, _, mockKeycloak, _, _, _ := newTestRun(action.GetKeycloakAccessToken)

	expectedToken := "master-admin-token"
	mockKeycloak.On("GetMasterAccessToken", mock.AnythingOfType("constant.KeycloakGrantType")).Return(expectedToken, nil)

	// Act
	token, err := run.GetKeycloakAccessToken(constant.MasterAdminCLIToken, "")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedToken, token)
	mockKeycloak.AssertExpectations(t)
}

func TestGetKeycloakAccessToken_TenantToken(t *testing.T) {
	// Arrange
	run, _, mockKeycloak, _, _, _ := newTestRun(action.GetKeycloakAccessToken)

	expectedToken := "tenant-token"
	run.Config.Action.KeycloakAccessToken = expectedToken
	mockKeycloak.On("GetAccessToken", "test-tenant").Return(expectedToken, nil)

	// Act
	token, err := run.GetKeycloakAccessToken("tenant", "test-tenant")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedToken, token)
	mockKeycloak.AssertExpectations(t)
}

func TestGetKeycloakAccessToken_TenantTokenMissingTenant(t *testing.T) {
	// Arrange
	run, _, _, _, _, _ := newTestRun(action.GetKeycloakAccessToken)

	// Act
	_, err := run.GetKeycloakAccessToken("tenant", "")

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tenant")
}

func TestGetKeycloakAccessToken_GetAccessTokenError(t *testing.T) {
	// Arrange
	run, _, mockKeycloak, _, _, _ := newTestRun(action.GetKeycloakAccessToken)

	expectedError := assert.AnError
	mockKeycloak.On("GetAccessToken", "test-tenant").Return("", expectedError)

	// Act
	_, err := run.GetKeycloakAccessToken("tenant", "test-tenant")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockKeycloak.AssertExpectations(t)
}

// ==================== GetEdgeApiKey Tests ====================

func TestGetEdgeApiKey_Success(t *testing.T) {
	// Arrange
	run, _, _, _, _, _ := newTestRun(action.GetEdgeApiKey)
	params.Length = 32
	params.Tenant = "test-tenant"
	params.User = "test-user"

	// Act
	err := run.GetEdgeApiKey()

	// Assert
	assert.NoError(t, err)
}

func TestGetRandomString_Success(t *testing.T) {
	// Arrange
	run, _, _, _, _, _ := newTestRun(action.GetEdgeApiKey)

	// Act
	result, err := run.getRandomString(16)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, result, 16)
}

func TestGetRandomString_ZeroLength(t *testing.T) {
	// Arrange
	run, _, _, _, _, _ := newTestRun(action.GetEdgeApiKey)

	// Act
	result, err := run.getRandomString(0)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, result, 0)
}

// ==================== CreateTenantEntitlements Tests ====================

func TestCreateTenantEntitlements_Success(t *testing.T) {
	// Arrange
	run, mockManagement, mockKeycloak, _, _, _ := newTestRun(action.CreateTenantEntitlements)

	mockKeycloak.On("GetMasterAccessToken", mock.AnythingOfType("constant.KeycloakGrantType")).Return("", nil)
	mockManagement.On("CreateTenantEntitlement", mock.Anything, mock.Anything).Return(nil)

	// Act
	err := run.CreateTenantEntitlements(constant.NoneConsortium, constant.Default)

	// Assert
	assert.NoError(t, err)
	mockKeycloak.AssertExpectations(t)
	mockManagement.AssertExpectations(t)
}

func TestCreateTenantEntitlements_GetMasterTokenError(t *testing.T) {
	// Arrange
	run, _, mockKeycloak, _, _, _ := newTestRun(action.CreateTenantEntitlements)

	expectedError := assert.AnError
	mockKeycloak.On("GetMasterAccessToken", mock.AnythingOfType("constant.KeycloakGrantType")).Return("", expectedError)

	// Act
	err := run.CreateTenantEntitlements(constant.NoneConsortium, constant.Default)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockKeycloak.AssertExpectations(t)
}

func TestCreateTenantEntitlements_CreateError(t *testing.T) {
	// Arrange
	run, mockManagement, mockKeycloak, _, _, _ := newTestRun(action.CreateTenantEntitlements)

	mockKeycloak.On("GetMasterAccessToken", mock.AnythingOfType("constant.KeycloakGrantType")).Return("", nil)
	mockManagement.On("CreateTenantEntitlement", mock.Anything, mock.Anything).Return(assert.AnError)

	// Act
	err := run.CreateTenantEntitlements(constant.NoneConsortium, constant.Default)

	// Assert - function returns nil despite error
	assert.NoError(t, err)
	mockKeycloak.AssertExpectations(t)
	mockManagement.AssertExpectations(t)
}

// ==================== RemoveTenantEntitlements Tests ====================

func TestRemoveTenantEntitlements_Success(t *testing.T) {
	// Arrange
	run, mockManagement, mockKeycloak, _, _, _ := newTestRun(action.RemoveTenantEntitlements)

	mockKeycloak.On("GetMasterAccessToken", mock.AnythingOfType("constant.KeycloakGrantType")).Return("", nil)
	mockManagement.On("RemoveTenantEntitlements", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// Act
	err := run.RemoveTenantEntitlements(constant.NoneConsortium, constant.Default)

	// Assert
	assert.NoError(t, err)
	mockKeycloak.AssertExpectations(t)
	mockManagement.AssertExpectations(t)
}

func TestRemoveTenantEntitlements_GetMasterTokenError(t *testing.T) {
	// Arrange
	run, _, mockKeycloak, _, _, _ := newTestRun(action.RemoveTenantEntitlements)

	expectedError := assert.AnError
	mockKeycloak.On("GetMasterAccessToken", mock.AnythingOfType("constant.KeycloakGrantType")).Return("", expectedError)

	// Act
	err := run.RemoveTenantEntitlements(constant.NoneConsortium, constant.Default)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockKeycloak.AssertExpectations(t)
}

func TestRemoveTenantEntitlements_RemoveError(t *testing.T) {
	// Arrange
	run, mockManagement, mockKeycloak, _, _, _ := newTestRun(action.RemoveTenantEntitlements)

	expectedError := assert.AnError
	mockKeycloak.On("GetMasterAccessToken", mock.AnythingOfType("constant.KeycloakGrantType")).Return("", nil)
	mockManagement.On("RemoveTenantEntitlements", mock.Anything, mock.Anything, mock.Anything).Return(expectedError)

	// Act
	err := run.RemoveTenantEntitlements(constant.NoneConsortium, constant.Default)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockKeycloak.AssertExpectations(t)
	mockManagement.AssertExpectations(t)
}

func TestReindexIndices_InstanceError(t *testing.T) {
	// Arrange
	viper.Set(field.Consortiums, map[string]any{"test-consortium": map[string]any{}})
	defer viper.Set(field.Consortiums, nil)

	run, mockManagement, mockKeycloak, _, mockDocker, mockModule := newTestRun(action.ReindexIndices)
	mockSearchSvc := &MockSearchSvc{}
	run.Config.SearchSvc = mockSearchSvc
	run.Config.Action.ConfigConsortiums = map[string]any{
		"test-consortium": map[string]any{
			"name": "test-consortium",
		},
	}
	run.Config.Action.ConfigTenants = map[string]any{
		"test-consortium-central": map[string]any{},
	}

	expectedError := assert.AnError
	mockDocker.On("Create").Return(nil, nil)
	mockModule.On("GetVaultRootToken", mock.Anything).Return("", nil)
	mockKeycloak.On("GetMasterAccessToken", mock.AnythingOfType("constant.KeycloakGrantType")).Return("", nil)
	mockManagement.On("GetTenants", mock.Anything, mock.Anything).
		Return([]any{map[string]any{"name": "test-consortium-central", "description": "test-consortium-central"}}, nil)
	mockKeycloak.On("GetAccessToken", mock.Anything).Return("", nil)
	mockSearchSvc.On("ReindexInventoryRecords", "test-consortium-central").Return(nil)
	mockSearchSvc.On("ReindexInstanceRecords", "test-consortium-central").Return(expectedError)

	// Act
	err := run.ReindexIndices("test-consortium", constant.Central)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockSearchSvc.AssertExpectations(t)
}

// ==================== UndeployUI Tests ====================

func TestUndeployUI_Success(t *testing.T) {
	// Arrange
	run, mockManagement, mockKeycloak, _, mockDocker, mockModule := newTestRun(action.UndeployUi)

	mockDocker.On("Create").Return(nil, nil)
	mockKeycloak.On("GetMasterAccessToken", mock.AnythingOfType("constant.KeycloakGrantType")).Return("", nil)
	mockManagement.On("GetTenants", mock.Anything, mock.Anything).
		Return([]any{map[string]any{"name": "test-tenant"}}, nil)
	mockModule.On("UndeployModuleByNamePattern", mock.Anything, mock.Anything).Return(nil)
	mockDocker.On("Close", mock.Anything).Return(nil)

	// Act
	err := run.UndeployUI()

	// Assert
	assert.NoError(t, err)
	mockModule.AssertExpectations(t)
}

func TestUndeployUI_UndeployError(t *testing.T) {
	// Arrange
	run, mockManagement, mockKeycloak, _, mockDocker, mockModule := newTestRun(action.UndeployUi)

	expectedError := assert.AnError
	mockDocker.On("Create").Return(nil, nil)
	mockKeycloak.On("GetMasterAccessToken", mock.AnythingOfType("constant.KeycloakGrantType")).Return("", nil)
	mockManagement.On("GetTenants", mock.Anything, mock.Anything).
		Return([]any{map[string]any{"name": "test-tenant"}}, nil)
	mockModule.On("UndeployModuleByNamePattern", mock.Anything, mock.Anything).Return(expectedError)
	mockDocker.On("Close", mock.Anything).Return(nil)

	// Act
	err := run.UndeployUI()

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
}

// ==================== UndeploySystem Tests ====================

func TestUndeploySystem_Success(t *testing.T) {
	// Arrange
	run, _, _, _, _, _ := newTestRun(action.UndeploySystem)
	mockExecSvc := &MockExecSvc{}
	run.Config.ExecSvc = mockExecSvc

	mockExecSvc.On("Exec", mock.Anything).Return(nil)

	// Act
	err := run.UndeploySystem()

	// Assert
	assert.NoError(t, err)
	mockExecSvc.AssertExpectations(t)
}

func TestUndeploySystem_ExecError(t *testing.T) {
	// Arrange
	run, _, _, _, _, _ := newTestRun(action.UndeploySystem)
	mockExecSvc := &MockExecSvc{}
	run.Config.ExecSvc = mockExecSvc

	expectedError := assert.AnError
	mockExecSvc.On("Exec", mock.Anything).Return(expectedError)

	// Act
	err := run.UndeploySystem()

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
}

// ==================== UndeployModules Tests ====================

func TestUndeployModules_WithoutRemoveApplication(t *testing.T) {
	// Arrange
	run, _, _, _, mockDocker, mockModule := newTestRun(action.UndeployModules)

	mockDocker.On("Create").Return(nil, nil)
	mockModule.On("UndeployModuleByNamePattern", mock.Anything, mock.Anything).Return(nil)
	mockDocker.On("Close", mock.Anything).Return(nil)

	// Act
	err := run.UndeployModules(false)

	// Assert
	assert.NoError(t, err)
	mockModule.AssertExpectations(t)
}

func TestUndeployModules_WithRemoveApplication(t *testing.T) {
	// Arrange
	run, mockManagement, mockKeycloak, _, mockDocker, mockModule := newTestRun(action.UndeployModules)

	mockKeycloak.On("GetMasterAccessToken", mock.AnythingOfType("constant.KeycloakGrantType")).Return("", nil)
	mockManagement.On("RemoveApplication", mock.Anything).Return(nil)
	mockDocker.On("Create").Return(nil, nil)
	mockModule.On("UndeployModuleByNamePattern", mock.Anything, mock.Anything).Return(nil)
	mockDocker.On("Close", mock.Anything).Return(nil)

	// Act
	err := run.UndeployModules(true)

	// Assert
	assert.NoError(t, err)
	mockManagement.AssertExpectations(t)
}

func TestUndeployModules_UndeployError(t *testing.T) {
	// Arrange
	run, _, _, _, mockDocker, mockModule := newTestRun(action.UndeployModules)

	expectedError := assert.AnError
	mockDocker.On("Create").Return(nil, nil)
	mockModule.On("UndeployModuleByNamePattern", mock.Anything, mock.Anything).Return(expectedError)
	mockDocker.On("Close", mock.Anything).Return(nil)

	// Act
	err := run.UndeployModules(false)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
}

// ==================== DeployModules Tests ====================

func TestDeployModules_Success(t *testing.T) {
	// Arrange
	run, mockManagement, mockKeycloak, _, mockDocker, mockModule := newTestRun(action.DeployModules)
	mockModuleProps := &MockModuleProps{}
	mockRegistrySvc := &MockRegistrySvc{}
	run.Config.ModuleProps = mockModuleProps
	run.Config.RegistrySvc = mockRegistrySvc

	mockModuleProps.On("ReadBackendModules", false, true).Return(map[string]models.BackendModule{}, nil)
	mockModuleProps.On("ReadFrontendModules", true).Return(map[string]models.FrontendModule{}, nil)
	mockRegistrySvc.On("GetModules", mock.Anything, true, true).Return(&models.ProxyModulesByRegistry{}, nil)
	mockRegistrySvc.On("ExtractModuleMetadata", mock.Anything).Return()
	mockDocker.On("Create").Return(nil, nil)
	mockModule.On("GetVaultRootToken", mock.Anything).Return("", nil)
	mockModule.On("GetSidecarImage", mock.Anything).Return("test-sidecar:latest", false, nil)
	mockModule.On("DeployModules", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(map[string]int{"test-module": 8080}, nil)
	mockModule.On("CheckModuleReadiness", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
	mockKeycloak.On("GetMasterAccessToken", mock.Anything).Return("access-token", nil)
	mockManagement.On("CreateApplication", mock.Anything).Return(nil)
	mockDocker.On("Close", mock.Anything).Return(nil)

	// Act
	err := run.DeployModules()

	// Assert
	assert.NoError(t, err)
}

// ==================== DeploySystem Tests ====================

func TestDeploySystem_Success(t *testing.T) {
	// Arrange
	run, _, _, _, _, _ := newTestRun(action.DeploySystem)
	mockGitClient := &testhelpers.MockGitClient{}
	mockExecSvc := &MockExecSvc{}
	run.Config.GitClient = mockGitClient
	run.Config.ExecSvc = mockExecSvc
	params.BuildImages = false

	mockGitClient.On("KongRepository").Return(&gitrepository.GitRepository{}, nil)
	mockGitClient.On("KeycloakRepository").Return(&gitrepository.GitRepository{}, nil)
	mockGitClient.On("Clone", mock.Anything).Return(nil)
	mockExecSvc.On("ExecFromDir", mock.Anything, mock.Anything).Return(nil)

	// Act
	err := run.DeploySystem()

	// Assert
	assert.NoError(t, err)
	mockExecSvc.AssertExpectations(t)
}

func TestDeploySystem_ExecError(t *testing.T) {
	// Arrange
	run, _, _, _, _, _ := newTestRun(action.DeploySystem)
	mockGitClient := &testhelpers.MockGitClient{}
	mockExecSvc := &MockExecSvc{}
	run.Config.GitClient = mockGitClient
	run.Config.ExecSvc = mockExecSvc
	params.BuildImages = false

	expectedError := assert.AnError
	mockGitClient.On("KongRepository").Return(&gitrepository.GitRepository{}, nil)
	mockGitClient.On("KeycloakRepository").Return(&gitrepository.GitRepository{}, nil)
	mockGitClient.On("Clone", mock.Anything).Return(nil)
	mockExecSvc.On("ExecFromDir", mock.Anything, mock.Anything).Return(expectedError)

	// Act
	err := run.DeploySystem()

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
}

// ==================== InterceptModule Tests ====================

func TestInterceptModule_Success(t *testing.T) {
	// Arrange
	run, mockManagement, mockKeycloak, _, mockDocker, mockModule := newTestRun(action.InterceptModule)
	mockModuleProps := &MockModuleProps{}
	mockRegistrySvc := &MockRegistrySvc{}
	mockInterceptSvc := &MockInterceptModuleSvc{}
	run.Config.ModuleProps = mockModuleProps
	run.Config.RegistrySvc = mockRegistrySvc
	run.Config.InterceptModuleSvc = mockInterceptSvc
	params.ModuleName = "test-module"
	params.Restore = false

	mockKeycloak.On("GetMasterAccessToken", mock.Anything).Return("access-token", nil)
	mockManagement.On("GetModuleDiscovery", "test-module").Return(models.ModuleDiscoveryResponse{
		Discovery: []models.ModuleDiscovery{{ID: "module-id-123", Name: "test-module"}},
	}, nil)
	mockModuleProps.On("ReadBackendModules", false, false).Return(map[string]models.BackendModule{}, nil)
	mockRegistrySvc.On("GetModules", mock.Anything, true, false).Return(&models.ProxyModulesByRegistry{}, nil)
	mockRegistrySvc.On("ExtractModuleMetadata", mock.Anything).Return()
	mockDocker.On("Create").Return(nil, nil)
	mockModule.On("GetVaultRootToken", mock.Anything).Return("", nil)
	mockInterceptSvc.On("DeployCustomSidecarForInterception", mock.Anything, mock.Anything).Return(nil)
	mockDocker.On("Close", mock.Anything).Return(nil)

	// Act
	err := run.InterceptModule()

	// Assert
	assert.NoError(t, err)
}

func TestInterceptModule_InterceptError(t *testing.T) {
	// Arrange
	run, mockManagement, mockKeycloak, _, mockDocker, mockModule := newTestRun(action.InterceptModule)
	mockModuleProps := &MockModuleProps{}
	mockRegistrySvc := &MockRegistrySvc{}
	mockInterceptSvc := &MockInterceptModuleSvc{}
	run.Config.ModuleProps = mockModuleProps
	run.Config.RegistrySvc = mockRegistrySvc
	run.Config.InterceptModuleSvc = mockInterceptSvc
	params.ModuleName = "test-module"
	params.Restore = false

	expectedError := assert.AnError
	mockKeycloak.On("GetMasterAccessToken", mock.Anything).Return("access-token", nil)
	mockManagement.On("GetModuleDiscovery", "test-module").Return(models.ModuleDiscoveryResponse{
		Discovery: []models.ModuleDiscovery{{ID: "module-id-123", Name: "test-module"}},
	}, nil)
	mockModuleProps.On("ReadBackendModules", false, false).Return(map[string]models.BackendModule{}, nil)
	mockRegistrySvc.On("GetModules", mock.Anything, true, false).Return(&models.ProxyModulesByRegistry{}, nil)
	mockRegistrySvc.On("ExtractModuleMetadata", mock.Anything).Return()
	mockDocker.On("Create").Return(nil, nil)
	mockModule.On("GetVaultRootToken", mock.Anything).Return("", nil)
	mockInterceptSvc.On("DeployCustomSidecarForInterception", mock.Anything, mock.Anything).Return(expectedError)
	mockDocker.On("Close", mock.Anything).Return(nil)

	// Act
	err := run.InterceptModule()

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
}
