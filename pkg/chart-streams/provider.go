package chartstreams

import (
	"fmt"
	"io"
	"io/ioutil"

	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"k8s.io/helm/pkg/chartutil"
	repo "k8s.io/helm/pkg/repo"
)

type ChartProvider interface {
	Initialize() error
	GetChart(name, version string) (*Package, error)
	GetIndexFile() (*repo.IndexFile, error)
}

type StreamChartProvider struct {
	config                    *Config
	gitRepo                   *Git
	index                     *repo.IndexFile
	chartNameVersionCommitMap map[string]plumbing.Hash
}

func NewStreamChartProvider(config *Config) *StreamChartProvider {
	g := NewGit(config)

	return &StreamChartProvider{
		config:                    config,
		gitRepo:                   g,
		chartNameVersionCommitMap: make(map[string]plumbing.Hash),
	}
}

func (gs *StreamChartProvider) AddVersionMapping(name, version string, hash plumbing.Hash) {
	chartNameVersion := fmt.Sprintf("%s/%s", name, version)

	gs.chartNameVersionCommitMap[chartNameVersion] = hash
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

			gs.AddVersionMapping(chartName, chartVersion, c.Hash)
		}
	}

	return nil
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

func (gs *StreamChartProvider) GetHashForChartVersion(name string, version string) *plumbing.Hash {
	if v, ok := gs.chartNameVersionCommitMap[version]; ok {
		return &v
	}
	return nil
}

func (gs *StreamChartProvider) GetChart(name string, version string) (*Package, error) {
	w, err := gs.gitRepo.r.Worktree()
	if err != nil {
		return nil, err
	}

	hash := gs.GetHashForChartVersion(name, version)
	if hash == nil {
		return nil, fmt.Errorf("GetChart(): couldn't find commit hash from specified version")
	}

	checkoutErr := w.Checkout(&git.CheckoutOptions{Hash: *hash})
	if checkoutErr != nil {
		return nil, checkoutErr
	}

	chartPath := w.Filesystem.Join(defaultChartRelativePath, name)

	p, err := buildChart(w, chartPath)
	if err != nil {
		return nil, fmt.Errorf("GetChart(): couldn't build package %s: %s", name, err)
	}

	return p, nil
}

func (gs *StreamChartProvider) GetIndexFile() (*repo.IndexFile, error) {
	return gs.index, nil
}
