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
		URL:      "https://github.com/helm/charts.git",
		Depth:    1,
		Progress: os.Stdout,
	})
	return err
}

func NewGit(config *Config) *Git {
	return &Git{config: config}
}
