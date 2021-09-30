package chartstreams

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/otaviof/chart-streams/test/util"

	git "github.com/libgit2/git2go/v31"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitRepo_NewGitRepoUpstream(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "chart-streams-")
	require.NoError(t, err)

	cfg := &Config{
		RepoURL:     "https://github.com/helm/charts.git",
		RelativeDir: "stable",
		CloneDepth:  1,
	}
	r, err := NewGitRepo(cfg, tempDir)
	require.NoError(t, err)

	t.Run("CommitIter", func(t *testing.T) {
		err := r.CommitIter(func(branch string, c *git.Commit, tree *git.Tree, head bool) error {
			requirePopulatedGitDir(t, tempDir)
			return nil
		})
		assert.NoError(t, err)
	})

	r, err = NewGitRepo(cfg, tempDir)
	require.NoError(t, err, "re-opening working-dir after successful clone")

	_ = os.RemoveAll(tempDir)
}

func TestGitRepo_NewGitRepoLocal(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "chart-streams-")
	require.NoError(t, err)

	localRepoDir, err := util.ChartsRepoDir("../..")
	require.NoError(t, err)

	cfg := &Config{RepoURL: fmt.Sprintf("file://%s", localRepoDir), RelativeDir: "/"}

	r, err := NewGitRepo(cfg, tempDir)
	require.NoError(t, err)

	err = r.CommitIter(func(branch string, c *git.Commit, tree *git.Tree, head bool) error {
		requirePopulatedGitDir(t, tempDir)
		return nil
	})
	require.NoError(t, err)

	_, err = NewGitRepo(cfg, tempDir)
	require.NoError(t, err, "re-opening working-dir after successful clone")

	_ = os.RemoveAll(tempDir)
}
