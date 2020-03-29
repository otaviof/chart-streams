package e2e

import (
	"io/ioutil"
	"testing"

	"github.com/otaviof/chart-streams/pkg/chartstreams"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"helm.sh/helm/v3/cmd/helm/search"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
	helmrepo "helm.sh/helm/v3/pkg/repo"
	"sigs.k8s.io/yaml"
)

// IntegrationTest comprises all the steps needed to test chart-streams features. Using parts of
// upstream Helm pkg and cmd to simulate end-users.
type IntegrationTest struct {
	cfg           *chartstreams.Config
	helmChartRepo *helmrepo.ChartRepository
	indexFile     *helmrepo.IndexFile
}

// repoName helm-chart repository name.
const repoName = "chartstreams"

// helmRepoAddTest test "helm repo add".
func (i *IntegrationTest) helmRepoAddTest(t *testing.T) {
	repoCfg := &helmrepo.Entry{
		Name: repoName,
		URL:  serverURL(i.cfg.ListenAddr),
	}

	t.Logf("Adding Helm Charts repository '%s'", i.cfg.ListenAddr)
	var err error
	i.helmChartRepo, err = helmrepo.NewChartRepository(repoCfg, getter.All(cli.New()))
	require.NoError(t, err, "on adding charts repository")
}

// helmRepoUpdateTest test "helm repo update".
func (i *IntegrationTest) helmRepoUpdateTest(t *testing.T) {
	t.Log("Downloading index.yaml from repository")
	indexFilePath, err := i.helmChartRepo.DownloadIndexFile()
	require.NoError(t, err, "on downloading index.yaml file")

	b, err := ioutil.ReadFile(indexFilePath)
	require.NoError(t, err, "on reading index.yaml payload")

	t.Logf("Executing unmarshal on contents of '%s' (%d bytes)", indexFilePath, len(b))
	err = yaml.Unmarshal(b, i.indexFile)
	require.NoError(t, err, "on unmarshal index.yaml payload")
}

// helmSearchRepoTest test "helm search".
func (i *IntegrationTest) helmSearchRepoTest(t *testing.T) {
	searchIndex := search.NewIndex()
	searchIndex.AddRepo(repoName, i.indexFile, true)
	all := searchIndex.All()
	for _, r := range all {
		t.Logf("result:  '%#v'", r)
	}
	assert.True(t, len(all) > 0, "on checking amount of available results")
}

// Run all integration tests.
func (i *IntegrationTest) Run(t *testing.T) {
	t.Run("Helm: Adding Chart-Streams as repository", i.helmRepoAddTest)
	t.Run("Helm: Updating Chart-Streams repository", i.helmRepoUpdateTest)
	t.Run("Helm: Searching on Chart-Streams repository", i.helmSearchRepoTest)
}

// NewIntegrationTest instantiate tests.
func NewIntegrationTest(cfg *chartstreams.Config) *IntegrationTest {
	return &IntegrationTest{cfg: cfg, indexFile: &helmrepo.IndexFile{}}
}
