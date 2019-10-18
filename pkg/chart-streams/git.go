package chartstreams

import (
	"fmt"

	"gopkg.in/src-d/go-billy.v4/memfs"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

type Commit map[string]string

type CommitIter interface {
	ForEach(func(*Commit) error) error
}

type Git struct {
	config *Config
	r      *git.Repository
}

func (g *Git) Clone() error {
	var err error
	storage := memory.NewStorage()
	fs := memfs.New()
	g.r, err = git.Clone(storage, fs, &git.CloneOptions{
		URL:        g.config.RepoURL,
		Depth:      g.config.Depth,
		NoCheckout: true,
	})

	if err != nil {
		return fmt.Errorf("Clone(): error cloning the repository")
	}

	return nil
}

const defaultChartRelativePath = "stable"

func (g *Git) AllCommits() (object.CommitIter, error) {
	ref, err := g.r.Head()
	if err != nil {
		return nil, fmt.Errorf("AllCommits: error finding Head reference: %s", err)
	}

	commitIter, err := g.r.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return nil, fmt.Errorf("AllCommits: error obtaining commitIter: %s", err)
	}

	return commitIter, nil
}

func NewGit(config *Config) *Git {
	return &Git{config: config}
}
