package chartstreams

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
	"os"
)

// Package wraps the Helm Chart archive.
type Package struct {
	bytesBuffer *bytes.Buffer
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
		bytesBuffer: b,
		bytesWriter: bw,
		gzWriter:    gzw,
		tarWriter:   tw,
	}
}

// Read reads the Helm Chart archive bytes.
func (p *Package) Read(buf []byte) (int, error) {
	p.bytesWriter.Flush()
	return p.bytesBuffer.Read(buf)
}

func (p *Package) Close() {
	p.tarWriter.Flush()
	p.tarWriter.Close()

	p.gzWriter.Flush()
	p.gzWriter.Close()
}

// Add adds the contents of the given reader in the archive.
func (p *Package) Add(path string, fileInfo os.FileInfo, r io.Reader) error {

	header, err := tar.FileInfoHeader(fileInfo, fileInfo.Name())
	if err != nil {
		return err
	}

	header.Name = path

	// Write down the header since there's no reader available.
	if r == nil {
		header.Mode = 0755
		return p.tarWriter.WriteHeader(header)
	}

	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	header.Size = int64(len(b))
	header.Mode = 0644

	if err := p.tarWriter.WriteHeader(header); err != nil {
		return err
	}

	if _, err = p.tarWriter.Write(b); err != nil {
		return err
	}

	return nil
}
