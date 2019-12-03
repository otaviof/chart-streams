package index

import (
	"fmt"
	"testing"

	"gopkg.in/src-d/go-git.v4/plumbing"

	"github.com/otaviof/chart-streams/pkg/chartstreams/config"
	"github.com/otaviof/chart-streams/pkg/chartstreams/repo"
)

const helmChartsReferenceURL = "https://github.com/helm/charts.git"
const helmChartsReferenceBasePath = "stable"

type gitChartIndexBuilderTestCase struct {
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
) *gitChartIndexBuilderTestCase {
	name := fmt.Sprintf("depth %d expectedIndexFileEntryCount %d expectedChartVersionCount %d",
		depth, expectedIndexFileEntryCount, expectedChartVersionCount)
	return &gitChartIndexBuilderTestCase{
		basePath:                    helmChartsReferenceBasePath,
		name:                        name,
		repoURL:                     helmChartsReferenceURL,
		hash:                        plumbing.NewHash("d093c4dcc9e2c6aeeb9e81d4da428328c8d4a714"),
		shouldFail:                  false,
		depth:                       depth,
		expectedIndexFileEntryCount: expectedIndexFileEntryCount,
		expectedChartVersionCount:   expectedChartVersionCount,
	}
}

func TestNewGitChartIndexBuilder(t *testing.T) {
	tests := []*gitChartIndexBuilderTestCase{
		newGitChartIndexBuilderTestCase(1, 280, 280),
		newGitChartIndexBuilderTestCase(5, 280, 284),
		newGitChartIndexBuilderTestCase(50, 280, 328),
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

			i, err := NewGitChartIndexBuilder(r).SetBasePath(tt.basePath).Build()
			if err != nil {
				t.Fatal(err)
			}

			gotIndexFileEntryCount := len(i.IndexFile.Entries)
			if gotIndexFileEntryCount > int(tt.expectedIndexFileEntryCount) {
				t.Errorf("index file should have %d entries, found %d",
					tt.expectedIndexFileEntryCount, gotIndexFileEntryCount)
			}

			var gotChartVersionCount int
			for _, chartVersions := range i.IndexFile.Entries {
				gotChartVersionCount = gotChartVersionCount + len(chartVersions)
			}
			if gotChartVersionCount != int(tt.expectedChartVersionCount) {
				t.Errorf("index file should have %d chart versions, found %d",
					tt.expectedChartVersionCount, gotChartVersionCount)
			}
		})
	}
}
