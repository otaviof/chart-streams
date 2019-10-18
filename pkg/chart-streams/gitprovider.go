package chartstreams

import (
	"fmt"
	"github.com/otaviof/chart-streams/pkg/chart-streams/chart"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"io"
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
	gitRepo                   *Git
	index                     *repo.IndexFile
	chartNameVersionCommitMap map[string]commitInfo
}

// GetIndexFile returns the Helm server index file based on its Git repository contents.
func (gs *GitChartProvider) GetIndexFile() (*repo.IndexFile, error) {
	return gs.index, nil
}

// AddVersionMapping maps a specific chart version with a commit hash and commit time.
func (gs *GitChartProvider) AddVersionMapping(name, version string, hash plumbing.Hash, t time.Time) {
	chartNameVersion := fmt.Sprintf("%s/%s", name, version)
	gs.chartNameVersionCommitMap[chartNameVersion] = commitInfo{
		Time: t,
		Hash: hash,
	}
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
	defer chartYamlFile.Close()

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
	gs.index = repo.NewIndexFile()

	if err := gs.gitRepo.Clone(); err != nil {
		return fmt.Errorf("Initialize(): error cloning the repository: %s", err)
	}

	commitIter, err := gs.gitRepo.AllCommits()
	if err != nil {
		return fmt.Errorf("Initialize(): error getting commit iterator: %s", err)
	}
	defer commitIter.Close()

	for {
		c, err := commitIter.Next()
		if err != nil && err != io.EOF {
			break
		}

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
			chartVersion := getChartVersion(w, chartPath)

			gs.AddVersionMapping(chartName, chartVersion, c.Hash, c.Committer.When)
		}
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

	w, err := gs.gitRepo.r.Worktree()
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

func NewStreamChartProvider(config *Config) *GitChartProvider {
	return &GitChartProvider{
		config:                    config,
		gitRepo:                   NewGit(config),
		chartNameVersionCommitMap: make(map[string]commitInfo),
	}
}
