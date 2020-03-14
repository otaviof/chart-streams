package chart

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/go-git/go-billy/v5"

	"github.com/otaviof/chart-streams/pkg/billyutil"
)

type billyChartBuilder struct {
	Filesystem billy.Filesystem // where chart files will be read from
	ChartPath  string           // path for the chart inside the filesystem
	ChartName  string           // chart name
	CommitTime *time.Time       // commit time
}

var _ Builder = &billyChartBuilder{}

// Build walks the chart directory reading all its artifacts and streaming their contents to a
// gzip'ed tarball to be delivered to the caller. This method doesn't assume anything other than the
// available filesystem.
func (cb *billyChartBuilder) Build() (*Package, error) {
	b := bytes.NewBuffer([]byte{})
	bw := bufio.NewWriter(b)
	gzw := gzip.NewWriter(bw)
	tw := tar.NewWriter(gzw)

	if err := billyutil.Walk(cb.Filesystem, cb.ChartPath, appendToTarWriter(cb, tw)); err != nil {
		return nil, err
	}
	if err := tw.Close(); err != nil {
		return nil, err
	}
	if err := gzw.Close(); err != nil {
		return nil, err
	}
	if err := bw.Flush(); err != nil {
		return nil, err
	}

	return &Package{BytesBuffer: b}, nil
}

// appendToTarWriter returns a walkFn that appends each file into the given tar writer.
func appendToTarWriter(cb *billyChartBuilder, tw *tar.Writer) billyutil.WalkFn {
	return func(fs billy.Filesystem, entityPath string, fileInfo os.FileInfo) error {
		header, err := tar.FileInfoHeader(fileInfo, fileInfo.Name())
		if err != nil {
			return err
		}

		header.Name = fs.Join(cb.ChartName, strings.TrimPrefix(entityPath, cb.ChartPath))
		header.Mode = int64(fileInfo.Mode())
		header.ModTime = *cb.CommitTime

		// when current path is a directory, it is only required to write a header for it
		if fileInfo.IsDir() {
			header.Size = fileInfo.Size()
			return tw.WriteHeader(header)
		}

		// when current path is a regular file, write its header and bytes to the tar writer
		if !fileInfo.Mode().IsRegular() {
			return nil
		}

		f, err := fs.Open(entityPath)
		if err != nil {
			return err
		}
		defer f.Close()

		b, err := ioutil.ReadAll(f)
		if err != nil {
			return err
		}

		header.Size = int64(len(b))
		if err = tw.WriteHeader(header); err != nil {
			return err
		}
		_, err = tw.Write(b)
		return err
	}
}

// NewBillyChartBuilder builds charts with content stored in billy filesystem.
func NewBillyChartBuilder(fs billy.Filesystem, name, path string, t *time.Time) Builder {
	return &billyChartBuilder{
		Filesystem: fs,
		ChartName:  name,
		ChartPath:  path,
		CommitTime: t,
	}
}
