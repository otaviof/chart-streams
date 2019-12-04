package index

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"sort"
	"time"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/proto/hapi/chart"
	helmrepo "k8s.io/helm/pkg/repo"

	"github.com/otaviof/chart-streams/pkg/chartstreams/repo"
)

// ChartNameVersion tuple with name and version.
type ChartNameVersion struct {
	name    string
	version string
}

// gitChartIndexBuilder creates an "index.yaml" representation from a git repo.
type gitChartIndexBuilder struct {
	gitChartRepo     *repo.GitChartRepo
	versionCommitMap map[ChartNameVersion]repo.CommitInfo
	basePath         string
}

// ErrChartNotFound not-found error.
var ErrChartNotFound = errors.New("chart not found")

// SetBasePath set basePath attribute.
func (g *gitChartIndexBuilder) SetBasePath(basePath string) Builder {
	g.basePath = basePath
	return g
}

var _ Builder = &gitChartIndexBuilder{}

// addChart mark the tuple name-version of a given chart, together with the commit-id data.
func (g gitChartIndexBuilder) addChart(
	name string,
	version string,
	hash plumbing.Hash,
	t time.Time,
) {
	cnv := ChartNameVersion{name: name, version: version}
	g.versionCommitMap[cnv] = repo.CommitInfo{
		Time: t,
		Hash: hash,
	}
}

// getChartVersion returns the version of a chart in the current Git repository work tree.
func getChartVersion(wt *git.Worktree, chartPath string) (string, error) {
	chartDirInfo, err := wt.Filesystem.Lstat(chartPath)
	if err != nil {
		return "", err
	}

	if !chartDirInfo.IsDir() {
		return "", err
	}

	chartYamlPath := wt.Filesystem.Join(chartPath, "Chart.yaml")
	chartYamlFile, err := wt.Filesystem.Open(chartYamlPath)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = chartYamlFile.Close()
	}()

	b, err := ioutil.ReadAll(chartYamlFile)
	if err != nil {
		return "", err
	}

	chartYaml, err := chartutil.UnmarshalChartfile(b)
	if err != nil {
		return "", err
	}

	return chartYaml.GetVersion(), nil
}

// Build index. It will walk by all commits available from the repository, and identify charts and
// versions per commit.
func (g *gitChartIndexBuilder) Build() (*Index, error) {
	commitIter, err := g.gitChartRepo.AllCommits()
	if err != nil {
		return nil, fmt.Errorf("Initialize(): error getting commit iterator: %s", err)
	}
	defer commitIter.Close()

	for {
		c, err := commitIter.Next()
		if err != nil && err != io.EOF {
			break
		}

		w, err := g.gitChartRepo.Worktree()
		if err != nil {
			return nil, err
		}

		checkoutErr := w.Checkout(&git.CheckoutOptions{Hash: c.Hash})
		if checkoutErr != nil {
			return nil, checkoutErr
		}

		chartDirEntries, readDirErr := w.Filesystem.ReadDir(g.basePath)
		if readDirErr != nil {
			return nil, readDirErr
		}

		for _, entry := range chartDirEntries {
			chartName := entry.Name()
			chartPath := w.Filesystem.Join(g.basePath, chartName)
			chartVersion, err := getChartVersion(w, chartPath)
			if err != nil {
				if err != ErrChartNotFound {
					return nil, err
				}
				continue
			}

			g.addChart(chartName, chartVersion, c.Hash, c.Committer.When)
		}
	}

	indexFile := helmrepo.NewIndexFile()

	for _, c := range g.getAllChartsVersions() {
		m := &chart.Metadata{
			Name:       c.name,
			ApiVersion: c.version,
		}
		baseUrl := fmt.Sprintf("/chart/%s/%s", c.name, c.version)
		indexFile.Add(m, "chart.tgz", baseUrl, "deadbeef")
	}

	file := &Index{
		IndexFile:                   indexFile,
		chartNameVersionToCommitMap: g.versionCommitMap,
	}

	return file, nil
}

// getAllChartsVersions returns a sorted slice containing all known charts versions.
func (g *gitChartIndexBuilder) getAllChartsVersions() []ChartNameVersion {
	var allChartsVersions []ChartNameVersion
	for k := range g.versionCommitMap {
		allChartsVersions = append(allChartsVersions, k)
	}
	sort.Sort(byChartNameAndVersion(allChartsVersions))
	return allChartsVersions
}

// NewGitChartIndexBuilder creates an index builder for GitChartRepository.
func NewGitChartIndexBuilder(r *repo.GitChartRepo) Builder {
	return &gitChartIndexBuilder{
		gitChartRepo:     r,
		versionCommitMap: map[ChartNameVersion]repo.CommitInfo{},
	}
}
