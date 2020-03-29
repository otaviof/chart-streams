package chartstreams

import (
	"io/ioutil"
	"os"
	"testing"

	git "github.com/libgit2/git2go/v28"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepoGitRepo(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "chart-streams")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	cfg := &Config{
		RepoURL:     "https://github.com/helm/charts.git",
		RelativeDir: "stable",
		CloneDepth:  1,
	}
	r, err := NewGitRepo(cfg, tempDir)
	assert.NoError(t, err)

	t.Run("CommitIter", func(t *testing.T) {
		err := r.CommitIter(func(branch string, c *git.Commit, tree *git.Tree, head bool) error {
			_, err := c.ShortId()
			if err != nil {
				return err
			}
			// log.Printf("%s\n", shortID)

			return tree.Walk(func(entryPath string, t *git.TreeEntry) int {
				// log.Printf("entryPath='%s'\n", entryPath)
				return 0
			})
		})
		assert.NoError(t, err)
	})
}
