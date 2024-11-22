package internal

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func GitCloneRepository(commandName string, enableDebug bool, repositoryUrl string, branchName plumbing.ReferenceName, outputDir string, panicIfExists bool) {
	targetRepository, err := git.PlainClone(outputDir, false, &git.CloneOptions{URL: repositoryUrl, ReferenceName: branchName, Progress: os.Stdout})
	if err != nil {
		if panicIfExists {
			slog.Error(commandName, GetFuncName(), "git.PlainClone error")
			panic(err)
		} else {
			slog.Info(commandName, GetFuncName(), fmt.Sprintf("Message: %s", err.Error()))
			return
		}
	}

	ref, err := targetRepository.Head()
	if err != nil {
		slog.Error(commandName, GetFuncName(), "targetRepository.Head error")
		panic(err)
	}

	slog.Info(commandName, GetFuncName(), fmt.Sprintf("Ref: %s", ref))
}

func GitResetHardPullFromOriginRepository(commandName string, enableDebug bool, repositoryUrl string, branchName plumbing.ReferenceName, outputDir string) {
	targetRepository, err := git.PlainOpen(outputDir)
	if err != nil {
		slog.Error(commandName, GetFuncName(), "git.PlainOpen error")
		panic(err)
	}

	ref, err := targetRepository.Head()
	if err != nil {
		slog.Error(commandName, GetFuncName(), "targetRepository.Head error")
		panic(err)
	}

	worktree, err := targetRepository.Worktree()
	if err != nil {
		slog.Error(commandName, GetFuncName(), "targetRepository.Worktree error")
		panic(err)
	}

	if err = worktree.Reset(&git.ResetOptions{Mode: git.HardReset}); err != nil {
		slog.Error(commandName, GetFuncName(), "worktree.Reset error")
		panic(err)
	}

	if err = worktree.Pull(&git.PullOptions{RemoteURL: repositoryUrl, RemoteName: "origin", ReferenceName: ref.Name(), Progress: os.Stdout}); err != nil {
		if strings.Contains(err.Error(), "already up-to-date") {
			slog.Info(commandName, GetFuncName(), fmt.Sprintf("Message: %s", err.Error()))
			return
		}

		slog.Error(commandName, GetFuncName(), "worktree.Pull error")
		panic(err)
	}
}
