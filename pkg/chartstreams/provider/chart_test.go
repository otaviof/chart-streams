package provider

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/otaviof/chart-streams/pkg/chartstreams/config"
)

const (
	helmRepoURL         = "https://github.com/helm/charts.git"
	helmRepoRelativeDir = "stable"
	helmChartName       = "traefik"
)

func TestGitChartProvider(t *testing.T) {
	config := &config.Config{
		RepoURL:     helmRepoURL,
		RelativeDir: helmRepoRelativeDir,
		CloneDepth:  1,
	}
	g := NewGitChartProvider(config)

	t.Run("Initialize", func(t *testing.T) {
		err := g.Initialize()
		assert.NoError(t, err)
	})

	var helmChartVersion string
	t.Run("GetIndexFile", func(t *testing.T) {
		index, err := g.GetIndexFile()
		assert.NoError(t, err)
		assert.NotNil(t, index)
		assert.True(t, len(index.Entries) > 0)

		metadata, found := index.Entries[helmChartName]
		assert.True(t, found)
		helmChartVersion = metadata[0].GetApiVersion()
		assert.NotEmpty(t, helmChartVersion)
	})

	t.Run("GetChart", func(t *testing.T) {
		t.Logf("Helm-Chart '%s' version '%s'", helmChartName, helmChartVersion)
		pkg, err := g.GetChart(helmChartName, helmChartVersion)
		assert.NoError(t, err)
		assert.NotNil(t, pkg)
	})
}
