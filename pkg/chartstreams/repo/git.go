package repo

import (
	"fmt"
	"time"

	"github.com/otaviof/chart-streams/pkg/chartstreams/config"

	"gopkg.in/src-d/go-git.v4/plumbing"

	"gopkg.in/src-d/go-billy.v4/memfs"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

type CommitInfo struct {
	Time time.Time
	Hash plumbing.Hash
}

type GitChartRepository struct {
	*git.Repository
	config *config.Config
}

func (r *GitChartRepository) AllCommits() (object.CommitIter, error) {
	ref, err := r.Head()
	if err != nil {
		return nil, fmt.Errorf("AllCommits: error finding Head reference: %s", err)
	}

	commitIter, err := r.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return nil, fmt.Errorf("AllCommits: error obtaining commitIter: %s", err)
	}

	return commitIter, nil
}

func NewGitChartRepository(config *config.Config) (*GitChartRepository, error) {
	var err error
	storage := memory.NewStorage()
	fs := memfs.New()
	r, err := git.Clone(storage, fs, &git.CloneOptions{
		URL:        config.RepoURL,
		Depth:      config.Depth,
		NoCheckout: true,
	})

	if err != nil {
		return nil, fmt.Errorf("error cloning the repository: %w", err)
	}

	return &GitChartRepository{r, config}, nil
}
