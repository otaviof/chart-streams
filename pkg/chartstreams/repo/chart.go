package repo

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"

	log "github.com/sirupsen/logrus"

	"github.com/otaviof/chart-streams/pkg/chartstreams/config"
)

// CommitInfo holds together time and commit hash.
type CommitInfo struct {
	Time   *time.Time    // commit time
	Hash   plumbing.Hash // commit hash
	Digest string        // data digest
}

// GitChartRepo represents a chart repository having Git as backend.
type GitChartRepo struct {
	*git.Repository                          // embedded git.Repository
	config          *config.Config           // application configuration
	Revisions       map[plumbing.Hash]string // mapping revision commit to name
}

// filterPaths check if informed path has relative-dir as prefix.
func (g *GitChartRepo) filterPaths(entityPath string) bool {
	if g.config.RelativeDir == "/" {
		return true
	}
	return strings.HasPrefix(entityPath, g.config.RelativeDir)
}

// AllCommits returns a iterator with all commits available. It can return error on reading from the
// tree, or logs.
func (g *GitChartRepo) AllCommits() (object.CommitIter, error) {
	ref, err := g.Head()
	if err != nil {
		return nil, fmt.Errorf("AllCommits: error finding Head reference: %s", err)
	}
	// making sure all commits are included in iterator, from all branches
	logOptions := &git.LogOptions{
		From:       ref.Hash(),
		All:        true,
		PathFilter: g.filterPaths,
	}
	commitIter, err := g.Log(logOptions)
	if err != nil {
		return nil, fmt.Errorf("AllCommits: error obtaining commitIter: %s", err)
	}
	return commitIter, nil
}

// NewGitChartRepo instantiate a Git based chart repository, initializing Git repo clone first.
func NewGitChartRepo(config *config.Config) (*GitChartRepo, error) {
	storage := memory.NewStorage()
	fs := memfs.New()
	options := &git.CloneOptions{
		URL:        config.RepoURL,
		Depth:      config.CloneDepth,
		NoCheckout: true,
		Progress:   os.Stdout,
		Tags:       git.NoTags,
	}
	r, err := git.Clone(storage, fs, options)
	if err != nil {
		return nil, fmt.Errorf("error cloning the repository: '%v'", err)
	}

	rs, err := r.References()
	if err != nil {
		return nil, fmt.Errorf("error on inspecting references: '%v'", err)
	}
	defer rs.Close()

	revisions := make(map[plumbing.Hash]string)
	_ = rs.ForEach(func(ref *plumbing.Reference) error {
		if !ref.Hash().IsZero() {
			revisions[ref.Hash()] = strings.TrimPrefix(ref.Name().Short(), "origin/")
			log.Debugf("Revision: hash='%s', name='%s'", ref.Hash(), ref.Name())
		}
		return nil
	})

	return &GitChartRepo{
		Repository: r,
		config:     config,
		Revisions:  revisions,
	}, nil
}
