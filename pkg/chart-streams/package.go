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
	content bytes.Buffer
}

// Read reads the Helm Chart archive bytes.
func (p *Package) Read(buf []byte) (int, error) {
	return p.content.Read(buf)
}

// Add adds the contents of the given reader in the archive.
func (p *Package) Add(path string, fileInfo os.FileInfo, r io.Reader) error {
	bw := bufio.NewWriter(&p.content)

	gzw := gzip.NewWriter(bw)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	header, err := tar.FileInfoHeader(fileInfo, fileInfo.Name())
	if err != nil {
		return err
	}

	header.Name = path

	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	if _, err := io.Copy(tw, r); err != nil {
		return err
	}

	return nil
}
