package gitclient

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/constant"
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

func (gc *GitClient) KeycloakRepository() *GitRepository {
	var (
		url                           = constant.FolioKongRepositoryURL
		dir                           = constant.FolioKongOutputDir
		branch plumbing.ReferenceName = constant.FolioKongBranch
	)
	return NewRepository(gc.Action, url, dir, branch)
}

func (gc *GitClient) KongRepository() *GitRepository {
	var (
		url                           = constant.FolioKeycloakRepositoryURL
		dir                           = constant.FolioKeycloakOutputDir
		branch plumbing.ReferenceName = constant.FolioKeycloakBranch
	)
	return NewRepository(gc.Action, url, dir, branch)
}

func (gc *GitClient) PlatformCompleteRepository(branch plumbing.ReferenceName) *GitRepository {
	return NewRepository(gc.Action, constant.PlatformCompleteRepositoryURL, constant.PlatformCompleteOutputDir, branch)
}

func (rc *GitClient) Clone(panicIfExists bool, repository *GitRepository) {
	targetRepository, err := git.PlainClone(repository.Dir, false, &git.CloneOptions{
		URL:           repository.URL,
		ReferenceName: repository.Branch,
		Progress:      os.Stdout,
	})
	if err != nil {
		if panicIfExists {
			slog.Error(rc.Action.Name, "error", err)
			panic(err)
		}
		slog.Info(rc.Action.Name, "text", fmt.Sprintf("Cloning repository, url: %s, clone message: %s", repository.URL, err.Error()))

		return
	}

	ref, err := targetRepository.Head()
	if err != nil {
		slog.Error(rc.Action.Name, "error", err)
		panic(err)
	}

	slog.Info(rc.Action.Name, "text", fmt.Sprintf("Ref: %s", ref))
}

func (rc *GitClient) ResetHardPullFromOrigin(repository *GitRepository) {
	slog.Info(rc.Action.Name, "text", fmt.Sprintf("Updating repository, url: %s, branch: %s", repository.URL, repository.Branch))

	targetRepository, err := git.PlainOpen(repository.Dir)
	if err != nil {
		slog.Error(rc.Action.Name, "error", err)
		panic(err)
	}

	if err = targetRepository.Fetch(&git.FetchOptions{
		Force:    true,
		Progress: os.Stdout,
	}); err != nil {
		slog.Info(rc.Action.Name, "text", fmt.Sprintf("Updating repository, url: %s, fetch message: %s", repository.URL, err.Error()))
	}

	worktree, err := targetRepository.Worktree()
	if err != nil {
		slog.Error(rc.Action.Name, "error", err)
		panic(err)
	}

	rc.printStatus(worktree, "Before Clean & Reset")

	if err = worktree.Clean(&git.CleanOptions{Dir: true}); err != nil {
		slog.Error(rc.Action.Name, "error", err)
		panic(err)
	}

	if err = worktree.Reset(&git.ResetOptions{Mode: git.HardReset}); err != nil {
		slog.Error(rc.Action.Name, "error", err)
		panic(err)
	}

	rc.printStatus(worktree, "After Clean & Reset")

	ref, err := targetRepository.Head()
	if err != nil {
		slog.Error(rc.Action.Name, "error", err)
		panic(err)
	}

	if err = worktree.Pull(&git.PullOptions{
		RemoteName:    "origin",
		ReferenceName: ref.Name(),
		SingleBranch:  true,
		Progress:      os.Stdout,
	}); err != nil {
		if strings.Contains(err.Error(), "already up-to-date") {
			slog.Info(rc.Action.Name, "text", fmt.Sprintf("Updating repository, url: %s, pull message: %s", repository.URL, err.Error()))
			return
		}

		slog.Error(rc.Action.Name, "error", err)
		panic(err)
	}
}

func (rc *GitClient) printStatus(wt *git.Worktree, message string) {
	status, err := wt.Status()
	if err != nil {
		slog.Error(rc.Action.Name, "error", err)
		panic(err)
	}

	if status != nil && status.String() != "" {
		fmt.Println(message + ":")
		fmt.Println(status.String())
	}
}
