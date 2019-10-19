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
		{name: "non-existing local repository", repoURL: "/tmp/non-existing.git", shouldFail: true},
		{name: "existing local repository", repoURL: "/tmp/existing.git", shouldFail: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := &config.Config{Depth: 1, RepoURL: test.repoURL}
			_, err := NewGitChartRepository(cfg)

			if test.shouldFail && err == nil {
				t.Error("operation should fail but did not")
			}

			if !test.shouldFail && err != nil {
				t.Errorf("operation should not fail, but did: %s", err)
			}
		})
	}
}
