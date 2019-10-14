package main

import (
	cs "github.com/otaviof/chart-streams/pkg/chart-streams"
)

func main() {
	config := &cs.Config{}
	s := cs.NewServer(config)
	if err := s.Start(); err != nil {
		panic(err)
	}
}
