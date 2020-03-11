package index

import (
	"fmt"
	"io/ioutil"
	"testing"

	log "github.com/sirupsen/logrus"
	"gopkg.in/src-d/go-git.v4/plumbing"

	"github.com/otaviof/chart-streams/pkg/chartstreams/config"
	"github.com/otaviof/chart-streams/pkg/chartstreams/repo"
	"github.com/otaviof/chart-streams/test/util"
)

type TestCase struct {
	basePath                    string
	name                        string
	repoURL                     string
	hash                        plumbing.Hash
	shouldFail                  bool
	depth                       uint
	expectedIndexFileEntryCount uint
	expectedChartVersionCount   uint
}

func newGitChartIndexBuilderTestCase(
	depth uint,
	expectedIndexFileEntryCount uint,
	expectedChartVersionCount uint,
) *TestCase {
	name := fmt.Sprintf("depth %d expectedIndexFileEntryCount %d expectedChartVersionCount %d",
		depth, expectedIndexFileEntryCount, expectedChartVersionCount)

	helmRepoDir, _ := util.ChartsRepoDir("../../..")
	helmRepoURL := fmt.Sprintf("file://%s", helmRepoDir)

	return &TestCase{
		name:                        name,
		repoURL:                     helmRepoURL,
		basePath:                    "/",
		hash:                        plumbing.NewHash("d093c4dcc9e2c6aeeb9e81d4da428328c8d4a714"),
		shouldFail:                  false,
		depth:                       depth,
		expectedIndexFileEntryCount: expectedIndexFileEntryCount,
		expectedChartVersionCount:   expectedChartVersionCount,
	}
}

func TestNewGitChartIndexBuilder(t *testing.T) {
	log.SetOutput(ioutil.Discard)

	tests := []*TestCase{
		newGitChartIndexBuilderTestCase(1, 1, 1),
		newGitChartIndexBuilderTestCase(2, 2, 2),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				RepoURL:    tt.repoURL,
				CloneDepth: int(tt.depth),
			}
			r, err := repo.NewGitChartRepo(cfg)
			if err != nil {
				t.Fatal(err)
			}

			i, err := NewGitChartIndexBuilder(r, tt.basePath).Build()
			if err != nil {
				t.Fatal(err)
			}

			gotIndexFileEntryCount := len(i.IndexFile.Entries)
			if gotIndexFileEntryCount < int(tt.expectedIndexFileEntryCount) {
				t.Errorf("index file should have %d entries, found %d",
					tt.expectedIndexFileEntryCount, gotIndexFileEntryCount)
			}

			var gotChartVersionCount int
			for _, chartVersions := range i.IndexFile.Entries {
				gotChartVersionCount = gotChartVersionCount + len(chartVersions)
			}
			if gotChartVersionCount < int(tt.expectedChartVersionCount) {
				t.Errorf("index file should have %d chart versions, found %d",
					tt.expectedChartVersionCount, gotChartVersionCount)
			}
		})
	}
}
