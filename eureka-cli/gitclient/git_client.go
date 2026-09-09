package gitclient

import (
	stderrors "errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/folio-org/eureka-setup/eureka-cli/action"
	"github.com/folio-org/eureka-setup/eureka-cli/constant"
	"github.com/folio-org/eureka-setup/eureka-cli/errors"
	"github.com/folio-org/eureka-setup/eureka-cli/gitrepository"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
)

// GitClientRunner defines the interface for Git client operations
type GitClientRunner interface {
	GitClientRepositoryProvisioner
	GitClientManager
}

// GitClientRepositoryProvisioner defines the interface for Git repository provisioning
type GitClientRepositoryProvisioner interface {
	PlatformLspRepository(url string, branch plumbing.ReferenceName) (*gitrepository.GitRepository, error)
}

// GitClientManager defines the interface for Git repository management
type GitClientManager interface {
	EnsureCheckout(repository *gitrepository.GitRepository, update bool) error
	Clone(repository *gitrepository.GitRepository) error
}

// GitClient provides functionality for Git operations
type GitClient struct {
	Action *action.Action
}

// New creates a new GitClient instance
func New(action *action.Action) *GitClient {
	return &GitClient{Action: action}
}

func (gc *GitClient) PlatformLspRepository(url string, branch plumbing.ReferenceName) (*gitrepository.GitRepository, error) {
	return gitrepository.New(gc.Action, constant.PlatformLspLabel, url, constant.PlatformLspOutputDir, branch)
}

// EnsureCheckout makes sure repository.Dir holds a clone of repository.URL: it clones when there is no checkout, and when the
// checkout was cloned from another repository it replaces the checkout if update is set and fails otherwise.
// Without update, whatever is checked out (branches, commits, edits) is left alone and built as it is; with update, an existing
// checkout is brought to the configured branch as it is on origin.
func (rc *GitClient) EnsureCheckout(repository *gitrepository.GitRepository, update bool) error {
	targetRepository, err := git.PlainOpen(repository.Dir)
	if stderrors.Is(err, git.ErrRepositoryNotExists) {
		return rc.Clone(repository)
	}
	if err != nil {
		return err
	}

	originURL, err := originURL(targetRepository)
	if err != nil {
		return err
	}
	if !sameRemoteURL(originURL, repository.URL) {
		if !update {
			return errors.CheckoutOriginMismatch(repository.Label, repository.Dir, repository.URL, originURL)
		}
		slog.Warn(rc.Action.Name, "text", "Replacing checkout cloned from another repository", "label", repository.Label, "dir", repository.Dir, "origin", originURL, "configured", repository.URL)
		if err := os.RemoveAll(repository.Dir); err != nil {
			return err
		}
		return rc.Clone(repository)
	}
	if update {
		return rc.ResetHardPullFromOrigin(repository)
	}

	head, err := targetRepository.Head()
	if err != nil {
		return err
	}
	if local, _ := branchReferences(repository.Branch); head.Name() != local {
		slog.Warn(rc.Action.Name, "text", "Checkout is not on the configured branch and is built as it is, pass -u to switch", "label", repository.Label, "checkout", describeHead(head), "configured", local.Short())
	}

	return nil
}

func originURL(targetRepository *git.Repository) (string, error) {
	remote, err := targetRepository.Remote(git.DefaultRemoteName)
	if err != nil {
		return "", err
	}
	if urls := remote.Config().URLs; len(urls) > 0 {
		return urls[0], nil
	}

	return "", nil
}

func sameRemoteURL(a string, b string) bool {
	return strings.TrimSuffix(a, "/") == strings.TrimSuffix(b, "/")
}

// branchReferences returns the local and the origin tracking reference of a configured branch, given as a short or a full name
func branchReferences(branch plumbing.ReferenceName) (local plumbing.ReferenceName, remote plumbing.ReferenceName) {
	short := branch.Short()
	return plumbing.NewBranchReferenceName(short), plumbing.NewRemoteReferenceName(git.DefaultRemoteName, short)
}

func describeHead(head *plumbing.Reference) string {
	if head.Name().IsBranch() {
		return head.Name().Short()
	}

	return "detached at " + head.Hash().String()[:7]
}

func (rc *GitClient) Clone(repository *gitrepository.GitRepository) error {
	targetRepository, err := git.PlainClone(repository.Dir, false, &git.CloneOptions{
		URL:           repository.URL,
		ReferenceName: repository.Branch,
		SingleBranch:  true,
		Progress:      os.Stdout,
	})
	if err != nil {
		return errors.CloneFailed(repository.Label, err)
	}

	ref, err := targetRepository.Head()
	if err != nil {
		return err
	}
	slog.Info(rc.Action.Name, "text", "Ref", "ref", ref)

	return nil
}

// ResetHardPullFromOrigin brings the checkout to the configured branch as it is on origin: fetches that branch, points the
// local branch at the fetched commit, removes untracked files and checks the branch out. Uncommitted changes are discarded;
// other local branches are kept. A failed fetch is an error and leaves the checkout untouched.
func (rc *GitClient) ResetHardPullFromOrigin(repository *gitrepository.GitRepository) error {
	local, remote := branchReferences(repository.Branch)
	slog.Info(rc.Action.Name, "text", "Updating repository", "label", repository.Label, "branch", local.Short())
	targetRepository, err := git.PlainOpen(repository.Dir)
	if err != nil {
		return err
	}
	// A single-branch clone only fetches the branch it was cloned with, so the configured branch is named explicitly
	if err = targetRepository.Fetch(&git.FetchOptions{
		RefSpecs: []config.RefSpec{config.RefSpec(fmt.Sprintf("+%s:%s", local, remote))},
		Force:    true,
		Progress: os.Stdout,
	}); err != nil && !stderrors.Is(err, git.NoErrAlreadyUpToDate) {
		if stderrors.Is(err, git.NoMatchingRefSpecError{}) {
			return errors.RemoteBranchMissing(repository.Label, local.Short(), err)
		}
		return errors.FetchFailed(repository.Label, err)
	}
	remoteRef, err := targetRepository.Reference(remote, true)
	if err != nil {
		return err
	}

	worktree, err := targetRepository.Worktree()
	if err != nil {
		return err
	}
	if err = rc.printStatus(worktree, "Before Clean & Reset"); err != nil {
		return err
	}
	if err = worktree.Clean(&git.CleanOptions{Dir: true}); err != nil {
		return err
	}
	if err = targetRepository.Storer.SetReference(plumbing.NewHashReference(local, remoteRef.Hash())); err != nil {
		return err
	}
	// A forced checkout hard-resets the worktree to the branch tip
	if err = worktree.Checkout(&git.CheckoutOptions{Branch: local, Force: true}); err != nil {
		return err
	}

	return rc.printStatus(worktree, "After Clean & Reset")
}

func (rc *GitClient) printStatus(wt *git.Worktree, message string) error {
	status, err := wt.Status()
	if err != nil {
		return err
	}
	if status != nil && status.String() != "" {
		fmt.Println(message + ":")
		fmt.Println(status.String())
	}

	return nil
}
