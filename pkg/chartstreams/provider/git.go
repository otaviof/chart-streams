package provider

import (
	"fmt"

	"gopkg.in/src-d/go-git.v4"
	helmrepo "k8s.io/helm/pkg/repo"

	"github.com/otaviof/chart-streams/pkg/chartstreams/chart"
	"github.com/otaviof/chart-streams/pkg/chartstreams/config"
	"github.com/otaviof/chart-streams/pkg/chartstreams/index"
	"github.com/otaviof/chart-streams/pkg/chartstreams/repo"
)

const defaultChartRelativePath = "stable"

// GitChartProvider provides Helm charts from a specified Git repository.
type GitChartProvider struct {
	chartNameVersionCommitMap map[string]repo.CommitInfo
	config                    *config.Config
	gitRepository             *repo.GitChartRepository
	index                     *index.Index
}

var _ ChartProvider = &GitChartProvider{}

// GetIndexFile returns the Helm server index file based on its Git repository contents.
func (gs *GitChartProvider) GetIndexFile() (*helmrepo.IndexFile, error) {
	return gs.index.IndexFile, nil
}

func (gs *GitChartProvider) initializeRepository() error {
	var err error
	gs.gitRepository, err = repo.NewGitChartRepository(gs.config)
	return err

}

func (gs *GitChartProvider) buildIndexFile() error {
	var err error
	gs.index, err =
		index.NewGitChartIndexBuilder(gs.gitRepository).
			SetBasePath(defaultChartRelativePath).
			Build()
	return err
}

// Initialize clones a Git repository and harvests, for each commit, Helm charts and their versions.
func (gs *GitChartProvider) Initialize() error {
	if err := gs.initializeRepository(); err != nil {
		return err
	}

	if err := gs.buildIndexFile(); err != nil {
		return err
	}

	return nil
}

// GetChart returns a chart Package for the given chart name and version.
func (gs *GitChartProvider) GetChart(name string, version string) (*chart.Package, error) {
	mapping := gs.index.GetChartVersionMapping(name, version)
	if mapping == nil {
		return nil, fmt.Errorf("GetChart(): couldn't find commit hash from specified version")
	}

	w, err := gs.gitRepository.Worktree()
	if err != nil {
		return nil, err
	}

	checkoutErr := w.Checkout(&git.CheckoutOptions{Hash: mapping.Hash})
	if checkoutErr != nil {
		return nil, checkoutErr
	}

	chartPath := w.Filesystem.Join(defaultChartRelativePath, name)

	p, err := chart.NewBillyChartBuilder(w.Filesystem).
		SetChartName(name).
		SetChartPath(chartPath).
		SetCommitTime(mapping.Time).
		Build()
	if err != nil {
		return nil, fmt.Errorf("GetChart(): couldn't build package %s: %s", name, err)
	}

	return p, nil
}

// NewGitChartProvider returns an chart provider that can build and index charts from a Git repository.
func NewGitChartProvider(config *config.Config) *GitChartProvider {
	return &GitChartProvider{
		config:                    config,
		chartNameVersionCommitMap: make(map[string]repo.CommitInfo),
	}
}
