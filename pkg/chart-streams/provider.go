package chartstreams

import (
	"fmt"

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

	commitPrinterFunc := func(c *Commit) error {
		fmt.Printf("commit: %v", c)
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
