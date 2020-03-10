package index

import (
	"helm.sh/helm/v3/pkg/chart"
	helmrepo "helm.sh/helm/v3/pkg/repo"

	"github.com/otaviof/chart-streams/pkg/chartstreams/repo"
)

// Builder provides a fluent API to build Helm index files.
type Builder interface {
	SetBasePath(string) Builder
	Build() (*Index, error)
}

// Index represent a chart index of some sort.
type Index struct {
	IndexFile                *helmrepo.IndexFile
	chartMetadataToCommitMap map[*chart.Metadata]repo.CommitInfo
}

// GetChartVersionMapping based on chart name and version, return repository commit information.
func (i *Index) GetChartVersionMapping(name string, version string) *repo.CommitInfo {
	for metadata, commitInfo := range i.chartMetadataToCommitMap {
		if metadata.Name == name && metadata.Version == version {
			return &commitInfo
		}
	}
	return nil
}
