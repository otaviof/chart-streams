package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/otaviof/chart-streams/pkg/chartstreams"
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
	flags.String("relative-dir", "stable", "Relative charts directory in repository")
	flags.String("listen-addr", ":8080", "Address to listen")
	flags.String("working-dir", "/var/lib/chart-streams", "Git repository working directory")
	flags.String("log-level", "info", "Log verbosity level (error, warn, info, debug, trace)")
	flags.Bool("force-clone", false, "destroys working-dir and clones the repository")
	flags.String("github-webhook-secret", "", "GitHub's webhook secret for this repository")

	rootCmd.AddCommand(serveCmd)
	bindViperFlags(flags)
}

// runServeCmd execute chartstreams server.
func runServeCmd(cmd *cobra.Command, args []string) {
	chartstreams.SetLogLevel(viper.GetString("log-level"))

	cfg := &chartstreams.Config{
		RepoURL:             viper.GetString("repo-url"),
		CloneDepth:          viper.GetInt("clone-depth"),
		ListenAddr:          viper.GetString("listen-addr"),
		RelativeDir:         viper.GetString("relative-dir"),
		ForceClone:          viper.GetBool("force-clone"),
		GitHubWebhookSecret: viper.GetString("github-webhook-secret"),
	}

	log.Printf("Starting server with config: '%#v'", cfg)

	p := chartstreams.NewGitChartProvider(cfg, viper.GetString("working-dir"))
	if err := p.Initialize(); err != nil {
		log.Fatalf("Error during chart provider initialization: '%q'", err)
	}

	s := chartstreams.NewServer(cfg, p)
	if err := s.Start(); err != nil {
		log.Fatalf("Error during server start: '%q'", err)
	}
}
