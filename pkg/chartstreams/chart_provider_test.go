package chartstreams

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/otaviof/chart-streams/test/util"
)

func init() {
	SetLogLevel("trace")
}

func TestGitChartProvider(t *testing.T) {
	helmChartName := "one"
	helmRepoDir, err := util.ChartsRepoDir("../..")
	require.NoError(t, err, "on discovering test repo directory dir")
	helmRepoURL := fmt.Sprintf("file://%s", helmRepoDir)

	tempDir, err := ioutil.TempDir("", "chart-streams")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	cfg := &Config{
		RepoURL:     helmRepoURL,
		RelativeDir: "/",
		CloneDepth:  1,
	}
	g := NewGitChartProvider(cfg, tempDir)

	t.Run("Initialize", func(t *testing.T) {
		err := g.Initialize()
		assert.NoError(t, err)
	})

	var helmChartVersion string
	t.Run("GetIndexFile", func(t *testing.T) {
		index := g.GetIndexFile()
		assert.NotNil(t, index)
		assert.True(t, len(index.Entries) > 0)

		metadata, found := index.Entries[helmChartName]
		assert.True(t, found)
		helmChartVersion = metadata[0].Version
		assert.NotEmpty(t, helmChartVersion)
	})

	t.Run("GetChart", func(t *testing.T) {
		t.Logf("Helm-Chart '%s' version '%s'", helmChartName, helmChartVersion)
		pkg, err := g.GetChart(helmChartName, helmChartVersion)
		assert.NoError(t, err)
		assert.NotNil(t, pkg)
	})
}
