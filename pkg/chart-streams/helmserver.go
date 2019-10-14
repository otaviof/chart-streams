package chartstreams

import (
	"github.com/gin-gonic/gin"
)

// HelmServer represents the chart-streams server offering its API. The server puts together the routes,
// and bootstrap steps in order to respond as a valid Helm repository.
type HelmServer struct {
	config *Config
}

// Start executes the boostrap steps in order to start listening on configured address. It can return
// errors from "listen" method.
func (s *HelmServer) Start() error {
	return s.listen()
}

// listen on configured address, after adding the route handlers to the framework. It can return
// errors coming from Gin.
func (s *HelmServer) listen() error {
	g := gin.Default()

	g.GET("/index.yaml", IndexHandler)
	g.GET("/chart/:name/*version", DirectLinkHandler)

	return g.Run(s.config.ListenAddr)
}

func (s *HelmServer) initGit() error {
	g := NewGit(s.config)

	if err := g.Clone(); err != nil {
		return err
	}

	return nil
}

// NewServer instantiate a new server instance.
func NewServer(config *Config) *HelmServer {
	return &HelmServer{config: config}
}
