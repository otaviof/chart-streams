package index

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	helmchart "helm.sh/helm/v3/pkg/chart"
	helmrepo "helm.sh/helm/v3/pkg/repo"
	"sigs.k8s.io/yaml"

	"github.com/otaviof/chart-streams/pkg/chartstreams/repo"
	"github.com/otaviof/chart-streams/pkg/chartstreams/utils"
)

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

// modifiedChartDirs inspect commit object in order to select only the chart directories that have
// been affected in the commit.
func (g *gitChartIndexBuilder) modifiedChartDirs(c *object.Commit) ([]string, error) {
	stats, err := c.Stats()
	if err != nil {
		return nil, err
	}

	var absBasePath string
	if strings.HasPrefix(g.basePath, "/") {
		absBasePath = g.basePath
	} else {
		absBasePath = fmt.Sprintf("/%s", g.basePath)
	}
	if absBasePath != "/" && !strings.HasSuffix(absBasePath, "/") {
		absBasePath = fmt.Sprintf("%s/", absBasePath)
	}

	var chartDirPaths []string
	for _, stat := range stats {
		// skipping entries that haven't changed
		if stat.Addition == 0 && stat.Deletion == 0 {
			continue
		}

		absPath := fmt.Sprintf("/%s", stat.Name)

		// skipping entries that are not in pre-defined base path
		if !strings.HasPrefix(absPath, absBasePath) {
			continue
		}

		// discoverying the relative path by removing base path
		relativePath := strings.TrimPrefix(absPath, absBasePath)
		// splitting into path elements, in order to pick the chart root directory
		elements := strings.Split(relativePath, string(os.PathSeparator))
		if len(elements) <= 1 {
			continue
		}
		chartRootDir := elements[0]
		if !utils.ContainsStringSlice(chartDirPaths, chartRootDir) {
			chartDirPaths = append(chartDirPaths, elements[0])
		}
	}
	return chartDirPaths, nil
}

// allChartDirs inspect the base location to pick all directories, which are expected to be chart
// directories.
func (g *gitChartIndexBuilder) allChartDirs(w *git.Worktree) ([]string, error) {
	chartDirs, err := w.Filesystem.ReadDir(g.basePath)
	if err != nil {
		return nil, err
	}
	var charDirPaths []string
	for _, dir := range chartDirs {
		if !dir.IsDir() {
			continue
		}
		charDirPaths = append(charDirPaths, dir.Name())
	}
	return charDirPaths, nil
}

// checkoutWorkTree instantiate and checkout the commit hash in a working tree.
func (g *gitChartIndexBuilder) checkoutWorkTree(hash plumbing.Hash) (*git.Worktree, error) {
	w, err := g.gitChartRepo.Worktree()
	if err != nil {
		return nil, err
	}
	if err := w.Checkout(&git.CheckoutOptions{Hash: hash}); err != nil {
		return nil, err
	}
	return w, nil
}

// walk through existing commits, and on each interaction inspect directory tree searching for
// charts. From those the "Chart.yaml" is taken in consideration in order to extract and register
// its metadata.
func (g *gitChartIndexBuilder) walk(commitIter object.CommitIter) error {
	defer commitIter.Close()

	counter := 0
	for {
		c, err := commitIter.Next()
		if c == nil || (err != nil && err == io.EOF) {
			break
		}

		w, err := g.checkoutWorkTree(c.Hash)
		if err != nil {
			return err
		}

		var chartDirPaths []string
		if counter == 0 {
			log.Info("HEAD: Retrieving all chart directories...")
			chartDirPaths, err = g.allChartDirs(w)
		} else {
			log.Info("Commit: Retrieving modified chart directories...")
			chartDirPaths, err = g.modifiedChartDirs(c)
		}
		// ignoring "object not found errors" (https://github.com/src-d/go-git/issues/1151)
		if err != nil && err != plumbing.ErrObjectNotFound {
			return err
		}
		log.Infof("Chart directories: '%v'", chartDirPaths)

		// inspecting all directories where is expected to find a helm-chart
		for _, chartDir := range chartDirPaths {
			chartPath := w.Filesystem.Join(g.basePath, chartDir)
			log.Infof("Inspecting directory '%s' for chart '%s' on commit-id '%s'",
				chartPath, chartDir, c.ID().String())

			metadata, err := getChartMetadata(w, chartPath)
			if err != nil {
				if err != ErrChartNotFound {
					return err
				}
				log.Errorf("Error on inspecting chart '%s': '%#v'", chartPath, err)
				continue
			}

			g.addChart(metadata, c.Hash, c.Committer.When)
		}

		counter++
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
func NewGitChartIndexBuilder(r *repo.GitChartRepo, basePath string) Builder {
	return &gitChartIndexBuilder{
		gitChartRepo:      r,
		metadataCommitMap: map[*helmchart.Metadata]repo.CommitInfo{},
		basePath:          basePath,
	}
}
