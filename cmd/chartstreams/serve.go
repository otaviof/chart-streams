package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/otaviof/chart-streams/pkg/chartstreams"
	"github.com/otaviof/chart-streams/pkg/chartstreams/config"
	"github.com/otaviof/chart-streams/pkg/chartstreams/provider"
)

// serveCmd sub-command to represent the server.
var serveCmd = &cobra.Command{
	Use:   "serve",
	Run:   runServeCmd,
	Short: "Execute Helm Repository server",
	Long:  "Run the Helm-Charts server after cloning and preparing Git repository",
}

// init initialize the command-line flags and interpolation with environment.
func init() {
	flags := serveCmd.PersistentFlags()

	flags.Int("clone-depth", 1, "Git clone depth")
	flags.String("repo-url", "https://github.com/helm/charts.git", "Helm Charts Git repository URL")
	flags.String("listen-addr", ":8080", "Address to listen")

	rootCmd.AddCommand(serveCmd)
	bindViperFlags(flags)
}

// runServeCmd execute chartstreams server.
func runServeCmd(cmd *cobra.Command, args []string) {
	cfg := &config.Config{
		RepoURL:     viper.GetString("repo-url"),
		CloneDepth:  viper.GetInt("clone-depth"),
		ListenAddr:  viper.GetString("listen-addr"),
		RelativeDir: "stable",
	}

	log.Printf("Starting server with config: '%#v'", cfg)

	p := provider.NewGitChartProvider(cfg)
	if err := p.Initialize(); err != nil {
		log.Fatalf("Error during chart provider initialization: '%q'", err)
	}

	s := chartstreams.NewServer(cfg, p)
	if err := s.Start(); err != nil {
		log.Fatalf("Error during server start: '%q'", err)
	}
}
