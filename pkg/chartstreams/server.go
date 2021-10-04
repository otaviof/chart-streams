package chartstreams

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/contrib/ginrus"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/yaml"

	"github.com/google/go-github/v39/github"
	githubhook "gopkg.in/rjz/githubhook.v0"
)

// Server represents the chartstreams server offering its API. The server puts together the routes,
// and bootstrap steps in order to respond as a valid Helm repository.
type Server struct {
	cfg           *Config
	chartProvider ChartProvider
}

// RootHandler returns a simple string.
func (s *Server) RootHandler(c *gin.Context) {
	c.String(http.StatusOK, "chart-streams")
}

// IndexHandler endpoint handler to marshal and return index yaml payload.
func (s *Server) IndexHandler(c *gin.Context) {
	indexFile := s.chartProvider.GetIndexFile()
	b, err := yaml.Marshal(indexFile)
	if err != nil {
		_ = c.AbortWithError(500, err)
		return
	}

	c.Status(http.StatusOK)
	c.Header("Content-Type", "application/x-yaml; charset=utf-8")
	if _, err = c.Writer.Write(b); err != nil {
		_ = c.AbortWithError(500, err)
		return
	}
}

// DirectLinkHandler endpoint handler to directly load a chart tarball payload.
func (s *Server) DirectLinkHandler(c *gin.Context) {
	name := c.Param("name")
	version := c.Param("version")
	version = strings.TrimPrefix(version, "/")

	log.Debugf("Creating tarball for '%s' version '%s'", name, version)
	p, err := s.chartProvider.GetChart(name, version)
	if err != nil {
		_ = c.AbortWithError(500, err)
		return
	}

	c.Data(http.StatusOK, "application/gzip", p.Bytes())
}

func (s *Server) readEvent(c *gin.Context) (*github.PullRequestEvent, error) {
	evt := github.PullRequestEvent{}
	if s.cfg.GitHubWebhookSecret != "" {
		secret := []byte(s.cfg.GitHubWebhookSecret)
		hook, err := githubhook.Parse(secret, c.Request)
		if err != nil {
			log.Errorf("parsing GitHub webhook: %s", err)
			return nil, &GitHubWebhookParseErr{err}
		}
		payload := hook.Payload
		if err := json.Unmarshal(payload, &evt); err != nil {
			log.Errorf("parsing signed JSON payload: %s", err)
			return nil, &GitHubWebhookPayloadDecodeErr{err}
		}
	} else if err := c.BindJSON(&evt); err != nil {
		log.Errorf("parsing JSON payload: %s", err)
		return nil, &GitHubWebhookPayloadDecodeErr{err}
	}
	return &evt, nil
}

type GitHubWebhookPayloadDecodeErr struct {
	err error
}

func (e *GitHubWebhookPayloadDecodeErr) Error() string {
	return e.err.Error()
}

type GitHubWebhookParseErr struct {
	err error
}

func (e *GitHubWebhookParseErr) Error() string {
	return e.err.Error()
}

func (s *Server) GitHubPullTriggerHandler(c *gin.Context) {
	// The following block describes the behavior whether a GitHub's webhook
	// secret has been informed or not.

	evt, err := s.readEvent(c)
	if err != nil {
		switch err.(type) {
		case *GitHubWebhookParseErr:
			log.Errorf("parsing GitHub webhook: %s", err)
			c.String(http.StatusInternalServerError, "")
			return
		case *GitHubWebhookPayloadDecodeErr:
			log.Errorf("decoding event payload: %s", err)
			c.String(http.StatusBadRequest, "")
			return
		default:
			log.Errorf("unknown error: %s", err)
			c.String(http.StatusInternalServerError, "")
			return
		}
	}

	// do something with `evt`
	log.Debugf("GitHub event handled: %v", evt)
	c.String(http.StatusOK, "")
}

// SetupRoutes register routes
func (s *Server) SetupRoutes() *gin.Engine {
	g := gin.New()

	g.Use(ginrus.Ginrus(log.StandardLogger(), time.RFC3339, true))

	g.GET("/", s.RootHandler)
	g.GET("/index.yaml", s.IndexHandler)
	g.GET("/chart/:name/*version", s.DirectLinkHandler)
	g.POST("/api/webhooks/github", s.GitHubPullTriggerHandler)

	return g
}

// Start executes the boostrap steps in order to start listening on configured address.
func (s *Server) Start() error {
	g := s.SetupRoutes()
	return g.Run(s.cfg.ListenAddr)
}

// NewServer instantiate server.
func NewServer(cfg *Config, chartProvider ChartProvider) *Server {
	return &Server{cfg: cfg, chartProvider: chartProvider}
}
