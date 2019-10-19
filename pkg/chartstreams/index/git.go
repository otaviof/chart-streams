package index

import (
	"fmt"
	"io"
	"io/ioutil"
	"sort"
	"time"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/proto/hapi/chart"
	helmrepo "k8s.io/helm/pkg/repo"

	"github.com/otaviof/chart-streams/pkg/chartstreams/repo"
)

type ChartNameVersion struct {
	name    string
	version string
}

type gitChartIndexBuilder struct {
	gitChartRepository          *repo.GitChartRepository
	chartNameVersionToCommitMap map[ChartNameVersion]repo.CommitInfo
	basePath                    string
}

func (ib gitChartIndexBuilder) SetBasePath(basePath string) Builder {
	ib.basePath = basePath
	return ib
}

var _ Builder = gitChartIndexBuilder{}

func (ib gitChartIndexBuilder) addChartNameVersionToCommitMap(name string, version string, hash plumbing.Hash, t time.Time) {
	cnv := ChartNameVersion{name: name, version: version}
	ib.chartNameVersionToCommitMap[cnv] = repo.CommitInfo{
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

func (ib gitChartIndexBuilder) Build() (*Index, error) {
	commitIter, err := ib.gitChartRepository.AllCommits()
	if err != nil {
		return nil, fmt.Errorf("Initialize(): error getting commit iterator: %s", err)
	}
	defer commitIter.Close()

	for {
		c, err := commitIter.Next()
		if err != nil && err != io.EOF {
			break
		}

		w, err := ib.gitChartRepository.Worktree()
		if err != nil {
			return nil, err
		}

		checkoutErr := w.Checkout(&git.CheckoutOptions{Hash: c.Hash})
		if checkoutErr != nil {
			return nil, checkoutErr
		}

		chartDirEntries, readDirErr := w.Filesystem.ReadDir(ib.basePath)
		if readDirErr != nil {
			return nil, readDirErr
		}

		for _, entry := range chartDirEntries {
			chartName := entry.Name()
			chartPath := w.Filesystem.Join(ib.basePath, chartName)
			chartVersion := getChartVersion(w, chartPath)
			ib.addChartNameVersionToCommitMap(chartName, chartVersion, c.Hash, c.Committer.When)
		}
	}

	indexFile := helmrepo.NewIndexFile()

	var allChartsVersions []ChartNameVersion
	for k := range ib.chartNameVersionToCommitMap {
		allChartsVersions = append(allChartsVersions, k)
	}

	sort.Slice(allChartsVersions, func(i, j int) bool {
		if allChartsVersions[i].name < allChartsVersions[j].name {
			return true
		}
		if allChartsVersions[i].name == allChartsVersions[j].name {
			return allChartsVersions[i].version < allChartsVersions[j].version
		}
		return false
	})

	for _, c := range allChartsVersions {
		m := &chart.Metadata{
			Name:       c.name,
			ApiVersion: c.version,
		}
		baseUrl := fmt.Sprintf("/chart/%s/%s", c.name, c.version)
		indexFile.Add(m, "chart.tgz", baseUrl, "deadbeef")
	}

	file := &Index{
		IndexFile:                   indexFile,
		chartNameVersionToCommitMap: ib.chartNameVersionToCommitMap,
	}

	return file, nil
}

// NewGitChartIndexBuilder creates an index builder for GitChartRepository.
func NewGitChartIndexBuilder(r *repo.GitChartRepository) Builder {
	return &gitChartIndexBuilder{
		gitChartRepository:          r,
		chartNameVersionToCommitMap: map[ChartNameVersion]repo.CommitInfo{},
	}
}
