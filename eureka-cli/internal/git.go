package internal

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func GitCloneRepository(commandName string, enableDebug bool, repositoryUrl string, branchName plumbing.ReferenceName, outputDir string, panicIfExists bool) {
	targetRepository, err := git.PlainClone(outputDir, false, &git.CloneOptions{URL: repositoryUrl, Progress: os.Stdout, ReferenceName: branchName})
	if err != nil {
		if panicIfExists {
			slog.Error(commandName, "git.PlainClone error", "")
			panic(err)
		} else {
			slog.Warn(commandName, fmt.Sprintf("git.PlainClone warning, Repository already exists %s", outputDir), "")
			return
		}
	}

	ref, err := targetRepository.Head()
	if err != nil {
		slog.Error(commandName, "targetRepository.Head() error", "")
		panic(err)
	}

	slog.Info(commandName, "Last commit hash", ref.Hash())
}
