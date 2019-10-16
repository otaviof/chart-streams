package chartstreams

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"sort"

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

type commitIterWrapper struct {
	commitIter object.CommitIter
	r          *git.Repository
}

const defaultChartRelativePath = "stable"

func buildChart(wt *git.Worktree, chartPath string) ([]byte, error) {

	var b bytes.Buffer

	mw := bufio.NewWriter(&b)

	gzw := gzip.NewWriter(mw)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

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

		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}

		header.Name = path

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if _, err := io.Copy(tw, f); err != nil {
			return err
		}

		return nil
	})

	if walkErr != nil {
		return nil, walkErr
	}

	// fmt.Printf("chart path: %s, chart files: %v\n\n", chartPath, chart.GetFiles())

	return b.Bytes(), nil
}

func (i *commitIterWrapper) ForEach(f func(*Commit) error) error {
	err := i.commitIter.ForEach(func(c *object.Commit) error {

		w, err := i.r.Worktree()
		if err != nil {
			return err
		}

		// TODO: SemVer to Git SHA1

		checkoutErr := w.Checkout(&git.CheckoutOptions{Hash: c.Hash})
		if checkoutErr != nil {
			return checkoutErr
		}

		chartDirEntries, readDirErr := w.Filesystem.ReadDir(defaultChartRelativePath)
		if readDirErr != nil {
			return readDirErr
		}

		var charts []string
		for _, entry := range chartDirEntries {
			chartPath := w.Filesystem.Join(defaultChartRelativePath, entry.Name())
			charts = append(charts, chartPath)
			b, err := buildChart(w, chartPath)
			if err != nil {
				return err
			}
			fmt.Printf("chart: %s, bytes: %v", chartPath, b)
		}

		sort.Strings(charts)

		fmt.Printf("charts: %v", charts)

		m := &Commit{
			"commitId": c.Hash.String(),
		}

		return f(m)
	})

	if err.Error() != "object not found" {
		return err
	}

	return nil
}

func (g *Git) AllCommits() (CommitIter, error) {
	ref, err := g.r.Head()
	if err != nil {
		return nil, fmt.Errorf("AllCommits: error finding Head reference: %s", err)
	}

	commitIter, err := g.r.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return nil, fmt.Errorf("AllCommits: error obtaining commitIter: %s", err)
	}

	return &commitIterWrapper{commitIter: commitIter,
		r: g.r,
	}, nil
}

func NewGit(config *Config) *Git {
	return &Git{config: config}
}
