package chartstreams

import (
	"testing"
)

func TestGit_Clone(t *testing.T) {
	config := &Config{Depth: 1, RepoURL: "https://github.com/helm/charts.git"}
	g := NewGit(config)
	g.Clone()
}
