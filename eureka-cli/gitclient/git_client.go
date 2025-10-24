package gitclient

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/gitrepository"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

type GitClient struct {
	Action *action.Action
}

func New(action *action.Action) *GitClient {
	return &GitClient{
		Action: action,
	}
}

func (gc *GitClient) KeycloakRepository() (*gitrepository.GitRepository, error) {
	var (
		label                         = constant.FolioKongLabel
		url                           = constant.FolioKongRepositoryURL
		dir                           = constant.FolioKongOutputDir
		branch plumbing.ReferenceName = constant.FolioKongBranch
	)
	return gitrepository.New(gc.Action, label, url, dir, branch)
}

func (gc *GitClient) KongRepository() (*gitrepository.GitRepository, error) {
	var (
		label                         = constant.FolioKeycloakLabel
		url                           = constant.FolioKeycloakRepositoryURL
		dir                           = constant.FolioKeycloakOutputDir
		branch plumbing.ReferenceName = constant.FolioKeycloakBranch
	)
	return gitrepository.New(gc.Action, label, url, dir, branch)
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
		Progress:      os.Stdout,
	})
	if err != nil {
		return fmt.Errorf("cloning %s repository with error %w", repository.Label, err)
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
		slog.Warn(rc.Action.Name, "text", "Updating repository fetch message", "label", repository.Label, "message", err.Error())
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
