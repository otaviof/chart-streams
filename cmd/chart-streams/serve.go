package main

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	cs "github.com/otaviof/chart-streams/pkg/chart-streams"
)

// serveCmd sub-command to represent the server.
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Execute Helm Repository server",
	Run:   runServeCmd,
}

// init initialize the command-line flags and interpolation with environment.
func init() {
	flags := serveCmd.PersistentFlags()

	flags.Int("clone-depth", 1, "Git clone depth")
	flags.String("repo-url", "https://github.com/helm/charts.git", "Helm Charts Git repository URL")

	rootCmd.AddCommand(serveCmd)
	bindViperFlags(flags)
}

// runServeCmd execute chart-streams server.
func runServeCmd(cmd *cobra.Command, args []string) {
	config := &cs.Config{
		Depth:   viper.GetInt("clone-depth"),
		RepoURL: viper.GetString("repo-url"),
	}

	s := cs.NewServer(config)
	if err := s.Start(); err != nil {
		panic(err)
	}
}
