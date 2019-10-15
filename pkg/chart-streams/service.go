package chartstreams

import (
	repo "k8s.io/helm/pkg/repo"
)

type ChartProvider interface {
	Initialize() error
	GetHelmChart(name, version string) error
	GetIndexFile() (*repo.IndexFile, error)
}

type ChartStreamService struct {
	config  *Config
	gitRepo *Git
	index   *repo.IndexFile
}

func NewChartStreamService(config *Config) *ChartStreamService {
	g := NewGit(config)

	return &ChartStreamService{
		config:  config,
		gitRepo: g,
	}
}

func (gs *ChartStreamService) Initialize() error {
	gs.index = repo.NewIndexFile()
	return gs.gitRepo.Clone()
}

func (gs *ChartStreamService) GetHelmChart(name string, version string) error {
	return nil
}

func (gs *ChartStreamService) GetIndexFile() (*repo.IndexFile, error) {
	return gs.index, nil
}
