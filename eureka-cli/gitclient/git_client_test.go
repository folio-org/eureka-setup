package gitclient

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/folio-org/eureka-setup/eureka-cli/constant"
	"github.com/folio-org/eureka-setup/eureka-cli/errors"
	"github.com/folio-org/eureka-setup/eureka-cli/gitrepository"
	"github.com/folio-org/eureka-setup/eureka-cli/internal/testhelpers"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	repo, err := client.PlatformLspRepository(constant.PlatformLspRepositoryURL, customBranch)

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
	repo, err := client.PlatformLspRepository(constant.PlatformLspRepositoryURL, differentBranch)

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
	repo, err := client.PlatformLspRepository(constant.PlatformLspRepositoryURL, branch)

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
		url    string
		branch plumbing.ReferenceName
	}{
		{"Snapshot branch", constant.PlatformLspRepositoryURL, plumbing.NewBranchReferenceName("snapshot")},
		{"Poppy branch", constant.PlatformLspRepositoryURL, plumbing.NewBranchReferenceName("poppy")},
		{"Quesnelia branch", constant.PlatformLspRepositoryURL, plumbing.NewBranchReferenceName("quesnelia")},
		{"Fork", "https://example.org/vendor/platform-fork.git", plumbing.NewBranchReferenceName("stable")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			repo, err := client.PlatformLspRepository(tc.url, tc.branch)

			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, repo)
			assert.Equal(t, tc.url, repo.URL)
			assert.Equal(t, tc.branch, repo.Branch)
			assert.Equal(t, constant.PlatformLspLabel, repo.Label)
			assert.Contains(t, repo.Dir, constant.PlatformLspOutputDir)
		})
	}
}

func TestRepositoryProvisioner_AllMethodsReturnGitRepository(t *testing.T) {
	testhelpers.SetTempHome(t)

	// Arrange
	action := testhelpers.NewMockAction()
	client := New(action)

	// Act
	platformRepo, platformErr := client.PlatformLspRepository(constant.PlatformLspRepositoryURL, plumbing.NewBranchReferenceName("main"))

	// Assert
	assert.NoError(t, platformErr)
	assert.NotNil(t, platformRepo)
}

// newSourceRepository creates a local repository with one commit that EnsureCheckout can clone from
func newSourceRepository(t *testing.T) (url string, branch plumbing.ReferenceName) {
	t.Helper()
	dir := t.TempDir()
	repo, err := git.PlainInit(dir, false)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte(dir), 0644))
	worktree, err := repo.Worktree()
	require.NoError(t, err)
	_, err = worktree.Add("README.md")
	require.NoError(t, err)
	_, err = worktree.Commit("initial", &git.CommitOptions{Author: &object.Signature{Name: "test", Email: "test@example.org", When: time.Now()}})
	require.NoError(t, err)
	head, err := repo.Head()
	require.NoError(t, err)

	return dir, head.Name()
}

func originURL(t *testing.T, dir string) string {
	t.Helper()
	repo, err := git.PlainOpen(dir)
	require.NoError(t, err)
	remote, err := repo.Remote(git.DefaultRemoteName)
	require.NoError(t, err)

	return remote.Config().URLs[0]
}

func TestEnsureCheckout(t *testing.T) {
	client := New(testhelpers.NewMockAction())
	sourceA, branchA := newSourceRepository(t)
	sourceB, branchB := newSourceRepository(t)
	localFile := "local-work.txt"

	checkout := func(t *testing.T, url string, branch plumbing.ReferenceName) *gitrepository.GitRepository {
		t.Helper()
		repository := &gitrepository.GitRepository{Label: constant.PlatformLspLabel, URL: url, Dir: filepath.Join(t.TempDir(), constant.PlatformLspOutputDir), Branch: branch}
		require.NoError(t, client.EnsureCheckout(repository, false))
		require.NoError(t, os.WriteFile(filepath.Join(repository.Dir, localFile), []byte("keep"), 0644))

		return repository
	}

	t.Run("Clones when there is no checkout", func(t *testing.T) {
		repository := checkout(t, sourceA, branchA)

		assert.Equal(t, sourceA, originURL(t, repository.Dir))
		assert.FileExists(t, filepath.Join(repository.Dir, "README.md"))
	})

	t.Run("Keeps a checkout of the configured repository", func(t *testing.T) {
		for _, url := range []string{sourceA, sourceA + "/"} {
			repository := checkout(t, sourceA, branchA)
			repository.URL = url

			assert.NoError(t, client.EnsureCheckout(repository, false))
			assert.FileExists(t, filepath.Join(repository.Dir, localFile))
		}
	})

	t.Run("Fails on a checkout of another repository", func(t *testing.T) {
		repository := checkout(t, sourceA, branchA)
		repository.URL, repository.Branch = sourceB, branchB

		err := client.EnsureCheckout(repository, false)

		assert.ErrorIs(t, err, errors.ErrInvalidInput)
		assert.Contains(t, err.Error(), sourceA)
		assert.Contains(t, err.Error(), sourceB)
		assert.Contains(t, err.Error(), "-u")
		assert.Equal(t, sourceA, originURL(t, repository.Dir))
		assert.FileExists(t, filepath.Join(repository.Dir, localFile))
	})

	t.Run("Replaces a checkout of another repository", func(t *testing.T) {
		repository := checkout(t, sourceA, branchA)
		repository.URL, repository.Branch = sourceB, branchB

		assert.NoError(t, client.EnsureCheckout(repository, true))
		assert.Equal(t, sourceB, originURL(t, repository.Dir))
		assert.NoFileExists(t, filepath.Join(repository.Dir, localFile))
	})

	t.Run("Fails on a checkout without origin", func(t *testing.T) {
		dir := t.TempDir()
		_, err := git.PlainInit(dir, false)
		require.NoError(t, err)

		err = client.EnsureCheckout(&gitrepository.GitRepository{Label: constant.PlatformLspLabel, URL: sourceA, Dir: dir, Branch: branchA}, true)

		assert.ErrorIs(t, err, git.ErrRemoteNotFound)
	})
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
