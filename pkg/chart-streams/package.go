package chartstreams

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"io"
	"os"
)

type Package struct {
	b bytes.Buffer
}

func (p *Package) Add(path string, info os.FileInfo, b io.Reader) error {
	mw := bufio.NewWriter(&p.b)

	gzw := gzip.NewWriter(mw)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	header, err := tar.FileInfoHeader(info, info.Name())
	if err != nil {
		return err
	}

	header.Name = path

	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	if _, err := io.Copy(tw, b); err != nil {
		return err
	}

	return nil
}
