package main

import (
	cs "github.com/otaviof/chart-streams/pkg/chart-streams"
)

const defaultDepth = 100
const defaultRepoURL = "https://github.com/helm/charts.git"

func main() {
	config := &cs.Config{Depth: defaultDepth, RepoURL: defaultRepoURL}
	s := cs.NewServer(config)
	if err := s.Start(); err != nil {
		panic(err)
	}
}
