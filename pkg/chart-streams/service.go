package chartstreams

import (
	repo "k8s.io/helm/pkg/repo"
)

type ChartProvider interface {
	Initialize() error
	GetHelmChart(name, version string) error
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
	return gs.gitRepo.Clone()
}

func (gs *StreamChartProvider) GetHelmChart(name string, version string) error {
	return nil
}

func (gs *StreamChartProvider) GetIndexFile() (*repo.IndexFile, error) {
	return gs.index, nil
}
