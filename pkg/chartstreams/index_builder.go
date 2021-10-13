package chartstreams

import (
	"fmt"
	"os"
	"strings"

	"helm.sh/helm/v3/pkg/chart/loader"
	helmrepo "helm.sh/helm/v3/pkg/repo"

	git "github.com/libgit2/git2go/v31"
	log "github.com/sirupsen/logrus"
	helmchart "helm.sh/helm/v3/pkg/chart"
)

// commitInfoMap map to relate helm chart metadata with git repository metadata
type commitInfoMap map[*helmchart.Metadata]*CommitInfo

// Index represent a chart index of some sort.
type Index struct {
	IndexFile *helmrepo.IndexFile // repository index-file
}

// IndexBuilder navigate the repository commits in order to discover charts, and store the
// relationship between chart version and commit-id.
type IndexBuilder struct {
	cfg            *Config       // global configuration
	g              *GitRepo      // git repository instance
	metadataCommit commitInfoMap // map of chart and git repository metadata
}

// GetChartCommitInfo return the commit information based on chart name and version, or nil in case
// of not found.
func (i *IndexBuilder) GetChartCommitInfo(name, version string) *CommitInfo {
	for metadata, commitInfo := range i.metadataCommit {
		if metadata.Name == name && metadata.Version == version {
			return commitInfo
		}
	}
	return nil
}

// listAllChartDirs returns a list of all charts found on base chart directory.
func (i *IndexBuilder) listAllChartDirs(c *git.Commit) ([]string, error) {
	tree, err := c.Tree()
	if err != nil {
		return nil, fmt.Errorf("getting commit's tree: %w", err)
	}
	defer tree.Free()

	var chartDirs []string
	tree.Walk(func(curPath string, te *git.TreeEntry) int {
		if te.Filemode != git.FilemodeTree {
			return 0
		}
		// since this method is specific to chart dirs, it is fine to skip the
		// entry if it is in the root.
		if curPath == "" {
			return 0
		}
		parts := strings.Split(curPath, "/")
		if len(parts) > 1 || len(parts) == 0 {
			return 0
		}
		chartDirs = append(chartDirs, parts[0])
		return 0
	})
	return chartDirs, nil
}

// relativeDir prepare relative directory to be compared as string prefix.
func (i *IndexBuilder) relativeDir() string {
	relativeDir := i.cfg.RelativeDir
	if relativeDir == "/" {
		return relativeDir
	}
	return fmt.Sprintf("/%s/", relativeDir)
}

// listModifiedChartDirs inspect git commit to discover which charts have changed, return the list
// of modified charts.
func (i *IndexBuilder) listModifiedChartDirs(c *git.Commit) ([]string, error) {
	modified, err := i.g.ModifiedFiles(c)
	if err != nil {
		return nil, fmt.Errorf("obtaining modified files for commit-id %q: %w", c.Id(), err)
	}

	relativeDir := i.relativeDir()
	chartDirs := []string{}
	for _, entry := range modified {
		absEntryPath := fmt.Sprintf("/%s", entry)
		if !strings.HasPrefix(absEntryPath, relativeDir) {
			continue
		}

		chartRelativePath := strings.TrimPrefix(absEntryPath, relativeDir)
		pathElements := strings.Split(chartRelativePath, string(os.PathSeparator))
		if len(pathElements) <= 1 {
			continue
		}

		chartRootDir := pathElements[0]
		if !ContainsStringSlice(chartDirs, chartRootDir) {
			chartDirs = append(chartDirs, pathElements[0])
		}
	}
	return chartDirs, nil
}

// semVer creates a semantic version string out of chart and repository metadata.
func (i *IndexBuilder) semVer(metadata *helmchart.Metadata, branch, shortID string) string {
	if shortID == "" {
		return fmt.Sprintf("%s-%s", metadata.Version, branch)
	}
	return fmt.Sprintf("%s-%s-%s", metadata.Version, branch, shortID)
}

// exists checks if a given chart is already registered.
func (i *IndexBuilder) exists(metadata *helmchart.Metadata) bool {
	for m := range i.metadataCommit {
		if metadata.Name == m.Name && metadata.Version == m.Version {
			return true
		}
	}
	return false
}

// register chart in local cache.
func (i *IndexBuilder) register(
	metadata *helmchart.Metadata,
	revision string,
	c *git.Commit,
	head bool,
) error {
	shortID, err := c.ShortId()
	if err != nil {
		return err
	}

	semVerWithCommit := i.semVer(metadata, revision, shortID)
	versions := []string{}
	if revision == "master" {
		if i.exists(metadata) {
			versions = append(versions, semVerWithCommit)
		} else {
			versions = append(versions, metadata.Version)
		}
	} else {
		if head {
			versions = append(versions, i.semVer(metadata, revision, ""))
		} else {
			versions = append(versions, semVerWithCommit)
		}
	}

	commitInfo := &CommitInfo{
		Time:     &c.Author().When,
		ID:       c.Id().String(),
		Revision: revision,
		Digest:   "fixme",
	}
	for _, v := range versions {
		m := &helmchart.Metadata{}
		*m = *metadata
		m.Version = v
		i.metadataCommit[m] = commitInfo
	}
	return nil
}

// absChartPath returns the absolute chart path from the repository root.
func (i *IndexBuilder) absChartPath(chartPath string) string {
	if i.cfg.RelativeDir != "" {
		chartPath = i.cfg.RelativeDir + "/" + chartPath
	}
	chartPath = strings.TrimPrefix(chartPath, "//")
	return chartPath
}

// inspectDirs loop chart directories, loading charts and registering them.
func (i *IndexBuilder) inspectDirs(dirs []string, revision string, c *git.Commit, head bool) error {
	for _, dir := range dirs {
		chartPath := i.absChartPath(dir)
		files, err := i.g.GetFilesFromCommit(c, chartPath)
		if err != nil {
			return fmt.Errorf("getting files from commit '%s': %w", c.Id(), err)
		}
		if len(files) == 0 {
			log.Infof("no files found in '%s' for commit '%s'", chartPath, c.Id())
			continue
		}
		chart, err := loader.LoadFiles(files)
		if err != nil {
			log.Warnf("error loading chart with files from '%s': '%s'", chartPath, err)
			continue
		}
		if err = chart.Validate(); err != nil {
			log.Warnf("error validating chart: '%s'", err)
			continue
		}
		log.Infof("Found chart '%s' version '%s'", chart.Metadata.Name, chart.Metadata.Version)

		// TODO(isutton): Return the first error returned in the loop.
		_ = i.register(chart.Metadata, revision, c, head)

		// NOTE(isutton): I was tempted to change the behavior of this method
		// with the code that follows, but changed my mind; this will stay
		// for my future self's convenience.
		//
		// err = i.register(chart.Metadata, revision, c, head)
		// if err != nil {
		// 	return err
		// }
	}
	return nil
}

// walk through git commits in order to identify which charts have changed, or to load all charts
// from repository in case of master's HEAD.
func (i *IndexBuilder) walk() error {
	return i.g.CommitIter(func(branch string, c *git.Commit, head bool) error {
		log.Infof("Inspecting commit-id '%s/%s'", branch, c.Id().String())
		var chartDirs []string
		var err error
		if head && branch == "master" {
			log.Infof("HEAD: Retrieving all chart directories...")
			chartDirs, err = i.listAllChartDirs(c)
		} else {
			chartDirs, err = i.listModifiedChartDirs(c)
		}
		if err != nil {
			return err
		}
		log.Debugf("Chart directories: '%v'", chartDirs)
		return i.inspectDirs(chartDirs, branch, c, head)
	})
}

// Build create instance of index-file by inspecting commits.
func (i *IndexBuilder) Build() (*helmrepo.IndexFile, error) {
	if err := i.walk(); err != nil {
		return nil, fmt.Errorf("walking the repository: %w", err)
	}

	indexFile := helmrepo.NewIndexFile()
	for metadata, commitInfo := range i.metadataCommit {
		baseUrl := fmt.Sprintf("/chart/%s/%s", metadata.Name, metadata.Version)
		log.Infof("Adding '%s/%s' (%s) to index file", metadata.Name, metadata.Version, baseUrl)
		// FIXME(isutton): decide what to do with the error returned by MustAdd.
		_ = indexFile.MustAdd(metadata, "chart.tgz", baseUrl, commitInfo.Digest)

		// NOTE(isutton): I'm not convinced that just throwing the error up
		// is the right approach, as this would hang the builder at any
		// commit that might produce an error. As there might be different
		// approaches this change doesn't belong to this changeset and is
		// kept as reference.
		// err := indexFile.MustAdd(metadata, "chart.tgz", baseUrl, commitInfo.Digest)
		// if err != nil {
		// 	return nil, err
		// }
	}
	indexFile.SortEntries()
	return indexFile, nil
}

// NewIndexBuilder instantiate a Helm repository index-file builder.
func NewIndexBuilder(cfg *Config, g *GitRepo) *IndexBuilder {
	return &IndexBuilder{
		cfg:            cfg,
		g:              g,
		metadataCommit: make(commitInfoMap),
	}
}
