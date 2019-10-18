package chartstreams

import (
	"fmt"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"io"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/repo"
	"sort"
	"time"
)

type IndexBuilder interface {
	SetBasePath(string) IndexBuilder
	Build() (*repo.IndexFile, error)
}

type chartNameVersion struct {
	name    string
	version string
}

type gitChartIndexBuilder struct {
	gitChartRepository          *GitChartRepository
	chartNameVersionToCommitMap map[chartNameVersion]commitInfo
	basePath                    string
}

func (ib gitChartIndexBuilder) SetBasePath(basePath string) IndexBuilder {
	ib.basePath = basePath
	return ib
}

var _ IndexBuilder = gitChartIndexBuilder{}

func (ib gitChartIndexBuilder) addChartNameVersionToCommitMap(name string, version string, hash plumbing.Hash, t time.Time) {
	cnv := chartNameVersion{name: name, version: version}
	ib.chartNameVersionToCommitMap[cnv] = commitInfo{
		Time: t,
		Hash: hash,
	}
}

func (ib gitChartIndexBuilder) Build() (*repo.IndexFile, error) {
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

		w, err := ib.gitChartRepository.r.Worktree()
		if err != nil {
			return nil, err
		}

		checkoutErr := w.Checkout(&git.CheckoutOptions{Hash: c.Hash})
		if checkoutErr != nil {
			return nil, checkoutErr
		}

		chartDirEntries, readDirErr := w.Filesystem.ReadDir(defaultChartRelativePath)
		if readDirErr != nil {
			return nil, readDirErr
		}

		for _, entry := range chartDirEntries {
			chartName := entry.Name()
			chartPath := w.Filesystem.Join(defaultChartRelativePath, chartName)
			chartVersion := getChartVersion(w, chartPath)
			ib.addChartNameVersionToCommitMap(chartName, chartVersion, c.Hash, c.Committer.When)
		}
	}

	indexFile := repo.NewIndexFile()

	var allChartsVersions []chartNameVersion
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

	return indexFile, nil
}

func NewGitChartIndexBuilder(r *GitChartRepository) IndexBuilder {
	return &gitChartIndexBuilder{
		gitChartRepository:          r,
		chartNameVersionToCommitMap: map[chartNameVersion]commitInfo{},
	}
}
