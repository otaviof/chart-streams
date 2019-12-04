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

	"gopkg.in/src-d/go-billy.v4"

	"github.com/otaviof/chart-streams/pkg/billyutil"
)

type billyChartBuilder struct {
	Filesystem billy.Filesystem
	ChartPath  *string
	ChartName  *string
	CommitTime *time.Time
}

var _ Builder = &billyChartBuilder{}

// NewBillyChartBuilder builds charts with content stored in billy filesystem.
func NewBillyChartBuilder(fs billy.Filesystem) Builder {
	return &billyChartBuilder{Filesystem: fs}
}

func (cb *billyChartBuilder) SetChartName(n string) Builder {
	cb.ChartName = &n
	return cb
}

func (cb *billyChartBuilder) SetChartPath(p string) Builder {
	cb.ChartPath = &p
	return cb
}

func (cb *billyChartBuilder) SetCommitTime(t time.Time) Builder {
	cb.CommitTime = &t
	return cb
}

// Build walks the chart directory reading all its artifacts and streaming their contents to a
// gzip'ed tarball to be delivered to the caller. This method doesn't assume anything other than the
// available filesystem.
func (cb *billyChartBuilder) Build() (*Package, error) {
	b := bytes.NewBuffer([]byte{})
	bw := bufio.NewWriter(b)
	gzw := gzip.NewWriter(bw)
	tw := tar.NewWriter(gzw)

	walkErr := billyutil.Walk(cb.Filesystem, *cb.ChartPath, appendFileToTarWriter(cb, tw))

	if walkErr != nil {
		return nil, walkErr
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

	p := &Package{bytesBuffer: b}

	return p, nil
}

// appendFileToTarWriter returns a walkFn that appends each file into the given tar writer.
func appendFileToTarWriter(cb *billyChartBuilder, tw *tar.Writer) billyutil.WalkFn {
	return func(fs billy.Filesystem, path string, fileInfo os.FileInfo) error {
		header, err := tar.FileInfoHeader(fileInfo, fileInfo.Name())
		if err != nil {
			return err
		}

		// populate common header fields
		header.Name = fs.Join(*cb.ChartName, strings.TrimPrefix(path, *cb.ChartPath))
		header.Mode = int64(fileInfo.Mode())
		header.ModTime = *cb.CommitTime

		// if the current path is a directory, it is only required to write a header for it
		if fileInfo.IsDir() {
			header.Size = fileInfo.Size()
			return tw.WriteHeader(header)
		}

		// if the current path is a regular file, write its header and bytes to the tar writer
		if fileInfo.Mode().IsRegular() {
			b, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}

			header.Size = int64(len(b))
			if err := tw.WriteHeader(header); err != nil {
				return err
			}

			_, err = tw.Write(b)

			return err
		}

		return nil
	}
}
