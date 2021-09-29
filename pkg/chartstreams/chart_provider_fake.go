package chartstreams

import (
	"bytes"
	"fmt"

	helmchart "helm.sh/helm/v3/pkg/chart"
	helmrepo "helm.sh/helm/v3/pkg/repo"
)

type FakeChartProvider struct {
	cfg *Config
}

func (f *FakeChartProvider) Initialize() error {
	return nil
}

func (f *FakeChartProvider) GetChart(name, version string) (*Package, error) {
	return &Package{b: bytes.NewBuffer([]byte("package payload"))}, nil
}

func (f *FakeChartProvider) GetIndexFile() *helmrepo.IndexFile {
	indexFile := helmrepo.NewIndexFile()
	baseURL := fmt.Sprintf("http://%s", f.cfg.ListenAddr)
	metadata := &helmchart.Metadata{
		APIVersion: helmchart.APIVersionV1,
		Name:       "chart",
		Version:    "0.0.1",
	}

	// NOTE(isutton): I was inclined to not ignore the error and
	// replace with the code below, but it didn't feel important
	// enough to bother for now.
	_ = indexFile.MustAdd(metadata, "chart.tgz", baseURL, "digest")

	// err := indexFile.MustAdd(metadata, "chart.tgz", baseURL, "digest")
	// if err != nil {
	// 	panic(err)
	// }

	indexFile.SortEntries()
	return indexFile
}

func NewFakeChartProvider(cfg *Config) *FakeChartProvider {
	return &FakeChartProvider{cfg: cfg}
}
