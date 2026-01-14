package gitrepository

import (
	"path/filepath"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/j011195/eureka-setup/eureka-cli/action"
	"github.com/j011195/eureka-setup/eureka-cli/helpers"
)

// GitRepository represents a Git repository with its metadata
type GitRepository struct {
	Label  string
	URL    string
	Dir    string
	Branch plumbing.ReferenceName
}

func (gr *GitRepository) String() string {
	return gr.Label
}

func New(action *action.Action, label, url, dir string, branch plumbing.ReferenceName) (*GitRepository, error) {
	homeDir, err := helpers.GetHomeMiscDir()
	if err != nil {
		return nil, err
	}

	finalDir := filepath.Join(homeDir, dir)
	return &GitRepository{Label: label, URL: url, Dir: finalDir, Branch: branch}, nil
}
