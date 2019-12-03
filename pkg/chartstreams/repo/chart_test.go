package repo

import (
	"io/ioutil"
	"os"
	"testing"

	"gopkg.in/src-d/go-git.v4"

	"github.com/otaviof/chart-streams/pkg/chartstreams/config"
)

type TestCase struct {
	name       string
	repoURL    string
	shouldFail bool
	preCmd     *func(c *TestCase) error
	postCmd    *func()
}

const helmChartsReferenceURL = "https://github.com/helm/charts.git"

func (c *TestCase) PreCmd() error {
	if c.preCmd != nil {
		return (*c.preCmd)(c)
	}
	return nil
}

func (c *TestCase) PostCmd() {
	if c.postCmd != nil {
		(*c.postCmd)()
	}
}

// existingTestCase builds a test case for an existing local repository by creating a directory and
// cloning a repository into it, and use it as clone source.
func existingTestCase(t *testing.T) TestCase {
	repoPath, err := ioutil.TempDir("", "test-new-git-chart-repository-")
	if err != nil {
		t.Fatal(err)
	}

	preCmd := func(c *TestCase) error {
		_, err := git.PlainClone(
			repoPath,
			true,
			&git.CloneOptions{URL: helmChartsReferenceURL},
		)
		return err
	}

	postCmdFn := func() {
		_ = os.RemoveAll(repoPath)
	}

	return TestCase{
		name:       "existing local repository",
		repoURL:    repoPath,
		shouldFail: false,
		preCmd:     &preCmd,
		postCmd:    &postCmdFn,
	}
}

// nonExistingTestCase builds a test case for non existing local repositories by creating and
// removing a temporary directory and use it as clone source.
func nonExistingTestCase(t *testing.T) TestCase {
	repoPath, err := ioutil.TempDir("", "non-existing-local-repository-test-case-")
	if err != nil {
		t.Fatal(err)
	}

	_ = os.RemoveAll(repoPath)

	return TestCase{
		name:       "non-existing local repository",
		repoURL:    repoPath,
		shouldFail: true,
	}
}

func TestNewGitChartRepository(t *testing.T) {
	tests := []TestCase{
		{
			name:       "existing remote repository",
			repoURL:    helmChartsReferenceURL,
			shouldFail: false,
		},
		{
			name:       "non-existing remote repository",
			repoURL:    "https://example.com/charts.git",
			shouldFail: true,
		},
		existingTestCase(t),
		nonExistingTestCase(t),
	}

	for _, test := range tests {
		if err := test.PreCmd(); err != nil {
			t.Error(err)
			continue
		}

		t.Run(test.name, func(t *testing.T) {
			cfg := &config.Config{RepoURL: test.repoURL, CloneDepth: 1}
			_, err := NewGitChartRepo(cfg)

			if test.shouldFail && err == nil {
				t.Error("operation should fail but did not")
			}

			if !test.shouldFail && err != nil {
				t.Errorf("operation should not fail, but did: %s", err)
			}
		})

		test.PostCmd()
	}
}
