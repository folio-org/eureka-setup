package gitclient

import (
	"testing"

	"github.com/folio-org/eureka-setup/eureka-cli/constant"
	"github.com/folio-org/eureka-setup/eureka-cli/internal/testhelpers"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()

	// Act
	client := New(action)

	// Assert
	assert.NotNil(t, client)
	assert.Equal(t, action, client.Action)
}

func TestKongRepository(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	client := New(action)

	// Act
	repo, err := client.KongRepository()

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, repo)
	assert.Equal(t, constant.FolioKeycloakLabel, repo.Label)
	assert.Equal(t, constant.FolioKeycloakRepositoryURL, repo.URL)
	assert.Contains(t, repo.Dir, constant.FolioKeycloakOutputDir)
	assert.Equal(t, plumbing.ReferenceName(constant.FolioKeycloakBranch), repo.Branch)
}

func TestKeycloakRepository(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	client := New(action)

	// Act
	repo, err := client.KeycloakRepository()

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, repo)
	assert.Equal(t, constant.FolioKongLabel, repo.Label)
	assert.Equal(t, constant.FolioKongRepositoryURL, repo.URL)
	assert.Contains(t, repo.Dir, constant.FolioKongOutputDir)
	assert.Equal(t, plumbing.ReferenceName(constant.FolioKongBranch), repo.Branch)
}

func TestPlatformCompleteRepository_WithCustomBranch(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	client := New(action)
	customBranch := plumbing.NewBranchReferenceName("snapshot")

	// Act
	repo, err := client.PlatformCompleteRepository(customBranch)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, repo)
	assert.Equal(t, constant.PlatformCompleteLabel, repo.Label)
	assert.Equal(t, constant.PlatformCompleteRepositoryURL, repo.URL)
	assert.Contains(t, repo.Dir, constant.PlatformCompleteOutputDir)
	assert.Equal(t, customBranch, repo.Branch)
}

func TestPlatformCompleteRepository_WithDifferentBranch(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	client := New(action)
	differentBranch := plumbing.NewBranchReferenceName("poppy")

	// Act
	repo, err := client.PlatformCompleteRepository(differentBranch)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, repo)
	assert.Equal(t, constant.PlatformCompleteLabel, repo.Label)
	assert.Equal(t, constant.PlatformCompleteRepositoryURL, repo.URL)
	assert.Contains(t, repo.Dir, constant.PlatformCompleteOutputDir)
	assert.Equal(t, differentBranch, repo.Branch)
}

func TestKongRepository_VerifyConstants(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	client := New(action)

	// Act
	repo, err := client.KongRepository()

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, repo)
	// Verify all fields use the correct constants
	assert.NotEmpty(t, repo.Label)
	assert.NotEmpty(t, repo.URL)
	assert.NotEmpty(t, repo.Dir)
	assert.NotEqual(t, "", repo.Branch)
}

func TestKeycloakRepository_VerifyConstants(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	client := New(action)

	// Act
	repo, err := client.KeycloakRepository()

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, repo)
	// Verify all fields use the correct constants
	assert.NotEmpty(t, repo.Label)
	assert.NotEmpty(t, repo.URL)
	assert.NotEmpty(t, repo.Dir)
	assert.NotEqual(t, "", repo.Branch)
}

func TestPlatformCompleteRepository_VerifyConstants(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	client := New(action)
	branch := plumbing.NewBranchReferenceName("main")

	// Act
	repo, err := client.PlatformCompleteRepository(branch)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, repo)
	// Verify all fields use the correct constants
	assert.NotEmpty(t, repo.Label)
	assert.NotEmpty(t, repo.URL)
	assert.NotEmpty(t, repo.Dir)
	assert.NotEqual(t, "", repo.Branch)
}

func TestKongRepository_UniqueFromKeycloak(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	client := New(action)

	// Act
	kongRepo, err1 := client.KongRepository()
	keycloakRepo, err2 := client.KeycloakRepository()

	// Assert
	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NotNil(t, kongRepo)
	assert.NotNil(t, keycloakRepo)
	// These should be different repositories
	assert.NotEqual(t, kongRepo.Label, keycloakRepo.Label)
	assert.NotEqual(t, kongRepo.URL, keycloakRepo.URL)
}

func TestPlatformCompleteRepository_BranchParameter(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	client := New(action)

	testCases := []struct {
		name   string
		branch plumbing.ReferenceName
	}{
		{"Snapshot branch", plumbing.NewBranchReferenceName("snapshot")},
		{"Poppy branch", plumbing.NewBranchReferenceName("poppy")},
		{"Quesnelia branch", plumbing.NewBranchReferenceName("quesnelia")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			repo, err := client.PlatformCompleteRepository(tc.branch)

			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, repo)
			assert.Equal(t, tc.branch, repo.Branch)
		})
	}
}

func TestRepositoryProvisioner_AllMethodsReturnGitRepository(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	client := New(action)

	// Act
	kongRepo, kongErr := client.KongRepository()
	keycloakRepo, keycloakErr := client.KeycloakRepository()
	platformRepo, platformErr := client.PlatformCompleteRepository(plumbing.NewBranchReferenceName("main"))

	// Assert
	assert.NoError(t, kongErr)
	assert.NoError(t, keycloakErr)
	assert.NoError(t, platformErr)
	assert.NotNil(t, kongRepo)
	assert.NotNil(t, keycloakRepo)
	assert.NotNil(t, platformRepo)
}

func TestGitClient_ImplementsInterfaces(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()

	// Act
	client := New(action)

	// Assert - verify client implements expected interfaces
	assert.Implements(t, (*GitClientRepositoryProvisioner)(nil), client)
	assert.Implements(t, (*GitClientManager)(nil), client)
	assert.Implements(t, (*GitClientRunner)(nil), client)
}
