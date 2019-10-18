package chartstreams

import (
	"k8s.io/helm/pkg/repo"

	"github.com/otaviof/chart-streams/pkg/chart-streams/chart"
)

type ChartProvider interface {
	Initialize() error
	GetChart(name, version string) (*chart.Package, error)
	GetIndexFile() (*repo.IndexFile, error)
}
