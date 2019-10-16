package chartstreams

import "testing"

func TestInitialize(t *testing.T) {
	c := &Config{
		RepoURL: "https://github.com/helm/charts.git",
		Depth:   1,
	}

	p := NewStreamChartProvider(c)

	err := p.Initialize()

	if err != nil {
		t.Fatal(err)
	}
}
