package repo

import (
	"fmt"
	"time"

	"gopkg.in/src-d/go-billy.v4/memfs"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/storage/memory"

	"github.com/otaviof/chart-streams/pkg/chartstreams/config"
)

type CommitInfo struct {
	Time time.Time
	Hash plumbing.Hash
}

type GitChartRepo struct {
	*git.Repository
	config *config.Config
}

func (r *GitChartRepo) AllCommits() (object.CommitIter, error) {
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

func NewGitChartRepo(config *config.Config) (*GitChartRepo, error) {
	var err error
	storage := memory.NewStorage()
	fs := memfs.New()
	r, err := git.Clone(storage, fs, &git.CloneOptions{
		URL:        config.RepoURL,
		Depth:      config.CloneDepth,
		NoCheckout: true,
	})

	if err != nil {
		return nil, fmt.Errorf("error cloning the repository: %w", err)
	}

	return &GitChartRepo{r, config}, nil
}
