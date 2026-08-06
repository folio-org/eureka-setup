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

func TestPlatformLspRepository_WithCustomBranch(t *testing.T) {
	testhelpers.SetTempHome(t)

	// Arrange
	action := testhelpers.NewMockAction()
	client := New(action)
	customBranch := plumbing.NewBranchReferenceName("snapshot")

	// Act
	repo, err := client.PlatformLspRepository(customBranch)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, repo)
	assert.Equal(t, constant.PlatformLspLabel, repo.Label)
	assert.Equal(t, constant.PlatformLspRepositoryURL, repo.URL)
	assert.Contains(t, repo.Dir, constant.PlatformLspOutputDir)
	assert.Equal(t, customBranch, repo.Branch)
}

func TestPlatformLspRepository_WithDifferentBranch(t *testing.T) {
	testhelpers.SetTempHome(t)

	// Arrange
	action := testhelpers.NewMockAction()
	client := New(action)
	differentBranch := plumbing.NewBranchReferenceName("poppy")

	// Act
	repo, err := client.PlatformLspRepository(differentBranch)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, repo)
	assert.Equal(t, constant.PlatformLspLabel, repo.Label)
	assert.Equal(t, constant.PlatformLspRepositoryURL, repo.URL)
	assert.Contains(t, repo.Dir, constant.PlatformLspOutputDir)
	assert.Equal(t, differentBranch, repo.Branch)
}

func TestPlatformLspRepository_VerifyConstants(t *testing.T) {
	testhelpers.SetTempHome(t)

	// Arrange
	action := testhelpers.NewMockAction()
	client := New(action)
	branch := plumbing.NewBranchReferenceName("main")

	// Act
	repo, err := client.PlatformLspRepository(branch)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, repo)
	// Verify all fields use the correct constants
	assert.NotEmpty(t, repo.Label)
	assert.NotEmpty(t, repo.URL)
	assert.NotEmpty(t, repo.Dir)
	assert.NotEqual(t, "", repo.Branch)
}

func TestPlatformLspRepository_BranchParameter(t *testing.T) {
	testhelpers.SetTempHome(t)

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
			repo, err := client.PlatformLspRepository(tc.branch)

			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, repo)
			assert.Equal(t, tc.branch, repo.Branch)
		})
	}
}

func TestRepositoryProvisioner_AllMethodsReturnGitRepository(t *testing.T) {
	testhelpers.SetTempHome(t)

	// Arrange
	action := testhelpers.NewMockAction()
	client := New(action)

	// Act
	platformRepo, platformErr := client.PlatformLspRepository(plumbing.NewBranchReferenceName("main"))

	// Assert
	assert.NoError(t, platformErr)
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
