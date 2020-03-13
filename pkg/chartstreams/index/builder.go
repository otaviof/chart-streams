package index

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	log "github.com/sirupsen/logrus"
	helmchart "helm.sh/helm/v3/pkg/chart"
	helmrepo "helm.sh/helm/v3/pkg/repo"
	"sigs.k8s.io/yaml"

	"github.com/otaviof/chart-streams/pkg/chartstreams/repo"
	"github.com/otaviof/chart-streams/pkg/chartstreams/utils"
)

// gitChartIndexBuilder creates an "index.yaml" representation from a git repo.
type gitChartIndexBuilder struct {
	gitChartRepo      *repo.GitChartRepo
	metadataCommitMap map[*helmchart.Metadata]*repo.CommitInfo
	basePath          string
	depth             int
}

var _ Builder = &gitChartIndexBuilder{}

// ErrChartNotFound not-found error.
var ErrChartNotFound = errors.New("chart not found")

// exists check if a given metadata is already present in local cache.
func (g *gitChartIndexBuilder) exists(metadata *helmchart.Metadata) bool {
	for m := range g.metadataCommitMap {
		if metadata.Name == m.Name && metadata.Version == m.Version {
			return true
		}
	}
	return false
}

// semVer semantic versioning representation of given version, revision and commit hash.
func (g *gitChartIndexBuilder) semVer(version, revision string, hash plumbing.Hash) string {
	return fmt.Sprintf("%s-%s-%s", version, revision, hash.String()[:8])
}

// addChart adds informed chart using revision and commit-hash as available versions. When revision
// is HEAD or master, it publishes the original Chart versions. Otherwise, it will create a "semver"
// representation.
func (g gitChartIndexBuilder) addChart(
	metadata *helmchart.Metadata,
	revision string,
	hash plumbing.Hash,
	t time.Time,
) {
	versions := []string{}
	if revision == "HEAD" || revision == "master" {
		if g.exists(metadata) {
			versions = append(versions, g.semVer(metadata.Version, revision, hash))
		} else {
			versions = append(versions, metadata.Version)
		}
	} else {
		versions = append(versions, g.semVer(metadata.Version, revision, hash))
		if _, ok := g.gitChartRepo.Revisions[hash]; ok {
			versions = append(versions, fmt.Sprintf("%s-%s", metadata.Version, revision))
		}
	}

	log.Debugf("Publishing chart '%s' versions '%v'", metadata.Name, versions)
	for _, v := range versions {
		m := &helmchart.Metadata{}
		*m = *metadata
		m.Version = v
		g.metadataCommitMap[m] = &repo.CommitInfo{Time: t, Hash: hash}
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
		// making sure absolute path starts from "/"
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

	head, err := g.gitChartRepo.Head()
	if err != nil {
		return err
	}

	var revision string
	var counter int = 1
	return commitIter.ForEach(func(c *object.Commit) error {
		w, err := g.checkoutWorkTree(c.Hash)
		if err != nil {
			return fmt.Errorf("error checking out working tree: '%s'", err)
		}

		commitID := c.ID().String()
		if currentRevision, ok := g.gitChartRepo.Revisions[c.Hash]; ok {
			revision = currentRevision
		}

		var chartDirPaths []string
		if c.Hash == head.Hash() {
			log.Infof("HEAD (%s): Retrieving all chart directories...", commitID)
			chartDirPaths, err = g.allChartDirs(w)
		} else {
			log.Infof("[%d/%d] Commit (%s/%s): Retrieving modified charts...",
				counter, g.depth, revision, commitID)
			chartDirPaths, err = g.modifiedChartDirs(c)
		}
		// ignoring "object not found errors" (https://github.com/src-d/go-git/issues/1151)
		if err != nil && err != plumbing.ErrObjectNotFound {
			return err
		}
		log.Debugf("Chart directories: '%v'", chartDirPaths)

		// inspecting all directories where is expected to find a helm-chart
		for _, chartDir := range chartDirPaths {
			chartPath := w.Filesystem.Join(g.basePath, chartDir)
			log.Debugf("Inspecting directory '%s' for chart '%s'", chartPath, chartDir)

			if metadata, err := getChartMetadata(w, chartPath); err != nil {
				log.Warnf("error on inspecting chart: '%#v'", err)
				continue
			} else {
				g.addChart(metadata, revision, c.Hash, c.Committer.When)
			}
		}

		counter++
		return nil
	})
}

// Build index. It will walk by all commits available from the repository, and identify charts and
// versions per commit.
func (g *gitChartIndexBuilder) Build() (*Index, error) {
	commitIter, err := g.gitChartRepo.AllCommits()
	if err != nil {
		return nil, fmt.Errorf("error getting commit iterator: %s", err)
	}

	if err := g.walk(commitIter); err != nil {
		return nil, fmt.Errorf("error walking through the commits: %s", err)
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
func NewGitChartIndexBuilder(r *repo.GitChartRepo, basePath string, depth int) Builder {
	return &gitChartIndexBuilder{
		gitChartRepo:      r,
		metadataCommitMap: map[*helmchart.Metadata]*repo.CommitInfo{},
		basePath:          basePath,
		depth:             depth,
	}
}
