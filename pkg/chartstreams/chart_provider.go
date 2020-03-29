package chartstreams

import (
	"fmt"

	helmrepo "helm.sh/helm/v3/pkg/repo"
)

type ChartProvider interface {
	Initialize() error
	GetChart(name, version string) (*Package, error)
	GetIndexFile() *helmrepo.IndexFile
}

// versionCommitMap	mapping of chart version and git commits
type versionCommitMap map[string]CommitInfo

// GitChartProvider provides Helm charts from a specified Git repository.
type GitChartProvider struct {
	versionCommit versionCommitMap    // relationship between chart version and commit metdata
	cfg           *Config             // global configuration
	gitRepo       *GitRepo            // git repository handler
	workingDir    string              // git working directory
	indexBuilder  *IndexBuilder       // repository index builder
	indexFile     *helmrepo.IndexFile // repository index instance
}

var _ ChartProvider = &GitChartProvider{}

// GetIndexFile returns the helm repository index-file instance.
func (g *GitChartProvider) GetIndexFile() *helmrepo.IndexFile {
	return g.indexFile
}

// Initialize clones a Git repository and harvests, for each commit, Helm charts and their versions.
func (g *GitChartProvider) Initialize() error {
	var err error
	if g.gitRepo, err = NewGitRepo(g.cfg, g.workingDir); err != nil {
		return err
	}
	g.indexBuilder = NewIndexBuilder(g.cfg, g.gitRepo)
	if g.indexFile, err = g.indexBuilder.Build(); err != nil {
		return err
	}
	return nil
}

// GetChart returns a chart Package for the given chart name and version.
func (g *GitChartProvider) GetChart(name string, version string) (*Package, error) {
	commitInfo := g.indexBuilder.GetChartCommitInfo(name, version)
	if commitInfo == nil {
		return nil, fmt.Errorf("unable to find chart '%s' version '%s'", name, version)
	}

	c, err := g.gitRepo.LookupCommit(commitInfo.ID)
	if err != nil {
		return nil, err
	}
	if err = g.gitRepo.CheckoutCommit(commitInfo.Revision, c); err != nil {
		return nil, err
	}

	absPath := g.indexBuilder.ChartAbsPath(name)
	p, err := NewPackage(absPath, commitInfo.Time)
	if err != nil {
		return nil, err
	}
	if err = p.Build(); err != nil {
		return nil, err
	}
	return p, nil
}

// NewGitChartProvider returns an chart provider that can build and index charts from a Git repository.
func NewGitChartProvider(cfg *Config, workingDir string) *GitChartProvider {
	return &GitChartProvider{
		cfg:           cfg,
		workingDir:    workingDir,
		versionCommit: make(versionCommitMap),
	}
}
