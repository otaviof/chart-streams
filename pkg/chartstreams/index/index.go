package index

import (
	helmrepo "k8s.io/helm/pkg/repo"

	"github.com/otaviof/chart-streams/pkg/chartstreams/repo"
)

// Builder provides a fluent API to build Helm index files.
type Builder interface {
	SetBasePath(string) Builder
	Build() (*Index, error)
}

// Index represent a chart index of some sort.
type Index struct {
	IndexFile                   *helmrepo.IndexFile
	chartNameVersionToCommitMap map[ChartNameVersion]repo.CommitInfo
}

func (i *Index) GetChartVersionMapping(name string, version string) *repo.CommitInfo {
	chartNameVersion := ChartNameVersion{
		name:    name,
		version: version,
	}

	if m, ok := i.chartNameVersionToCommitMap[chartNameVersion]; ok {
		return &m
	}
	return nil
}
