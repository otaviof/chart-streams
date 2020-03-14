package chart

import (
	"bytes"
	"crypto/sha256"
	"fmt"
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

// SHA256 returns the sha256 digest of package contents.
func (p *Package) SHA256() (string, error) {
	digest := sha256.New()
	_, err := digest.Write(p.BytesBuffer.Bytes())
	if err != nil {
		return "", err
	}
	sum := digest.Sum(nil)
	return fmt.Sprintf("%x", sum), nil
}
