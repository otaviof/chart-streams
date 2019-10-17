package chartstreams

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/src-d/go-billy.v4/memfs"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/storage/memory"

	"github.com/otaviof/chart-streams/pkg/chart-streams/worktree"
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
		//Progress: os.Stdout,
	})

	if err != nil {
		return fmt.Errorf("Clone(): error cloning the repository")
	}

	return nil
}

const defaultChartRelativePath = "stable"

func buildChart(wt *git.Worktree, chartPath string, chartName string) (*Package, error) {

	p := NewPackage()
	defer p.Close()

	walkErr := worktree.Walk(wt, chartPath, func(wt *git.Worktree, path string, fileInfo os.FileInfo, err error) error {
		if fileInfo.IsDir() {
			return nil
		}

		f, openErr := wt.Filesystem.Open(path)
		if openErr != nil {
			return openErr
		}
		defer f.Close()

		if !fileInfo.Mode().IsRegular() {
			return nil
		}

		path = filepath.Join(chartName, strings.TrimPrefix(path, chartPath))

		if err := p.Add(path, fileInfo, f); err != nil {
			return err
		}

		return nil
	})

	if walkErr != nil {
		return nil, walkErr
	}

	return p, nil
}

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
