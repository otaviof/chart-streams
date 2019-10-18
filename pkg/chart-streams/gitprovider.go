package chartstreams

import (
	"fmt"
	"github.com/otaviof/chart-streams/pkg/chart-streams/chart"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"io/ioutil"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/repo"
	"time"
)

type commitInfo struct {
	Time time.Time
	Hash plumbing.Hash
}

// GitChartProvider provides Helm charts from a specified Git repository.
type GitChartProvider struct {
	config                    *Config
	gitRepository             *GitChartRepository
	indexFile                 *repo.IndexFile
	chartNameVersionCommitMap map[string]commitInfo
}

// GetIndexFile returns the Helm server index file based on its Git repository contents.
func (gs *GitChartProvider) GetIndexFile() (*repo.IndexFile, error) {
	return gs.indexFile, nil
}

// getChartVersion returns the version of a chart in the current Git repository work tree.
func getChartVersion(wt *git.Worktree, chartPath string) string {
	chartDirInfo, err := wt.Filesystem.Lstat(chartPath)
	if err != nil {
		return ""
	}

	if !chartDirInfo.IsDir() {
		return ""
	}

	chartYamlPath := wt.Filesystem.Join(chartPath, "Chart.yaml")
	chartYamlFile, err := wt.Filesystem.Open(chartYamlPath)
	if err != nil {
		return ""
	}
	defer func() {
		_ = chartYamlFile.Close()
	}()

	b, err := ioutil.ReadAll(chartYamlFile)
	if err != nil {
		return ""
	}

	chartYaml, err := chartutil.UnmarshalChartfile(b)
	if err != nil {
		return ""
	}

	return chartYaml.GetVersion()
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

func (gs *GitChartProvider) GetChartVersionMapping(name string, version string) *commitInfo {
	chartNameVersion := fmt.Sprintf("%s/%s", name, version)
	if m, ok := gs.chartNameVersionCommitMap[chartNameVersion]; ok {
		return &m
	}
	return nil
}

func (gs *GitChartProvider) GetChart(name string, version string) (*chart.Package, error) {
	mapping := gs.GetChartVersionMapping(name, version)
	if mapping == nil {
		return nil, fmt.Errorf("GetChart(): couldn't find commit hash from specified version")
	}

	w, err := gs.gitRepository.r.Worktree()
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

func (gs *GitChartProvider) buildIndexFile() error {
	var err error
	gs.indexFile, err =
		NewGitChartIndexBuilder(gs.gitRepository).
			SetBasePath(defaultChartRelativePath).
			Build()
	return err
}

func (gs *GitChartProvider) initializeRepository() error {
	return gs.gitRepository.Clone()
}

func NewStreamChartProvider(config *Config) *GitChartProvider {
	return &GitChartProvider{
		config:                    config,
		gitRepository:             NewGitChartRepository(config),
		chartNameVersionCommitMap: make(map[string]commitInfo),
	}
}
