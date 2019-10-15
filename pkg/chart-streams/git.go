package chartstreams

import (
	"os"

	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

type Git struct {
	config *Config
	r      *git.Repository
}

func (g *Git) Clone() error {
	var err error
	storage := memory.NewStorage()
	g.r, err = git.Clone(storage, nil, &git.CloneOptions{
		URL:      g.config.RepoURL,
		Depth:    g.config.Depth,
		Progress: os.Stdout,
	})
	return err
}

func NewGit(config *Config) *Git {
	return &Git{config: config}
}
