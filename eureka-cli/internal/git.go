package internal

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/go-git/go-git/v5"
)

func GitCloneRepository(commandName string, enableDebug bool, panicIfExists bool, repository *Repository) {
	slog.Info(commandName, GetFuncName(), fmt.Sprintf("Cloning repository, url: %s, output directory: %s, branch: %s", repository.RepositoryUrl, repository.OutputDir, repository.BranchName))

	targetRepository, err := git.PlainClone(repository.OutputDir, false, &git.CloneOptions{URL: repository.RepositoryUrl, ReferenceName: repository.BranchName, Progress: os.Stdout})
	if err != nil {
		if panicIfExists {
			slog.Error(commandName, GetFuncName(), "git.PlainClone error")
			panic(err)
		}
		slog.Info(commandName, GetFuncName(), fmt.Sprintf("Cloning repository, url: %s, clone message: %s", repository.RepositoryUrl, err.Error()))

		return
	}

	ref, err := targetRepository.Head()
	if err != nil {
		slog.Error(commandName, GetFuncName(), "targetRepository.Head error")
		panic(err)
	}

	slog.Info(commandName, GetFuncName(), fmt.Sprintf("Ref: %s", ref))
}

func GitResetHardPullFromOriginRepository(commandName string, enableDebug bool, repository *Repository) {
	slog.Info(commandName, GetFuncName(), fmt.Sprintf("Updating repository, url: %s, output directory: %s, branch: %s", repository.RepositoryUrl, repository.OutputDir, repository.BranchName))

	targetRepository, err := git.PlainOpen(repository.OutputDir)
	if err != nil {
		slog.Error(commandName, GetFuncName(), "git.PlainOpen error")
		panic(err)
	}

	if err = targetRepository.Fetch(&git.FetchOptions{Force: true, Progress: os.Stdout}); err != nil {
		slog.Info(commandName, GetFuncName(), fmt.Sprintf("Updating repository, url: %s, fetch message: %s", repository.RepositoryUrl, err.Error()))
	}

	worktree, err := targetRepository.Worktree()
	if err != nil {
		slog.Error(commandName, GetFuncName(), "targetRepository.Worktree error")
		panic(err)
	}

	printStatus(commandName, worktree, "Before Clean & Reset")

	if err = worktree.Clean(&git.CleanOptions{Dir: true}); err != nil {
		slog.Error(commandName, GetFuncName(), "worktree.Clean error")
		panic(err)
	}

	if err = worktree.Reset(&git.ResetOptions{Mode: git.HardReset}); err != nil {
		slog.Error(commandName, GetFuncName(), "worktree.Reset error")
		panic(err)
	}

	printStatus(commandName, worktree, "After Clean & Reset")

	ref, err := targetRepository.Head()
	if err != nil {
		slog.Error(commandName, GetFuncName(), "targetRepository.Head error")
		panic(err)
	}

	if err = worktree.Pull(&git.PullOptions{RemoteName: "origin", ReferenceName: ref.Name(), SingleBranch: true, Progress: os.Stdout}); err != nil {
		if strings.Contains(err.Error(), "already up-to-date") {
			slog.Info(commandName, GetFuncName(), fmt.Sprintf("Updating repository, url: %s, pull message: %s", repository.RepositoryUrl, err.Error()))
			return
		}

		slog.Error(commandName, GetFuncName(), "worktree.Pull error")
		panic(err)
	}
}

func printStatus(commandName string, worktree *git.Worktree, message string) {
	status, err := worktree.Status()
	if err != nil {
		slog.Error(commandName, GetFuncName(), "worktree.Status error")
		panic(err)
	}

	if status != nil && status.String() != "" {
		fmt.Println(message + ":")
		fmt.Println(status.String())
	}
}
