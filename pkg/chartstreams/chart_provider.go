package chartstreams

import (
	"fmt"

	git "github.com/libgit2/git2go/v33"
	helmrepo "helm.sh/helm/v3/pkg/repo"
)

type ChartProvider interface {
	Initialize() error
	GetChart(name, version string) (*Package, error)
	GetIndexFile() *helmrepo.IndexFile
	UpdateBranch(branch string) error
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

// UpdateBranch fetches the latest changes from the remote.
func (g *GitChartProvider) UpdateBranch(name string) error {

	if err := g.gitRepo.FetchBranch(name); err != nil {
		return fmt.Errorf("fetching branch '%s': %w", name, err)
	}
	if err := g.BuildIndexFile(); err != nil {
		return fmt.Errorf("building index file: %w", err)
	}
	return nil
}

func (g *GitChartProvider) BuildIndexFile() error {
	var err error
	g.indexFile, err = g.indexBuilder.Build()
	return err
}

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
	if err = g.BuildIndexFile(); err != nil {
		return err
	}
	return nil
}

func (g *GitChartProvider) BuildPackage(
	path string,
	info *CommitInfo,
	commit *git.Commit,
) (*Package, error) {
	files, err := g.gitRepo.GetFilesFromCommit(commit, path)
	if err != nil {
		return nil, fmt.Errorf("getting files from commit: %w", err)
	}
	p, err := LoadFiles(files, info.Time)
	if err != nil {
		return nil, fmt.Errorf("loading files from index: %w", err)
	}
	return p, nil
}

// GetChart returns a chart Package for the given chart name and version.
func (g *GitChartProvider) GetChart(name string, version string) (*Package, error) {
	commitInfo := g.indexBuilder.GetChartCommitInfo(name, version)
	if commitInfo == nil {
		return nil, fmt.Errorf("unable to find chart '%s' version '%s'", name, version)
	}

	c, err := g.gitRepo.LookupCommit(commitInfo.ID)
	if err != nil {
		return nil, fmt.Errorf("looking up commit %q: %w", commitInfo.ID, err)
	}

	p, err := g.BuildPackage(name, commitInfo, c)
	if err != nil {
		return nil, fmt.Errorf("instantiating package in %q: %w", name, err)
	}

	if err = p.Build(); err != nil {
		return nil, fmt.Errorf("building package for %q: %w", name, err)
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
