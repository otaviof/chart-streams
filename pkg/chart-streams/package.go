package chartstreams

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"io"
	"os"
)

// Package wraps the Helm Chart archive.
type Package struct {
	content     *bytes.Buffer
	bytesWriter *bufio.Writer
	gzWriter    *gzip.Writer
	tarWriter   *tar.Writer
}

func NewPackage() *Package {
	b := bytes.NewBuffer([]byte{})
	bw := bufio.NewWriter(b)
	gzw := gzip.NewWriter(bw)
	tw := tar.NewWriter(gzw)

	return &Package{
		content:     b,
		bytesWriter: bw,
		gzWriter:    gzw,
		tarWriter:   tw,
	}
}

// Read reads the Helm Chart archive bytes.
func (p *Package) Read(buf []byte) (int, error) {
	return p.content.Read(buf)
}

func (p *Package) Close() {
	p.gzWriter.Close()
	p.tarWriter.Close()
}

// Add adds the contents of the given reader in the archive.
func (p *Package) Add(path string, fileInfo os.FileInfo, r io.Reader) error {
	header, err := tar.FileInfoHeader(fileInfo, fileInfo.Name())
	if err != nil {
		return err
	}

	header.Name = path
	if err := p.tarWriter.WriteHeader(header); err != nil {
		return err
	}

	_, err = io.Copy(p.tarWriter, r)
	return err
}
