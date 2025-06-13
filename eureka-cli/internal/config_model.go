package internal

import (
	"path/filepath"

	"github.com/go-git/go-git/v5/plumbing"
)

type Repository struct {
	RepositoryUrl string
	OutputDir     string
	BranchName    plumbing.ReferenceName
}

func NewRepository(commandName string, repositoryUrl, outputDir string, branchName plumbing.ReferenceName) *Repository {
	homeMiscDir := GetHomeMiscDir(commandName)

	return &Repository{
		RepositoryUrl: repositoryUrl,
		OutputDir:     filepath.Join(homeMiscDir, outputDir),
		BranchName:    branchName,
	}
}
