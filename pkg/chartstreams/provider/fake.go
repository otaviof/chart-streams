package provider

import (
	"bytes"
	"fmt"

	helmchart "helm.sh/helm/v3/pkg/chart"
	helmrepo "helm.sh/helm/v3/pkg/repo"

	"github.com/otaviof/chart-streams/pkg/chartstreams/chart"
	"github.com/otaviof/chart-streams/pkg/chartstreams/config"
)

type FakeChartProvider struct {
	cfg *config.Config
}

func (f *FakeChartProvider) Initialize() error {
	return nil
}

func (f *FakeChartProvider) GetChart(name, version string) (*chart.Package, error) {
	return &chart.Package{BytesBuffer: bytes.NewBuffer([]byte("package payload"))}, nil
}

func (f *FakeChartProvider) GetIndexFile() (*helmrepo.IndexFile, error) {
	indexFile := helmrepo.NewIndexFile()
	baseURL := fmt.Sprintf("http://%s", f.cfg.ListenAddr)
	metadata := &helmchart.Metadata{
		APIVersion: helmchart.APIVersionV1,
		Name:       "chart",
		Version:    "0.0.1",
	}
	indexFile.Add(metadata, "chart.tgz", baseURL, "digest")
	indexFile.SortEntries()
	return indexFile, nil
}

func NewFakeChartProvider(cfg *config.Config) *FakeChartProvider {
	return &FakeChartProvider{cfg: cfg}
}
