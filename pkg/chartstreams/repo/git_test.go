package repo

import (
	"github.com/otaviof/chart-streams/pkg/chartstreams/config"

	"testing"
)

func TestGitChartRepository_RemoteRepository(t *testing.T) {
	cfg := &config.Config{Depth: 1, RepoURL: "https://github.com/helm/charts.git"}
	_, err := NewGitChartRepository(cfg)
	if err != nil {
		t.Error(err)
	}
}
