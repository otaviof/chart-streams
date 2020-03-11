package provider

import (
	"fmt"

	"gopkg.in/src-d/go-git.v4"
	helmrepo "helm.sh/helm/v3/pkg/repo"

	"github.com/otaviof/chart-streams/pkg/chartstreams/chart"
	"github.com/otaviof/chart-streams/pkg/chartstreams/config"
	"github.com/otaviof/chart-streams/pkg/chartstreams/index"
	"github.com/otaviof/chart-streams/pkg/chartstreams/repo"
)

// GitChartProvider provides Helm charts from a specified Git repository.
type GitChartProvider struct {
	versionCommitMap map[string]repo.CommitInfo // mapping of chart version and git commits
	config           *config.Config             // global configuration
	gitRepo          *repo.GitChartRepo         // git repository handler
	index            *index.Index               // chart repository index instance
}

var _ ChartProvider = &GitChartProvider{}

// GetIndexFile returns the Helm server index file based on its Git repository contents.
func (g *GitChartProvider) GetIndexFile() (*helmrepo.IndexFile, error) {
	return g.index.IndexFile, nil
}

// initializeRepository instantiate a new chart repository.
func (g *GitChartProvider) initializeRepository() error {
	var err error
	g.gitRepo, err = repo.NewGitChartRepo(g.config)
	return err
}

// buildIndexFile instantiate global index.
func (g *GitChartProvider) buildIndexFile() error {
	builder := index.NewGitChartIndexBuilder(g.gitRepo, g.config.RelativeDir)
	var err error
	g.index, err = builder.Build()
	return err
}

// Initialize clones a Git repository and harvests, for each commit, Helm charts and their versions.
func (g *GitChartProvider) Initialize() error {
	if err := g.initializeRepository(); err != nil {
		return err
	}

	if err := g.buildIndexFile(); err != nil {
		return err
	}

	return nil
}

// GetChart returns a chart Package for the given chart name and version.
func (g *GitChartProvider) GetChart(name string, version string) (*chart.Package, error) {
	mapping := g.index.GetChartVersionMapping(name, version)
	if mapping == nil {
		return nil, fmt.Errorf("GetChart(): couldn't find commit hash from specified version")
	}

	w, err := g.gitRepo.Worktree()
	if err != nil {
		return nil, err
	}

	err = w.Checkout(&git.CheckoutOptions{Hash: mapping.Hash})
	if err != nil {
		return nil, err
	}

	chartPath := w.Filesystem.Join(g.config.RelativeDir, name)

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
		config:           config,
		versionCommitMap: make(map[string]repo.CommitInfo),
	}
}
