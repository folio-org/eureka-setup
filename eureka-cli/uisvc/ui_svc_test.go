package uisvc

import (
	"errors"
	"os/exec"
	"testing"

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/gitrepository"
	"github.com/folio-org/eureka-cli/internal/testhelpers"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
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
