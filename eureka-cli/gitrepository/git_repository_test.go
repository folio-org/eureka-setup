package gitrepository

import (
	"path/filepath"
	"testing"

	"github.com/folio-org/eureka-cli/action"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/stretchr/testify/assert"
)

// newMockAction creates a minimal Action instance for testing
func newMockAction() *action.Action {
	params := &action.Param{}
	return action.New(
		"test-action",
		"http://localhost:%s",
		params,
	)
}

func TestNew_Success(t *testing.T) {
	// Arrange
	action := newMockAction()
	label := "test-repo"
	url := "https://github.com/test/repo.git"
	dir := "test-dir"
	branch := plumbing.NewBranchReferenceName("main")

	// Act
	repo, err := New(action, label, url, dir, branch)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, repo)
	assert.Equal(t, label, repo.Label)
	assert.Equal(t, url, repo.URL)
	assert.Equal(t, branch, repo.Branch)
	assert.Contains(t, repo.Dir, dir)
}

func TestNew_WithDifferentBranch(t *testing.T) {
	// Arrange
	action := newMockAction()
	label := "feature-repo"
	url := "https://github.com/test/feature.git"
	dir := "feature-dir"
	branch := plumbing.NewBranchReferenceName("feature-branch")

	// Act
	repo, err := New(action, label, url, dir, branch)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, repo)
	assert.Equal(t, label, repo.Label)
	assert.Equal(t, url, repo.URL)
	assert.Equal(t, branch, repo.Branch)
	assert.Contains(t, repo.Dir, dir)
}

func TestNew_DirectoryPathConstruction(t *testing.T) {
	// Arrange
	action := newMockAction()
	label := "path-test-repo"
	url := "https://github.com/test/path.git"
	dir := "subdir/nested"
	branch := plumbing.NewBranchReferenceName("develop")

	// Act
	repo, err := New(action, label, url, dir, branch)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, repo)
	// Verify the directory is properly joined with home directory
	expectedSuffix := filepath.Join("subdir", "nested")
	assert.Contains(t, repo.Dir, expectedSuffix)
}

func TestString_ReturnsLabel(t *testing.T) {
	// Arrange
	action := newMockAction()
	label := "stringify-repo"
	url := "https://github.com/test/stringify.git"
	dir := "stringify-dir"
	branch := plumbing.NewBranchReferenceName("main")

	repo, err := New(action, label, url, dir, branch)
	assert.NoError(t, err)

	// Act
	result := repo.String()

	// Assert
	assert.Equal(t, label, result)
}

func TestNew_WithTagReference(t *testing.T) {
	// Arrange
	action := newMockAction()
	label := "tag-repo"
	url := "https://github.com/test/tag.git"
	dir := "tag-dir"
	branch := plumbing.NewTagReferenceName("v1.0.0")

	// Act
	repo, err := New(action, label, url, dir, branch)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, repo)
	assert.Equal(t, label, repo.Label)
	assert.Equal(t, url, repo.URL)
	assert.Equal(t, branch, repo.Branch)
	assert.True(t, branch.IsTag())
}

func TestNew_VerifyFieldValues(t *testing.T) {
	// Arrange
	action := newMockAction()
	expectedLabel := "verify-repo"
	expectedURL := "https://github.com/folio-org/platform-complete.git"
	expectedDir := "platform-complete"
	expectedBranch := plumbing.NewBranchReferenceName("snapshot")

	// Act
	repo, err := New(action, expectedLabel, expectedURL, expectedDir, expectedBranch)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedLabel, repo.Label)
	assert.Equal(t, expectedURL, repo.URL)
	assert.Equal(t, expectedBranch, repo.Branch)
	assert.NotEmpty(t, repo.Dir)
}

func TestNew_EmptyValues(t *testing.T) {
	// Arrange
	action := newMockAction()
	label := ""
	url := ""
	dir := ""
	branch := plumbing.ReferenceName("")

	// Act
	repo, err := New(action, label, url, dir, branch)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, repo)
	assert.Equal(t, "", repo.Label)
	assert.Equal(t, "", repo.URL)
	assert.Equal(t, branch, repo.Branch)
}

func TestGitRepository_AllFieldsAccessible(t *testing.T) {
	// Arrange
	action := newMockAction()
	label := "access-test"
	url := "https://github.com/test/access.git"
	dir := "access-dir"
	branch := plumbing.NewBranchReferenceName("master")

	// Act
	repo, err := New(action, label, url, dir, branch)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, repo.Label)
	assert.NotEmpty(t, repo.URL)
	assert.NotEmpty(t, repo.Dir)
	assert.NotEqual(t, plumbing.ReferenceName(""), repo.Branch)
}
