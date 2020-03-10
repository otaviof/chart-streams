package chart

import (
	"bytes"
)

// Package wraps the Helm Chart archive, which consists of chart files in a given directory packaged
// as a tarball.
type Package struct {
	BytesBuffer *bytes.Buffer // package content
}

// Bytes return buffer bytes.
func (p *Package) Bytes() []byte {
	return p.BytesBuffer.Bytes()
}
