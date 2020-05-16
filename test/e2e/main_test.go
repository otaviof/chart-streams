package e2e

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/otaviof/chart-streams/pkg/chartstreams"
	"github.com/otaviof/chart-streams/test/util"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
}

// startServer run http server on background.
func startServer(cfg *chartstreams.Config) error {
	tempDir, err := ioutil.TempDir("", "chart-streams")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	p := chartstreams.NewGitChartProvider(cfg, tempDir)
	if err := p.Initialize(); err != nil {
		return err
	}
	s := chartstreams.NewServer(cfg, p)
	go func() {
		err := s.Start()
		log.Fatalf("Error running server: '%q'", err)
	}()
	return nil
}

// retry the informed method a few times, with sleep between attempts.
func retry(attempts int, sleep time.Duration, fn func() error) error {
	var err error
	for i := attempts; i > 0; i-- {
		if err = fn(); err == nil {
			break
		}
		time.Sleep(sleep)
	}
	return err
}

func serverURL(addr string) string {
	return fmt.Sprintf("http://%s/", addr)
}

// waitForServer will try to reach server's root, and if fail, sleep and retry for a certain amount
// of times.
func waitForServer(cfg *chartstreams.Config, sleep time.Duration) error {
	return retry(10, sleep, func() error {
		c := &http.Client{Timeout: 5 * time.Second}
		_, err := c.Get(serverURL(cfg.ListenAddr))
		return err
	})
}

// TestMain main integration test entry point	.
func TestMain(t *testing.T) {
	repoDir, err := util.ChartsRepoDir("../..")
	require.NoError(t, err, "on discovering test repo directory dir")
	t.Logf("Charts repository directory: '%s'", repoDir)

	cfg := &chartstreams.Config{
		RepoURL:     fmt.Sprintf("file://%s", repoDir),
		RelativeDir: "/",
		CloneDepth:  0,
		ListenAddr:  "127.0.0.1:8080",
	}

	err = startServer(cfg)
	assert.NoError(t, err)

	err = waitForServer(cfg, 10)
	assert.NoError(t, err)

	it := NewIntegrationTest(cfg)
	it.Run(t)
}
