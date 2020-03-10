package chartstreams

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/contrib/ginrus"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/yaml"

	"github.com/otaviof/chart-streams/pkg/chartstreams/config"
	"github.com/otaviof/chart-streams/pkg/chartstreams/provider"
)

// Server represents the chartstreams server offering its API. The server puts together the routes,
// and bootstrap steps in order to respond as a valid Helm repository.
type Server struct {
	cfg           *config.Config
	chartProvider provider.ChartProvider
}

// IndexHandler endpoint handler to marshal and return index yaml payload.
func (s *Server) IndexHandler(c *gin.Context) {
	index, err := s.chartProvider.GetIndexFile()
	if err != nil {
		c.AbortWithError(500, err)
	}

	// rendering yaml using k8s approach
	b, err := yaml.Marshal(index)
	if err != nil {
		c.AbortWithError(500, err)
	}

	c.Status(http.StatusOK)
	c.Header("Content-Type", "application/x-yaml; charset=utf-8")
	_, err = c.Writer.Write(b)
	if err != nil {
		c.AbortWithError(500, err)
	}
}

// DirectLinkHandler endpoint handler to directly load a chart tarball payload.
func (s *Server) DirectLinkHandler(c *gin.Context) {
	name := c.Param("name")
	version := c.Param("version")
	version = strings.TrimPrefix(version, "/")

	p, err := s.chartProvider.GetChart(name, version)
	if err != nil {
		c.AbortWithError(500, err)
	}

	c.Data(http.StatusOK, "application/gzip", p.Bytes())
}

// SetupRoutes register routes
func (s *Server) SetupRoutes() *gin.Engine {
	g := gin.New()

	g.Use(ginrus.Ginrus(log.StandardLogger(), time.RFC3339, true))

	g.GET("/index.yaml", s.IndexHandler)
	g.GET("/chart/:name/*version", s.DirectLinkHandler)

	return g
}

// Start executes the boostrap steps in order to start listening on configured address.
func (s *Server) Start() error {
	g := s.SetupRoutes()
	return g.Run(s.cfg.ListenAddr)
}

// NewServer instantiate server.
func NewServer(cfg *config.Config, chartProvider provider.ChartProvider) *Server {
	return &Server{cfg: cfg, chartProvider: chartProvider}
}
