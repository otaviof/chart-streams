package chartstreams

import (
	"fmt"
	"os"

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

func buildChart(wt *git.Worktree, chartPath string) (*Package, error) {

	p := &Package{}

	walkErr := worktree.Walk(wt, chartPath, func(wt *git.Worktree, path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		f, openErr := wt.Filesystem.Open(path)
		if openErr != nil {
			return openErr
		}
		defer f.Close()

		if !info.Mode().IsRegular() {
			return nil
		}

		p.Add(chartPath, info, f)

		return nil
	})

	if walkErr != nil {
		return nil, walkErr
	}

	// fmt.Printf("chart path: %s, chart files: %v\n\n", chartPath, chart.GetFiles())

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
