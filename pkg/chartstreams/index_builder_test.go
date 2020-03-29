package chartstreams

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
}

func TestIndexBuilder(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "chart-streams")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	cfg := &Config{
		RepoURL:     "https://github.com/helm/charts.git",
		RelativeDir: "stable",
		CloneDepth:  1,
	}

	g, err := NewGitRepo(cfg, tempDir)
	assert.NoError(t, err)

	i := NewIndexBuilder(cfg, g)
	index, err := i.Build()
	assert.NotNil(t, index)
	assert.NoError(t, err)
}
