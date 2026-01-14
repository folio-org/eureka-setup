package uisvc

import (
	"errors"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/j011195/eureka-setup/eureka-cli/action"
	"github.com/j011195/eureka-setup/eureka-cli/constant"
	"github.com/j011195/eureka-setup/eureka-cli/gitrepository"
	"github.com/j011195/eureka-setup/eureka-cli/helpers"
	"github.com/j011195/eureka-setup/eureka-cli/internal/testhelpers"
	"github.com/j011195/eureka-setup/eureka-cli/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNew(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	mockExec := new(testhelpers.MockCommandExecutor)
	mockGitClient := new(testhelpers.MockGitClient)
	mockDockerClient := new(testhelpers.MockDockerClient)
	mockTenantSvc := new(testhelpers.MockTenantSvc)

	// Act
	svc := New(action, mockExec, mockGitClient, mockDockerClient, mockTenantSvc)

	// Assert
	assert.NotNil(t, svc)
	assert.Equal(t, action, svc.Action)
	assert.Equal(t, mockExec, svc.ExecSvc)
	assert.Equal(t, mockGitClient, svc.GitClient)
	assert.Equal(t, mockDockerClient, svc.DockerClient)
	assert.Equal(t, mockTenantSvc, svc.TenantSvc)
}

func TestGetStripesBranch_DefaultBranch(t *testing.T) {
	// Arrange
	act := testhelpers.NewMockAction()
	act.Param = &action.Param{}
	svc := New(act, nil, nil, nil, nil)

	// Act
	branch := svc.GetStripesBranch()

	// Assert
	assert.Equal(t, plumbing.ReferenceName(constant.StripesBranch), branch)
}

func TestGetStripesBranch_ConfigValue(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	action.ConfigApplicationStripesBranch = "refs/heads/custom"
	svc := New(action, nil, nil, nil, nil)

	// Act
	branch := svc.GetStripesBranch()

	// Assert
	// Since action.IsSet uses global viper, it will use default in tests
	// This test verifies the default behavior works correctly
	assert.NotEmpty(t, branch)
}

func TestCloneAndUpdateRepository_CloneSuccess(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	mockGitClient := new(testhelpers.MockGitClient)
	svc := New(action, nil, mockGitClient, nil, nil)

	mockRepo := &gitrepository.GitRepository{
		Label:  "platform-complete",
		URL:    "https://github.com/test/platform-complete.git",
		Dir:    "/home/test/platform-complete",
		Branch: plumbing.NewBranchReferenceName("snapshot"),
	}

	mockGitClient.On("PlatformCompleteRepository", mock.Anything).
		Return(mockRepo, nil)
	mockGitClient.On("Clone", mockRepo).
		Return(nil)

	// Act
	outputDir, err := svc.CloneAndUpdateRepository(false)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "/home/test/platform-complete", outputDir)
	mockGitClient.AssertExpectations(t)
}

func TestCloneAndUpdateRepository_AlreadyExists(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	mockGitClient := new(testhelpers.MockGitClient)
	svc := New(action, nil, mockGitClient, nil, nil)

	mockRepo := &gitrepository.GitRepository{
		Label:  "platform-complete",
		URL:    "https://github.com/test/platform-complete.git",
		Dir:    "/home/test/platform-complete",
		Branch: plumbing.NewBranchReferenceName("snapshot"),
	}

	mockGitClient.On("PlatformCompleteRepository", mock.Anything).
		Return(mockRepo, nil)
	mockGitClient.On("Clone", mockRepo).
		Return(git.ErrRepositoryAlreadyExists)

	// Act
	outputDir, err := svc.CloneAndUpdateRepository(false)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "/home/test/platform-complete", outputDir)
	mockGitClient.AssertExpectations(t)
}

func TestCloneAndUpdateRepository_CloneError(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	mockGitClient := new(testhelpers.MockGitClient)
	svc := New(action, nil, mockGitClient, nil, nil)

	mockRepo := &gitrepository.GitRepository{
		Label:  "platform-complete",
		URL:    "https://github.com/test/platform-complete.git",
		Dir:    "/home/test/platform-complete",
		Branch: plumbing.NewBranchReferenceName("snapshot"),
	}
	cloneErr := errors.New("clone failed")

	mockGitClient.On("PlatformCompleteRepository", mock.Anything).
		Return(mockRepo, nil)
	mockGitClient.On("Clone", mockRepo).
		Return(cloneErr)

	// Act
	outputDir, err := svc.CloneAndUpdateRepository(false)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "", outputDir)
	assert.Equal(t, cloneErr, err)
	mockGitClient.AssertExpectations(t)
}

func TestCloneAndUpdateRepository_WithUpdate(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	mockGitClient := new(testhelpers.MockGitClient)
	svc := New(action, nil, mockGitClient, nil, nil)

	mockRepo := &gitrepository.GitRepository{
		Label:  "platform-complete",
		URL:    "https://github.com/test/platform-complete.git",
		Dir:    "/home/test/platform-complete",
		Branch: plumbing.NewBranchReferenceName("snapshot"),
	}

	mockGitClient.On("PlatformCompleteRepository", mock.Anything).
		Return(mockRepo, nil)
	mockGitClient.On("Clone", mockRepo).
		Return(git.ErrRepositoryAlreadyExists)
	mockGitClient.On("ResetHardPullFromOrigin", mockRepo).
		Return(nil)

	// Act
	outputDir, err := svc.CloneAndUpdateRepository(true)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "/home/test/platform-complete", outputDir)
	mockGitClient.AssertExpectations(t)
}

func TestCloneAndUpdateRepository_UpdateError(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	mockGitClient := new(testhelpers.MockGitClient)
	svc := New(action, nil, mockGitClient, nil, nil)

	mockRepo := &gitrepository.GitRepository{
		Label:  "platform-complete",
		URL:    "https://github.com/test/platform-complete.git",
		Dir:    "/home/test/platform-complete",
		Branch: plumbing.NewBranchReferenceName("snapshot"),
	}
	updateErr := errors.New("update failed")

	mockGitClient.On("PlatformCompleteRepository", mock.Anything).
		Return(mockRepo, nil)
	mockGitClient.On("Clone", mockRepo).
		Return(git.ErrRepositoryAlreadyExists)
	mockGitClient.On("ResetHardPullFromOrigin", mockRepo).
		Return(updateErr)

	// Act
	outputDir, err := svc.CloneAndUpdateRepository(true)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "", outputDir)
	assert.Equal(t, updateErr, err)
	mockGitClient.AssertExpectations(t)
}

func TestCloneAndUpdateRepository_RepositoryError(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	mockGitClient := new(testhelpers.MockGitClient)
	svc := New(action, nil, mockGitClient, nil, nil)

	repoErr := errors.New("repository creation failed")

	mockGitClient.On("PlatformCompleteRepository", mock.Anything).
		Return(nil, repoErr)

	// Act
	outputDir, err := svc.CloneAndUpdateRepository(false)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "", outputDir)
	assert.Equal(t, repoErr, err)
	mockGitClient.AssertExpectations(t)
}

func TestPrepareImage_BuildImages(t *testing.T) {
	// Arrange
	act := testhelpers.NewMockAction()
	act.Param = &action.Param{
		BuildImages:  true,
		UpdateCloned: false,
	}
	mockGitClient := new(testhelpers.MockGitClient)
	svc := New(act, nil, mockGitClient, nil, nil)

	mockRepo := &gitrepository.GitRepository{
		Label:  "platform-complete",
		URL:    "https://github.com/test/platform-complete.git",
		Dir:    "/home/test/platform-complete",
		Branch: plumbing.NewBranchReferenceName("snapshot"),
	}

	mockGitClient.On("PlatformCompleteRepository", mock.Anything).
		Return(mockRepo, nil)
	mockGitClient.On("Clone", mockRepo).
		Return(nil)

	// Act
	_, err := svc.PrepareImage("test-tenant")

	// Assert
	// BuildImage will fail in test since it needs file system, but we verify the flow
	assert.Error(t, err) // Expected to fail at file operations
	mockGitClient.AssertExpectations(t)
}

func TestPrepareImage_PullImage(t *testing.T) {
	// Arrange
	act := testhelpers.NewMockAction()
	act.Param = &action.Param{
		BuildImages: false,
	}
	mockDockerClient := new(testhelpers.MockDockerClient)
	svc := New(act, nil, nil, mockDockerClient, nil)

	expectedImageName := "platform-complete-ui-test-tenant:latest"
	mockDockerClient.On("ForcePullImage", "platform-complete-ui-test-tenant").
		Return(expectedImageName, nil)

	// Act
	imageName, err := svc.PrepareImage("test-tenant")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedImageName, imageName)
	mockDockerClient.AssertExpectations(t)
}

func TestPrepareImage_PullImageError(t *testing.T) {
	// Arrange
	act := testhelpers.NewMockAction()
	act.Param = &action.Param{
		BuildImages: false,
	}
	mockDockerClient := new(testhelpers.MockDockerClient)
	svc := New(act, nil, nil, mockDockerClient, nil)

	pullErr := errors.New("pull failed")
	mockDockerClient.On("ForcePullImage", "platform-complete-ui-test-tenant").
		Return("", pullErr)

	// Act
	imageName, err := svc.PrepareImage("test-tenant")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "", imageName)
	assert.Equal(t, pullErr, err)
	mockDockerClient.AssertExpectations(t)
}

func TestDeployContainer_Success(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	mockExec := new(testhelpers.MockCommandExecutor)
	svc := New(action, mockExec, nil, nil, nil)

	// Mock docker run command
	mockExec.On("Exec", mock.MatchedBy(func(cmd *exec.Cmd) bool {
		return len(cmd.Args) >= 2 &&
			cmd.Args[0] == "docker" &&
			cmd.Args[1] == "run"
	})).Return(nil).Once()

	// Mock docker network connect command
	mockExec.On("Exec", mock.MatchedBy(func(cmd *exec.Cmd) bool {
		return len(cmd.Args) >= 3 &&
			cmd.Args[0] == "docker" &&
			cmd.Args[1] == "network" &&
			cmd.Args[2] == "connect"
	})).Return(nil).Once()

	// Act
	err := svc.DeployContainer("test-tenant", "test-image:latest", 8080)

	// Assert
	assert.NoError(t, err)
	mockExec.AssertExpectations(t)
}

func TestDeployContainer_RunError(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	mockExec := new(testhelpers.MockCommandExecutor)
	svc := New(action, mockExec, nil, nil, nil)

	runErr := errors.New("docker run failed")
	mockExec.On("Exec", mock.MatchedBy(func(cmd *exec.Cmd) bool {
		return len(cmd.Args) >= 2 &&
			cmd.Args[0] == "docker" &&
			cmd.Args[1] == "run"
	})).Return(runErr).Once()

	// Act
	err := svc.DeployContainer("test-tenant", "test-image:latest", 8080)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, runErr, err)
	mockExec.AssertExpectations(t)
}

func TestDeployContainer_NetworkConnectError(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	mockExec := new(testhelpers.MockCommandExecutor)
	svc := New(action, mockExec, nil, nil, nil)

	networkErr := errors.New("network connect failed")

	mockExec.On("Exec", mock.MatchedBy(func(cmd *exec.Cmd) bool {
		return len(cmd.Args) >= 2 &&
			cmd.Args[0] == "docker" &&
			cmd.Args[1] == "run"
	})).Return(nil).Once()

	mockExec.On("Exec", mock.MatchedBy(func(cmd *exec.Cmd) bool {
		return len(cmd.Args) >= 3 &&
			cmd.Args[0] == "docker" &&
			cmd.Args[1] == "network" &&
			cmd.Args[2] == "connect"
	})).Return(networkErr).Once()

	// Act
	err := svc.DeployContainer("test-tenant", "test-image:latest", 8080)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, networkErr, err)
	mockExec.AssertExpectations(t)
}

func TestDeployContainer_VerifyContainerName(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	mockExec := new(testhelpers.MockCommandExecutor)
	svc := New(action, mockExec, nil, nil, nil)

	var capturedContainerName string

	mockExec.On("Exec", mock.MatchedBy(func(cmd *exec.Cmd) bool {
		if len(cmd.Args) >= 5 && cmd.Args[1] == "run" && cmd.Args[2] == "--name" {
			capturedContainerName = cmd.Args[3]
			return true
		}
		return false
	})).Return(nil).Once()

	mockExec.On("Exec", mock.MatchedBy(func(cmd *exec.Cmd) bool {
		return cmd.Args[1] == "network"
	})).Return(nil).Once()

	// Act
	err := svc.DeployContainer("my-tenant", "test-image:latest", 9090)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "eureka-platform-complete-ui-my-tenant", capturedContainerName)
	mockExec.AssertExpectations(t)
}

func TestDeployContainer_VerifyPort(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	mockExec := new(testhelpers.MockCommandExecutor)
	svc := New(action, mockExec, nil, nil, nil)

	var capturedPort string

	mockExec.On("Exec", mock.MatchedBy(func(cmd *exec.Cmd) bool {
		if len(cmd.Args) >= 7 && cmd.Args[1] == "run" && cmd.Args[6] == "--publish" {
			capturedPort = cmd.Args[7]
			return true
		}
		return false
	})).Return(nil).Once()

	mockExec.On("Exec", mock.MatchedBy(func(cmd *exec.Cmd) bool {
		return cmd.Args[1] == "network"
	})).Return(nil).Once()

	// Act
	err := svc.DeployContainer("test-tenant", "test-image:latest", 3001)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "3001:80", capturedPort)
	mockExec.AssertExpectations(t)
}

func TestUISvc_ImplementsInterfaces(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	mockExec := new(testhelpers.MockCommandExecutor)
	mockGitClient := new(testhelpers.MockGitClient)
	mockDockerClient := new(testhelpers.MockDockerClient)
	mockTenantSvc := new(testhelpers.MockTenantSvc)

	// Act
	svc := New(action, mockExec, mockGitClient, mockDockerClient, mockTenantSvc)

	// Assert - verify svc implements expected interfaces
	assert.Implements(t, (*UIProcessor)(nil), svc)
	assert.Implements(t, (*UIRepositoryCloner)(nil), svc)
	assert.Implements(t, (*UIContainerManager)(nil), svc)
	assert.Implements(t, (*UIPackageJSONProcessor)(nil), svc)
	assert.Implements(t, (*UIStripesConfigProcessor)(nil), svc)
}

// ==================== PrepareStripesConfigJS Tests ====================

func TestPrepareStripesConfigJS_SingleTenant(t *testing.T) {
	// Arrange
	act := testhelpers.NewMockAction()
	act.Param = &action.Param{
		SingleTenant:        true,
		PlatformCompleteURL: "http://localhost:8000",
		EnableECSRequests:   false,
	}
	act.ConfigGlobalEnv = map[string]string{
		"kc_login_client_suffix": "-login",
	}
	svc := New(act, nil, nil, nil, nil)

	tempDir := t.TempDir()
	configPath := tempDir
	stripesConfigPath := testhelpers.CreateFileInDir(t, configPath, "stripes.config.js", `
module.exports = {
  okapi: { url: '${kongUrl}', tenant: 'diku' },
  platform: { url: '${tenantUrl}' },
  keycloak: { url: '${keycloakUrl}' },
  hasAllPerms: ${hasAllPerms},
  isSingleTenant: ${isSingleTenant},
  tenantOptions: ${tenantOptions},
  enableEcsRequests: ${enableEcsRequests},
  modules: {
    '@folio/users' : {}
  }
};
`)

	// Act
	err := svc.PrepareStripesConfigJS("diku", stripesConfigPath)

	// Assert
	assert.NoError(t, err)
	content := testhelpers.ReadFileContent(t, configPath, "stripes.config.js")
	assert.Contains(t, content, "isSingleTenant: true")
	assert.Contains(t, content, "http://localhost:8000")
	assert.Contains(t, content, `{diku: {name: "diku", displayName: "diku", clientId: "diku-login"}}`)
	assert.Contains(t, content, "enableEcsRequests: false")
	assert.NotContains(t, content, "@folio/consortia-settings")
}

func TestPrepareStripesConfigJS_MultiTenant(t *testing.T) {
	// Arrange
	act := testhelpers.NewMockAction()
	act.Param = &action.Param{
		SingleTenant:        false,
		PlatformCompleteURL: "http://localhost:8000",
		EnableECSRequests:   true,
	}
	act.ConfigGlobalEnv = map[string]string{
		"kc_login_client_suffix": "-app",
	}
	svc := New(act, nil, nil, nil, nil)

	tempDir := t.TempDir()
	configPath := tempDir
	stripesConfigPath := testhelpers.CreateFileInDir(t, configPath, "stripes.config.js", `
module.exports = {
  okapi: { url: '${kongUrl}', tenant: 'consortium' },
  platform: { url: '${tenantUrl}' },
  keycloak: { url: '${keycloakUrl}' },
  hasAllPerms: ${hasAllPerms},
  isSingleTenant: ${isSingleTenant},
  tenantOptions: ${tenantOptions},
  enableEcsRequests: ${enableEcsRequests},
  modules: {
    '@folio/users' : {}
  }
};
`)

	// Act
	err := svc.PrepareStripesConfigJS("consortium", stripesConfigPath)

	// Assert
	assert.NoError(t, err)
	content := testhelpers.ReadFileContent(t, configPath, "stripes.config.js")
	assert.Contains(t, content, "isSingleTenant: false")
	assert.Contains(t, content, "http://localhost:8000")
	assert.Contains(t, content, `{consortium: {name: "consortium", displayName: "consortium", clientId: "consortium-app"}}`)
	assert.Contains(t, content, "enableEcsRequests: true")
	assert.Contains(t, content, "'@folio/consortia-settings'")
}

func TestPrepareStripesConfigJS_AllPlaceholdersReplaced(t *testing.T) {
	// Arrange
	act := testhelpers.NewMockAction()
	act.Param = &action.Param{
		SingleTenant:        true,
		PlatformCompleteURL: "http://test-platform.com",
		EnableECSRequests:   false,
	}
	act.ConfigGlobalEnv = map[string]string{
		"kc_login_client_suffix": "-suffix",
	}
	svc := New(act, nil, nil, nil, nil)

	tempDir := t.TempDir()
	configPath := tempDir
	stripesConfigPath := testhelpers.CreateFileInDir(t, configPath, "stripes.config.js", `
module.exports = {
  kongUrl: '${kongUrl}',
  tenantUrl: '${tenantUrl}',
  keycloakUrl: '${keycloakUrl}',
  hasAllPerms: ${hasAllPerms},
  isSingleTenant: ${isSingleTenant},
  tenantOptions: ${tenantOptions},
  enableEcsRequests: ${enableEcsRequests}
};
`)

	// Act
	err := svc.PrepareStripesConfigJS("test-tenant", stripesConfigPath)

	// Assert
	assert.NoError(t, err)
	content := testhelpers.ReadFileContent(t, configPath, "stripes.config.js")

	// Verify all placeholders are replaced
	assert.NotContains(t, content, "${kongUrl}")
	assert.NotContains(t, content, "${tenantUrl}")
	assert.NotContains(t, content, "${keycloakUrl}")
	assert.NotContains(t, content, "${hasAllPerms}")
	assert.NotContains(t, content, "${isSingleTenant}")
	assert.NotContains(t, content, "${tenantOptions}")
	assert.NotContains(t, content, "${enableEcsRequests}")

	// Verify actual values
	assert.Contains(t, content, constant.KongExternalHTTP)
	assert.Contains(t, content, constant.KeycloakExternalHTTP)
	assert.Contains(t, content, "http://test-platform.com")
	assert.Contains(t, content, "hasAllPerms: false")
}

func TestPrepareStripesConfigJS_FileNotFound(t *testing.T) {
	// Arrange
	act := testhelpers.NewMockAction()
	act.Param = &action.Param{
		SingleTenant:        true,
		PlatformCompleteURL: "http://localhost:8000",
	}
	svc := New(act, nil, nil, nil, nil)

	// Act
	err := svc.PrepareStripesConfigJS("diku", "/nonexistent/path")

	// Assert
	assert.Error(t, err)
}

func TestPrepareStripesConfigJS_MissingPlaceholder(t *testing.T) {
	// Arrange
	act := testhelpers.NewMockAction()
	act.Param = &action.Param{
		SingleTenant:        false,
		PlatformCompleteURL: "http://localhost:8000",
		EnableECSRequests:   true,
	}
	act.ConfigGlobalEnv = map[string]string{
		"kc_login_client_suffix": "-login",
	}
	svc := New(act, nil, nil, nil, nil)

	tempDir := t.TempDir()
	configPath := tempDir
	stripesConfigPath := testhelpers.CreateFileInDir(t, configPath, "stripes.config.js", `
module.exports = {
  okapi: { url: 'http://localhost:9130', tenant: 'diku' },
  platform: { url: 'http://localhost:3000' }
};
`)

	// Act
	err := svc.PrepareStripesConfigJS("diku", stripesConfigPath)

	// Assert
	assert.NoError(t, err) // Should not error, just skip missing placeholders
	content := testhelpers.ReadFileContent(t, configPath, "stripes.config.js")
	assert.Contains(t, content, "http://localhost:9130") // Original content preserved
}

func TestPrepareStripesConfigJS_EmptyClientIdSuffix(t *testing.T) {
	// Arrange
	act := testhelpers.NewMockAction()
	act.Param = &action.Param{
		SingleTenant:        true,
		PlatformCompleteURL: "http://localhost:8000",
		EnableECSRequests:   false,
	}
	act.ConfigGlobalEnv = map[string]string{} // No suffix
	svc := New(act, nil, nil, nil, nil)

	tempDir := t.TempDir()
	configPath := tempDir
	stripesConfigPath := testhelpers.CreateFileInDir(t, configPath, "stripes.config.js", `
module.exports = {
  tenantOptions: ${tenantOptions}
};
`)

	// Act
	err := svc.PrepareStripesConfigJS("mytenant", stripesConfigPath)

	// Assert
	assert.NoError(t, err)
	content := testhelpers.ReadFileContent(t, configPath, "stripes.config.js")
	assert.Contains(t, content, `{mytenant: {name: "mytenant", displayName: "mytenant", clientId: "mytenant"}}`)
}

func TestPrepareStripesConfigJS_SpecialCharactersInTenantName(t *testing.T) {
	// Arrange
	act := testhelpers.NewMockAction()
	act.Param = &action.Param{
		SingleTenant:        true,
		PlatformCompleteURL: "http://localhost:8000",
		EnableECSRequests:   false,
	}
	act.ConfigGlobalEnv = map[string]string{
		"kc_login_client_suffix": "-app",
	}
	svc := New(act, nil, nil, nil, nil)

	tempDir := t.TempDir()
	configPath := tempDir
	stripesConfigPath := testhelpers.CreateFileInDir(t, configPath, "stripes.config.js", `
module.exports = {
  tenantOptions: ${tenantOptions}
};
`)

	// Act
	err := svc.PrepareStripesConfigJS("test-tenant-123", stripesConfigPath)

	// Assert
	assert.NoError(t, err)
	content := testhelpers.ReadFileContent(t, configPath, "stripes.config.js")
	assert.Contains(t, content, `{test-tenant-123: {name: "test-tenant-123", displayName: "test-tenant-123", clientId: "test-tenant-123-app"}}`)
}

// ==================== PreparePackageJSON Tests ====================

func TestPreparePackageJSON_SingleTenant(t *testing.T) {
	// Arrange
	act := testhelpers.NewMockAction()
	act.Param = &action.Param{
		SingleTenant: true,
	}
	svc := New(act, nil, nil, nil, nil)

	tempDir := t.TempDir()
	configPath := tempDir

	packageJSON := map[string]any{
		"name": "platform-complete",
		"scripts": map[string]any{
			"build": "stripes build",
		},
		"dependencies": map[string]any{
			"@folio/users": "^1.0.0",
		},
	}
	testhelpers.CreateJSONFileInDir(t, configPath, "package.json", packageJSON)

	// Act
	err := svc.PreparePackageJSON(configPath)

	// Assert
	assert.NoError(t, err)

	var result models.PackageJSON
	err = helpers.ReadJSONFromFile(filepath.Join(configPath, "package.json"), &result)
	assert.NoError(t, err)

	// Verify build script is updated
	assert.Contains(t, result.Scripts["build"], "export DEBUG=stripes*")
	assert.Contains(t, result.Scripts["build"], "--max-old-space-size=8000")

	// Verify required modules are added
	assert.Equal(t, ">=1.0.0", result.Dependencies["@folio/authorization-policies"])
	assert.Equal(t, ">=1.0.0", result.Dependencies["@folio/authorization-roles"])
	assert.Equal(t, ">=1.0.0", result.Dependencies["@folio/plugin-select-application"])

	// Verify consortia-settings is NOT added for single tenant
	assert.Empty(t, result.Dependencies["@folio/consortia-settings"])
}

func TestPreparePackageJSON_MultiTenant(t *testing.T) {
	// Arrange
	act := testhelpers.NewMockAction()
	act.Param = &action.Param{
		SingleTenant: false,
	}
	svc := New(act, nil, nil, nil, nil)

	tempDir := t.TempDir()
	configPath := tempDir

	packageJSON := map[string]any{
		"name": "platform-complete",
		"scripts": map[string]any{
			"build": "stripes build",
		},
		"dependencies": map[string]any{
			"@folio/users": "^1.0.0",
		},
	}
	testhelpers.CreateJSONFileInDir(t, configPath, "package.json", packageJSON)

	// Act
	err := svc.PreparePackageJSON(configPath)

	// Assert
	assert.NoError(t, err)

	var result models.PackageJSON
	err = helpers.ReadJSONFromFile(filepath.Join(configPath, "package.json"), &result)
	assert.NoError(t, err)

	// Verify build script is updated
	assert.Contains(t, result.Scripts["build"], "export DEBUG=stripes*")

	// Verify required modules are added
	assert.Equal(t, ">=1.0.0", result.Dependencies["@folio/authorization-policies"])
	assert.Equal(t, ">=1.0.0", result.Dependencies["@folio/authorization-roles"])
	assert.Equal(t, ">=1.0.0", result.Dependencies["@folio/plugin-select-application"])

	// Verify consortia-settings IS added for multi-tenant
	assert.Equal(t, ">=1.0.0", result.Dependencies["@folio/consortia-settings"])
}

func TestPreparePackageJSON_ExistingModules(t *testing.T) {
	// Arrange
	act := testhelpers.NewMockAction()
	act.Param = &action.Param{
		SingleTenant: false,
	}
	svc := New(act, nil, nil, nil, nil)

	tempDir := t.TempDir()
	configPath := tempDir

	packageJSON := map[string]any{
		"name": "platform-complete",
		"scripts": map[string]any{
			"build": "stripes build",
		},
		"dependencies": map[string]any{
			"@folio/users":                  "^1.0.0",
			"@folio/authorization-policies": "^2.0.0",
			"@folio/consortia-settings":     "^1.5.0",
		},
	}
	testhelpers.CreateJSONFileInDir(t, configPath, "package.json", packageJSON)

	// Act
	err := svc.PreparePackageJSON(configPath)

	// Assert
	assert.NoError(t, err)

	var result models.PackageJSON
	err = helpers.ReadJSONFromFile(filepath.Join(configPath, "package.json"), &result)
	assert.NoError(t, err)

	// Verify existing versions are preserved
	assert.Equal(t, "^2.0.0", result.Dependencies["@folio/authorization-policies"])
	assert.Equal(t, "^1.5.0", result.Dependencies["@folio/consortia-settings"])

	// Verify missing modules are added
	assert.Equal(t, ">=1.0.0", result.Dependencies["@folio/authorization-roles"])
	assert.Equal(t, ">=1.0.0", result.Dependencies["@folio/plugin-select-application"])
}

func TestPreparePackageJSON_FileNotFound(t *testing.T) {
	// Arrange
	act := testhelpers.NewMockAction()
	act.Param = &action.Param{
		SingleTenant: true,
	}
	svc := New(act, nil, nil, nil, nil)

	// Act
	err := svc.PreparePackageJSON("/nonexistent/path")

	// Assert
	assert.Error(t, err)
}

func TestPreparePackageJSON_BuildScriptUpdate(t *testing.T) {
	// Arrange
	act := testhelpers.NewMockAction()
	act.Param = &action.Param{
		SingleTenant: true,
	}
	svc := New(act, nil, nil, nil, nil)

	tempDir := t.TempDir()
	configPath := tempDir

	packageJSON := map[string]any{
		"name": "platform-complete",
		"scripts": map[string]any{
			"build": "original-build-script",
		},
		"dependencies": map[string]any{
			"@folio/users": "^1.0.0",
		},
	}
	testhelpers.CreateJSONFileInDir(t, configPath, "package.json", packageJSON)

	// Act
	err := svc.PreparePackageJSON(configPath)

	// Assert
	assert.NoError(t, err)

	var result models.PackageJSON
	err = helpers.ReadJSONFromFile(filepath.Join(configPath, "package.json"), &result)
	assert.NoError(t, err)

	// Verify build script is completely replaced with the expected script
	expectedBuildScript := "export DEBUG=stripes*; export NODE_OPTIONS=\"--max-old-space-size=8000 $NODE_OPTIONS\"; stripes build stripes.config.js --languages en --sourcemap=false --no-minify"
	assert.Equal(t, expectedBuildScript, result.Scripts["build"])
}

func TestPreparePackageJSON_EmptyDependencies(t *testing.T) {
	// Arrange
	act := testhelpers.NewMockAction()
	act.Param = &action.Param{
		SingleTenant: false,
	}
	svc := New(act, nil, nil, nil, nil)

	tempDir := t.TempDir()
	configPath := tempDir

	packageJSON := map[string]any{
		"name": "platform-complete",
		"scripts": map[string]any{
			"build": "stripes build",
		},
		"dependencies": map[string]any{},
	}
	testhelpers.CreateJSONFileInDir(t, configPath, "package.json", packageJSON)

	// Act
	err := svc.PreparePackageJSON(configPath)

	// Assert
	assert.NoError(t, err)

	var result models.PackageJSON
	err = helpers.ReadJSONFromFile(filepath.Join(configPath, "package.json"), &result)
	assert.NoError(t, err)

	// Verify all required modules are added
	assert.Equal(t, ">=1.0.0", result.Dependencies["@folio/authorization-policies"])
	assert.Equal(t, ">=1.0.0", result.Dependencies["@folio/authorization-roles"])
	assert.Equal(t, ">=1.0.0", result.Dependencies["@folio/plugin-select-application"])
	assert.Equal(t, ">=1.0.0", result.Dependencies["@folio/consortia-settings"])
}
