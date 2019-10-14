package chartstreams

import (
	"testing"
)

func TestGit_Clone(t *testing.T) {
	config := &Config{Depth: 1}
	g := NewGit(config)
	g.Clone()
}
