package gitclient

import (
	"path/filepath"

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/go-git/go-git/v5/plumbing"
)

type GitRepository struct {
	Label  string
	URL    string
	Dir    string
	Branch plumbing.ReferenceName
}

func (gr *GitRepository) String() string {
	return gr.Label
}

func NewRepository(action *action.Action, label, url, dir string, branch plumbing.ReferenceName) (*GitRepository, error) {
	homeMiscDir, err := helpers.GetHomeMiscDir(action)
	if err != nil {
		return nil, err
	}

	return &GitRepository{
		Label:  label,
		URL:    url,
		Dir:    filepath.Join(homeMiscDir, dir),
		Branch: branch,
	}, nil
}
