package main

import (
	cs "github.com/otaviof/chart-streams/pkg/chart-streams"
)

const defaultDepth = 100

func main() {
	config := &cs.Config{Depth: defaultDepth}
	s := cs.NewServer(config)
	if err := s.Start(); err != nil {
		panic(err)
	}
}
