package repo

import (
	"io/ioutil"
	"os"
	"testing"

	"gopkg.in/src-d/go-git.v4"

	"github.com/otaviof/chart-streams/pkg/chartstreams/config"
)

func TestNewGitChartRepository(t *testing.T) {
	tests := []newGitChartRepositoryTestCase{
		{name: "existing remote repository", repoURL: helmChartsReferenceURL, shouldFail: false},
		{name: "non-existing remote repository", repoURL: "https://example.com/charts.git", shouldFail: true},
		newNonExistingLocalRepositoryTestCase(t),
		newExistingLocalRepositoryTestCase(t),
	}
	for _, test := range tests {
		if err := test.PreCmd(); err != nil {
			t.Error(err)
			continue
		}

		t.Run(test.name, func(t *testing.T) {
			cfg := &config.Config{Depth: 1, RepoURL: test.repoURL}
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

type newGitChartRepositoryTestCase struct {
	name       string
	repoURL    string
	shouldFail bool
	preCmd     *func(c *newGitChartRepositoryTestCase) error
	postCmd    *func()
}

func (c *newGitChartRepositoryTestCase) PreCmd() error {
	if c.preCmd != nil {
		return (*c.preCmd)(c)
	}
	return nil
}

func (c *newGitChartRepositoryTestCase) PostCmd() {
	if c.postCmd != nil {
		(*c.postCmd)()
	}
}

const helmChartsReferenceURL = "https://github.com/helm/charts.git"

// newExistingLocalRepositoryTestCase builds a test case for an existing local repository by creating a directory and
// cloning a repository into it, and use it as clone source.
func newExistingLocalRepositoryTestCase(t *testing.T) newGitChartRepositoryTestCase {
	existingLocalRepositoryPath, err := ioutil.TempDir("", "test-new-git-chart-repository-")
	if err != nil {
		t.Fatal(err)
	}

	existingLocalRepositoryPreCmd := func(c *newGitChartRepositoryTestCase) error {
		_, err := git.PlainClone(existingLocalRepositoryPath, true, &git.CloneOptions{URL: helmChartsReferenceURL})
		return err
	}

	cleanupExistingLocalRepositoryCmd := func() {
		_ = os.RemoveAll(existingLocalRepositoryPath)
	}

	return newGitChartRepositoryTestCase{
		name:       "existing local repository",
		repoURL:    existingLocalRepositoryPath,
		shouldFail: false,
		preCmd:     &existingLocalRepositoryPreCmd,
		postCmd:    &cleanupExistingLocalRepositoryCmd,
	}
}

// newNonExistingLocalRepositoryTestCase builds a test case for non existing local repositories by creating and removing
// a temp directory and use it as clone source.
func newNonExistingLocalRepositoryTestCase(t *testing.T) newGitChartRepositoryTestCase {
	nonExistingLocalRepositoryPath, err := ioutil.TempDir("", "non-existing-local-repository-test-case-")
	if err != nil {
		t.Fatal(err)
	}

	_ = os.RemoveAll(nonExistingLocalRepositoryPath)

	return newGitChartRepositoryTestCase{
		name:       "non-existing local repository",
		repoURL:    nonExistingLocalRepositoryPath,
		shouldFail: true,
	}

}
