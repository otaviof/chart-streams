package chartstreams

import (
	"fmt"
	"sort"

	"gopkg.in/src-d/go-git.v4/plumbing/object"
	git "gopkg.in/src-d/go-git.v4"
	repo "k8s.io/helm/pkg/repo"
)

type ChartProvider interface {
	Initialize() error
	GetChart(name, version string) error
	GetIndexFile() (*repo.IndexFile, error)
}

type StreamChartProvider struct {
	config  *Config
	gitRepo *Git
	index   *repo.IndexFile
}

func NewStreamChartProvider(config *Config) *StreamChartProvider {
	g := NewGit(config)

	return &StreamChartProvider{
		config:  config,
		gitRepo: g,
	}
}

func (gs *StreamChartProvider) Initialize() error {
	gs.index = repo.NewIndexFile()

	if err := gs.gitRepo.Clone(); err != nil {
		return fmt.Errorf("Initialize(): error cloning the repository: %s", err)
	}

	commitIter, err := gs.gitRepo.AllCommits()
	if err != nil {
		return fmt.Errorf("Initialize(): error getting commit iterator: %s", err)
	}

	commitPrinterFunc := func(c *object.Commit) error {
		w, err := gs.gitRepo.r.Worktree()
		if err != nil {
			return err
		}

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

		return nil
	}

	err = commitIter.ForEach(commitPrinterFunc)
	if err != nil {
		return fmt.Errorf("Initialize(): error iterating commits: %s", err)
	}

	return nil
}

func (gs *StreamChartProvider) GetChart(name string, version string) error {
	return nil
}

func (gs *StreamChartProvider) GetIndexFile() (*repo.IndexFile, error) {
	return gs.index, nil
}
