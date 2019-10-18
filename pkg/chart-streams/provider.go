package chartstreams

import (
	"fmt"
	"io"
	"io/ioutil"
	"time"

	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"k8s.io/helm/pkg/chartutil"
	repo "k8s.io/helm/pkg/repo"

	"github.com/otaviof/chart-streams/pkg/chart-streams/chart"
)

type ChartProvider interface {
	Initialize() error
	GetChart(name, version string) (*chart.Package, error)
	GetIndexFile() (*repo.IndexFile, error)
}

type ChartNameVersionCommitMap struct {
	Time time.Time
	Hash plumbing.Hash
}

type StreamChartProvider struct {
	config                    *Config
	gitRepo                   *Git
	index                     *repo.IndexFile
	chartNameVersionCommitMap map[string]ChartNameVersionCommitMap
}

func (gs *StreamChartProvider) GetIndexFile() (*repo.IndexFile, error) {
	return gs.index, nil
}

func (gs *StreamChartProvider) AddVersionMapping(name, version string, hash plumbing.Hash, t time.Time) {
	chartNameVersion := fmt.Sprintf("%s/%s", name, version)
	gs.chartNameVersionCommitMap[chartNameVersion] = ChartNameVersionCommitMap{
		Time: t,
		Hash: hash,
	}
}

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

func (gs *StreamChartProvider) Initialize() error {
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

func (gs *StreamChartProvider) GetChartVersionMapping(name string, version string) *ChartNameVersionCommitMap {
	chartNameVersion := fmt.Sprintf("%s/%s", name, version)
	if m, ok := gs.chartNameVersionCommitMap[chartNameVersion]; ok {
		return &m
	}
	return nil
}

func (gs *StreamChartProvider) GetChart(name string, version string) (*chart.Package, error) {
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

	p, err := chart.
		NewBillyChartBuilder(w.Filesystem).
		SetChartName(name).
		SetChartPath(chartPath).
		SetCommitTime(mapping.Time).
		Build()
	if err != nil {
		return nil, fmt.Errorf("GetChart(): couldn't build package %s: %s", name, err)
	}

	return p, nil
}

func NewStreamChartProvider(config *Config) *StreamChartProvider {
	return &StreamChartProvider{
		config:                    config,
		gitRepo:                   NewGit(config),
		chartNameVersionCommitMap: make(map[string]ChartNameVersionCommitMap),
	}
}
