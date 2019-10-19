package repo

import (
	"testing"

	"github.com/otaviof/chart-streams/pkg/chartstreams/config"
)

func TestNewGitChartRepository(t *testing.T) {
	tests := []struct {
		name       string
		repoURL    string
		shouldFail bool
	}{
		{name: "existing remote repository", repoURL: "https://github.com/helm/charts.git", shouldFail: false},
		{name: "non-existing remote repository", repoURL: "https://example.com/charts.git", shouldFail: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := &config.Config{Depth: 1, RepoURL: test.repoURL}
			_, err := NewGitChartRepository(cfg)

			if test.shouldFail && err == nil {
				t.Fail()
			}

			if !test.shouldFail && err != nil {
				t.Fail()
			}
		})
	}
}
