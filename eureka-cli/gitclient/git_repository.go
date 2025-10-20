package gitclient

import (
	"path/filepath"

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/go-git/go-git/v5/plumbing"
)

type GitRepository struct {
	URL    string
	Dir    string
	Branch plumbing.ReferenceName
}

func NewRepository(action *action.Action, url, dir string, branch plumbing.ReferenceName) (*GitRepository, error) {
	homeMiscDir, err := helpers.GetHomeMiscDir(action)
	if err != nil {
		return nil, err
	}

	return &GitRepository{
		URL:    url,
		Dir:    filepath.Join(homeMiscDir, dir),
		Branch: branch,
	}, nil
}
