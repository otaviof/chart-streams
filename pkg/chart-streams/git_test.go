package chartstreams

import (
	"testing"
)

func TestGit_Clone(t *testing.T) {
	config := &Config{}
	g := NewGit(config)
	g.Clone()
}
