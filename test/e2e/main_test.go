package e2e

import (
	"fmt"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/otaviof/chart-streams/pkg/chartstreams"
	"github.com/otaviof/chart-streams/pkg/chartstreams/config"
	"github.com/otaviof/chart-streams/pkg/chartstreams/provider"
	"github.com/otaviof/chart-streams/test/util"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
}

// startServer run http server on background.
func startServer(cfg *config.Config) error {
	p := provider.NewGitChartProvider(cfg)
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
func waitForServer(cfg *config.Config, sleep time.Duration) error {
	return retry(10, sleep, func() error {
		c := &http.Client{Timeout: 5 * time.Second}
		_, err := c.Get(serverURL(cfg.ListenAddr))
		return err
	})
}

// TestMain main integration test entry point.
func TestMain(t *testing.T) {
	repoDir, err := util.ChartsRepoDir("../..")
	require.NoError(t, err, "on discovering test repo directory dir")

	cfg := &config.Config{
		RepoURL:     fmt.Sprintf("file://%s", repoDir),
		RelativeDir: "/",
		CloneDepth:  1,
		ListenAddr:  "127.0.0.1:8080",
	}

	err = startServer(cfg)
	assert.NoError(t, err)

	err = waitForServer(cfg, 10)
	assert.NoError(t, err)

	it := NewIntegrationTest(cfg)
	it.Run(t)
}
