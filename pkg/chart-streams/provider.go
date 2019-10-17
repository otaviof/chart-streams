package chartstreams

import (
	"fmt"

	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	repo "k8s.io/helm/pkg/repo"
)

type ChartProvider interface {
	Initialize() error
	GetChart(name, version string) (*Package, error)
	GetIndexFile() (*repo.IndexFile, error)
}

type StreamChartProvider struct {
	config                    *Config
	gitRepo                   *Git
	index                     *repo.IndexFile
	chartNameVersionCommitMap map[string]plumbing.Hash
}

func NewStreamChartProvider(config *Config) *StreamChartProvider {
	g := NewGit(config)

	return &StreamChartProvider{
		config:                    config,
		gitRepo:                   g,
		chartNameVersionCommitMap: make(map[string]plumbing.Hash),
	}
}

func (gs *StreamChartProvider) AddVersionMapping(name, version string, hash plumbing.Hash) {
	chartNameVersion := fmt.Sprintf("%s/%s", name, version)

	gs.chartNameVersionCommitMap[chartNameVersion] = hash
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

		for _, entry := range chartDirEntries {
			chartName := entry.Name()
			chartPath := w.Filesystem.Join(defaultChartRelativePath, chartName)
			chartVersion := getChartVersion(chartPath)

			gs.AddVersionMapping(chartName, chartVersion, c.Hash)
		}

		return nil
	}

	err = commitIter.ForEach(commitPrinterFunc)
	if err != nil {
		return fmt.Errorf("Initialize(): error iterating commits: %s", err)
	}

	return nil
}

func getChartVersion(chartPath string) string {
	return "1.0.0"
}

func getCommitForVersion(version string) *object.Commit {
	return nil
}

func (gs *StreamChartProvider) GetChart(name string, version string) (*Package, error) {
	w, err := gs.gitRepo.r.Worktree()
	if err != nil {
		return nil, err
	}

	commit := getCommitForVersion(version)
	if commit == nil {
		return nil, fmt.Errorf("GetChart(): couldn't find commit hash from specified version")
	}

	checkoutErr := w.Checkout(&git.CheckoutOptions{Hash: commit.Hash})
	if checkoutErr != nil {
		return nil, checkoutErr
	}

	chartPath := w.Filesystem.Join(defaultChartRelativePath, name)

	p, err := buildChart(w, chartPath)
	if err != nil {
		return nil, fmt.Errorf("GetChart(): couldn't build package %s: %s", name, err)
	}

	return p, nil
}

func (gs *StreamChartProvider) GetIndexFile() (*repo.IndexFile, error) {
	return gs.index, nil
}
