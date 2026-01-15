package gitclient

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/folio-org/eureka-setup/eureka-cli/action"
	"github.com/folio-org/eureka-setup/eureka-cli/constant"
	"github.com/folio-org/eureka-setup/eureka-cli/errors"
	"github.com/folio-org/eureka-setup/eureka-cli/gitrepository"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// GitClientRunner defines the interface for Git client operations
type GitClientRunner interface {
	GitClientRepositoryProvisioner
	GitClientManager
}

// GitClientRepositoryProvisioner defines the interface for Git repository provisioning
type GitClientRepositoryProvisioner interface {
	KongRepository() (*gitrepository.GitRepository, error)
	KeycloakRepository() (*gitrepository.GitRepository, error)
	PlatformCompleteRepository(branch plumbing.ReferenceName) (*gitrepository.GitRepository, error)
}

// GitClientManager defines the interface for Git repository management
type GitClientManager interface {
	Clone(repository *gitrepository.GitRepository) error
	ResetHardPullFromOrigin(repository *gitrepository.GitRepository) error
}

// GitClient provides functionality for Git operations
type GitClient struct {
	Action *action.Action
}

// New creates a new GitClient instance
func New(action *action.Action) *GitClient {
	return &GitClient{Action: action}
}

func (gc *GitClient) KongRepository() (*gitrepository.GitRepository, error) {
	var (
		label = constant.FolioKeycloakLabel
		url   = constant.FolioKeycloakRepositoryURL
		dir   = constant.FolioKeycloakOutputDir
	)
	return gitrepository.New(gc.Action, label, url, dir, constant.FolioKeycloakBranch)
}

func (gc *GitClient) KeycloakRepository() (*gitrepository.GitRepository, error) {
	var (
		label = constant.FolioKongLabel
		url   = constant.FolioKongRepositoryURL
		dir   = constant.FolioKongOutputDir
	)
	return gitrepository.New(gc.Action, label, url, dir, constant.FolioKongBranch)
}

func (gc *GitClient) PlatformCompleteRepository(branch plumbing.ReferenceName) (*gitrepository.GitRepository, error) {
	var (
		label = constant.PlatformCompleteLabel
		url   = constant.PlatformCompleteRepositoryURL
		dir   = constant.PlatformCompleteOutputDir
	)
	return gitrepository.New(gc.Action, label, url, dir, branch)
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

func (rc *GitClient) ResetHardPullFromOrigin(repository *gitrepository.GitRepository) error {
	slog.Info(rc.Action.Name, "text", "Updating repository", "label", repository.Label, "branch", repository.Branch)
	targetRepository, err := git.PlainOpen(repository.Dir)
	if err != nil {
		return err
	}
	if err = targetRepository.Fetch(&git.FetchOptions{
		Force:    true,
		Progress: os.Stdout,
	}); err != nil {
		slog.Warn(rc.Action.Name, "text", "Fetching repository changes", "label", repository.Label, "message", err.Error())
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
	if err = worktree.Reset(&git.ResetOptions{Mode: git.HardReset}); err != nil {
		return err
	}
	if err = rc.printStatus(worktree, "After Clean & Reset"); err != nil {
		return err
	}

	ref, err := targetRepository.Head()
	if err != nil {
		return err
	}
	if err = worktree.Pull(&git.PullOptions{
		RemoteName:    "origin",
		ReferenceName: ref.Name(),
		SingleBranch:  true,
		Progress:      os.Stdout,
	}); err != nil {
		if strings.Contains(err.Error(), "already up-to-date") {
			slog.Info(rc.Action.Name, "text", "Updating repository pull message", "label", repository.Label, "message", err.Error())
			return nil
		}
		return err
	}

	return nil
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
