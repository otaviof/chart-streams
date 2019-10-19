package repo

import (
	"testing"

	"github.com/otaviof/chart-streams/pkg/chartstreams/config"
)

func TestNewGitChartRepository_ExistingRemoteRepository_ShouldSucceed(t *testing.T) {
	cfg := &config.Config{Depth: 1, RepoURL: "https://github.com/helm/charts.git"}
	_, err := NewGitChartRepository(cfg)
	if err != nil {
		t.Error(err)
	}
}

func TestNewGitChartRepository_UnknownRemoteRepository_ShouldFail(t *testing.T) {
	cfg := &config.Config{Depth: 1, RepoURL: "https://example.com/charts.git"}
	_, err := NewGitChartRepository(cfg)
	if err == nil {
		t.Error("unknown remote repository should fail, but it isn't")
	}
}
