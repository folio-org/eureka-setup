package testhelpers

import (
	"bytes"
	"net/url"
	"os/exec"

	"github.com/docker/docker/client"
	"github.com/j011195/eureka-setup/eureka-cli/action"
	"github.com/j011195/eureka-setup/eureka-cli/models"
	"github.com/stretchr/testify/mock"
)

// MockHTTPClient is a mock implementation of httpclient.HTTPClientRunner
type MockHTTPClient struct {
	mock.Mock
}

func (m *MockHTTPClient) PingRetry(url string) error {
	args := m.Called(url)
	return args.Error(0)
}

func (m *MockHTTPClient) Ping(url string) (int, error) {
	args := m.Called(url)
	return args.Int(0), args.Error(1)
}

func (m *MockHTTPClient) GetReturnRawBytes(url string, headers map[string]string) ([]byte, error) {
	args := m.Called(url, headers)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockHTTPClient) GetReturnStruct(url string, headers map[string]string, target any) error {
	args := m.Called(url, headers, target)
	return args.Error(0)
}

func (m *MockHTTPClient) GetRetryReturnStruct(url string, headers map[string]string, target any) error {
	args := m.Called(url, headers, target)
	return args.Error(0)
}

func (m *MockHTTPClient) PostReturnNoContent(url string, payload []byte, headers map[string]string) error {
	args := m.Called(url, payload, headers)
	return args.Error(0)
}

func (m *MockHTTPClient) PostRetryReturnNoContent(url string, payload []byte, headers map[string]string) error {
	args := m.Called(url, payload, headers)
	return args.Error(0)
}

func (m *MockHTTPClient) PostReturnStruct(url string, payload []byte, headers map[string]string, target any) error {
	args := m.Called(url, payload, headers, target)
	return args.Error(0)
}

func (m *MockHTTPClient) PostFormDataReturnStruct(urlStr string, formValues url.Values, headers map[string]string, target any) error {
	args := m.Called(urlStr, formValues, headers, target)
	return args.Error(0)
}

func (m *MockHTTPClient) PutReturnNoContent(url string, payload []byte, headers map[string]string) error {
	args := m.Called(url, payload, headers)
	return args.Error(0)
}

func (m *MockHTTPClient) PutReturnStruct(url string, payload []byte, headers map[string]string, target any) error {
	args := m.Called(url, payload, headers, target)
	return args.Error(0)
}

func (m *MockHTTPClient) DeleteReturnNoContent(url string, headers map[string]string) error {
	args := m.Called(url, headers)
	return args.Error(0)
}

func (m *MockHTTPClient) DeleteRetryReturnNoContent(url string, headers map[string]string) error {
	args := m.Called(url, headers)
	return args.Error(0)
}

func (m *MockHTTPClient) Delete(url string, headers map[string]string) error {
	args := m.Called(url, headers)
	return args.Error(0)
}

func (m *MockHTTPClient) DeleteReturnStruct(url string, headers map[string]string, target any) error {
	args := m.Called(url, headers, target)
	return args.Error(0)
}

func (m *MockHTTPClient) DeleteWithPayloadReturnStruct(url string, payload []byte, headers map[string]string, target any) error {
	args := m.Called(url, payload, headers, target)
	return args.Error(0)
}

// NewMockAction creates a minimal Action instance for testing
func NewMockAction() *action.Action {
	params := &action.Param{}
	return action.New(
		"test-action",
		"http://localhost:%s", // Gateway URL template
		params,
	)
}

// MockCommandExecutor is a mock implementation of execsvc.CommandRunner
type MockCommandExecutor struct {
	mock.Mock
}

func (m *MockCommandExecutor) Exec(cmd *exec.Cmd) error {
	args := m.Called(cmd)
	return args.Error(0)
}

func (m *MockCommandExecutor) ExecIgnoreError(cmd *exec.Cmd) {
	m.Called(cmd)
}

func (m *MockCommandExecutor) ExecReturnOutput(cmd *exec.Cmd) (bytes.Buffer, bytes.Buffer, error) {
	args := m.Called(cmd)
	return args.Get(0).(bytes.Buffer), args.Get(1).(bytes.Buffer), args.Error(2)
}

func (m *MockCommandExecutor) ExecFromDir(cmd *exec.Cmd, workDir string) error {
	args := m.Called(cmd, workDir)
	return args.Error(0)
}

// MockRegistrySvc is a mock implementation of registrysvc.RegistryProcessor
type MockRegistrySvc struct {
	mock.Mock
}

func (m *MockRegistrySvc) GetNamespace(version string) string {
	args := m.Called(version)
	return args.String(0)
}

func (m *MockRegistrySvc) ExtractModuleMetadata(modules *models.ProxyModulesByRegistry) {
	m.Called(modules)
}

func (m *MockRegistrySvc) GetModules(installJsonURLs map[string]string, useRemote, verbose bool) (*models.ProxyModulesByRegistry, error) {
	args := m.Called(installJsonURLs, useRemote, verbose)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ProxyModulesByRegistry), args.Error(1)
}

func (m *MockRegistrySvc) GetAuthorizationToken() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

// MockModuleEnv is a mock implementation of moduleenv.ModuleEnvProcessor
type MockModuleEnv struct {
	mock.Mock
}

func (m *MockModuleEnv) VaultEnv(env []string, vaultToken string) []string {
	args := m.Called(env, vaultToken)
	return args.Get(0).([]string)
}

func (m *MockModuleEnv) KeycloakEnv(env []string) []string {
	args := m.Called(env)
	return args.Get(0).([]string)
}

func (m *MockModuleEnv) OkapiEnv(env []string, sidecarName string, port int) []string {
	args := m.Called(env, sidecarName, port)
	return args.Get(0).([]string)
}

func (m *MockModuleEnv) DisabledSystemUserEnv(env []string, moduleName string) []string {
	args := m.Called(env, moduleName)
	return args.Get(0).([]string)
}

func (m *MockModuleEnv) ModuleEnv(env []string, moduleEnv map[string]any) []string {
	args := m.Called(env, moduleEnv)
	return args.Get(0).([]string)
}

func (m *MockModuleEnv) SidecarEnv(env []string, module *models.ProxyModule, privatePort int, moduleURL, sidecarURL string) []string {
	args := m.Called(env, module, privatePort, moduleURL, sidecarURL)
	return args.Get(0).([]string)
}

// MockDockerClient is a mock implementation of dockerclient.DockerClientRunner
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

func (m *MockDockerClient) Close(dockerClient *client.Client) {
	m.Called(dockerClient)
}

func (m *MockDockerClient) PushImage(namespace string, imageName string) error {
	args := m.Called(namespace, imageName)
	return args.Error(0)
}

func (m *MockDockerClient) ForcePullImage(image string) (string, error) {
	args := m.Called(image)
	return args.String(0), args.Error(1)
}

// MockTenantSvc is a mock implementation of tenantsvc.TenantProcessor
type MockTenantSvc struct {
	mock.Mock
}

func (m *MockTenantSvc) GetEntitlementTenantParameters(consortiumName string) (string, error) {
	args := m.Called(consortiumName)
	return args.String(0), args.Error(1)
}

func (m *MockTenantSvc) SetConfigTenantParams(tenantName string) error {
	args := m.Called(tenantName)
	return args.Error(0)
}
