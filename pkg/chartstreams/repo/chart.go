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

// CommitInfo holds together time and commit hash.
type CommitInfo struct {
	Time time.Time
	Hash plumbing.Hash
}

// GitChartRepo represents a chart repository having Git as backend.
type GitChartRepo struct {
	*git.Repository
	config *config.Config
}

// AllCommits returns a iterator with all commits available. It can return error on reading from the
// tree, or logs.
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

// NewGitChartRepo instantiate a Git based chart repository, initializing Git repo clone first.
func NewGitChartRepo(config *config.Config) (*GitChartRepo, error) {
	storage := memory.NewStorage()
	fs := memfs.New()
	r, err := git.Clone(storage, fs, &git.CloneOptions{
		URL:        config.RepoURL,
		Depth:      config.CloneDepth,
		NoCheckout: true,
	})
	if err != nil {
		return nil, fmt.Errorf("error cloning the repository: %v", err)
	}

	return &GitChartRepo{r, config}, nil
}
