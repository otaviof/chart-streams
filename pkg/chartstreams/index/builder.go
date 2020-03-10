package index

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"time"

	log "github.com/sirupsen/logrus"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	helmchart "helm.sh/helm/v3/pkg/chart"
	helmrepo "helm.sh/helm/v3/pkg/repo"
	"sigs.k8s.io/yaml"

	"github.com/otaviof/chart-streams/pkg/chartstreams/repo"
)

// ChartNameVersion tuple with name and version.
type ChartNameVersion struct {
	name    string
	version string
}

// gitChartIndexBuilder creates an "index.yaml" representation from a git repo.
type gitChartIndexBuilder struct {
	gitChartRepo      *repo.GitChartRepo
	metadataCommitMap map[*helmchart.Metadata]repo.CommitInfo
	basePath          string
}

var _ Builder = &gitChartIndexBuilder{}

// ErrChartNotFound not-found error.
var ErrChartNotFound = errors.New("chart not found")

// SetBasePath set basePath attribute.
func (g *gitChartIndexBuilder) SetBasePath(basePath string) Builder {
	g.basePath = basePath
	return g
}

// addChart mark the tuple name-version of a given chart, together with the commit-id data.
func (g gitChartIndexBuilder) addChart(metadata *helmchart.Metadata, hash plumbing.Hash, t time.Time) {
	g.metadataCommitMap[metadata] = repo.CommitInfo{
		Time: t,
		Hash: hash,
	}
}

// getChartMetadata returns the validated helmchart.Metadata payload.
func getChartMetadata(wt *git.Worktree, chartPath string) (*helmchart.Metadata, error) {
	chartDirInfo, err := wt.Filesystem.Lstat(chartPath)
	if err != nil {
		return nil, err
	}
	if !chartDirInfo.IsDir() {
		return nil, fmt.Errorf("path is not a directory: '%s'", chartPath)
	}

	chartYamlPath := wt.Filesystem.Join(chartPath, "Chart.yaml")
	chartYamlFile, err := wt.Filesystem.Open(chartYamlPath)
	if err != nil {
		return nil, err
	}
	defer chartYamlFile.Close()

	b, err := ioutil.ReadAll(chartYamlFile)
	if err != nil {
		return nil, err
	}
	metadata := &helmchart.Metadata{}
	if err = yaml.Unmarshal(b, metadata); err != nil {
		return nil, err
	}
	if metadata.APIVersion == "" {
		metadata.APIVersion = helmchart.APIVersionV1
	}
	return metadata, metadata.Validate()
}

// walk through existing commits, and on each interaction inspect directory tree searching for
// charts. From those the "Chart.yaml" is taken in consideration in order to extract and register
// its metadata.
func (g *gitChartIndexBuilder) walk(commitIter object.CommitIter) error {
	defer commitIter.Close()
	for {
		c, err := commitIter.Next()
		if err != nil && err != io.EOF {
			break
		}

		w, err := g.gitChartRepo.Worktree()
		if err != nil {
			return err
		}
		if err := w.Checkout(&git.CheckoutOptions{Hash: c.Hash}); err != nil {
			return err
		}
		chartDirs, err := w.Filesystem.ReadDir(g.basePath)
		if err != nil {
			return err
		}

		commitID := c.ID().String()
		// inspecting all directories where is expected to find a helm-chart
		for _, entry := range chartDirs {
			chartName := entry.Name()
			chartPath := w.Filesystem.Join(g.basePath, chartName)
			log.Infof("Inspecting directory '%s' for chart '%s' on commit-id '%s'",
				chartPath, chartName, commitID)

			metadata, err := getChartMetadata(w, chartPath)
			if err != nil {
				if err != ErrChartNotFound {
					return err
				}
				continue
			}

			g.addChart(metadata, c.Hash, c.Committer.When)
		}
	}
	return nil
}

// Build index. It will walk by all commits available from the repository, and identify charts and
// versions per commit.
func (g *gitChartIndexBuilder) Build() (*Index, error) {
	commitIter, err := g.gitChartRepo.AllCommits()
	if err != nil {
		return nil, fmt.Errorf("Initialize(): error getting commit iterator: %s", err)
	}

	if err := g.walk(commitIter); err != nil {
		return nil, err
	}

	indexFile := helmrepo.NewIndexFile()
	for metadata := range g.metadataCommitMap {
		baseUrl := fmt.Sprintf("/chart/%s/%s", metadata.Name, metadata.Version)
		log.Infof("Adding '%s/%s' (%s) to index file", metadata.Name, metadata.Version, baseUrl)
		indexFile.Add(metadata, "chart.tgz", baseUrl, "deadbeef")
	}
	indexFile.SortEntries()

	return &Index{
		IndexFile:                indexFile,
		chartMetadataToCommitMap: g.metadataCommitMap,
	}, nil
}

// NewGitChartIndexBuilder creates an index builder for GitChartRepository.
func NewGitChartIndexBuilder(r *repo.GitChartRepo) Builder {
	return &gitChartIndexBuilder{
		gitChartRepo:      r,
		metadataCommitMap: map[*helmchart.Metadata]repo.CommitInfo{},
	}
}
