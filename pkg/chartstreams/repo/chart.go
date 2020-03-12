package repo

import (
	"fmt"
	"os"
	"time"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"

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
	// making sure all commits are included in iterator, from all branches
	commitIter, err := r.Log(&git.LogOptions{From: ref.Hash(), All: true})
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
		Progress:   os.Stdout,
		NoCheckout: true,
	})
	if err != nil {
		return nil, fmt.Errorf("error cloning the repository: %v", err)
	}

	return &GitChartRepo{r, config}, nil
}
