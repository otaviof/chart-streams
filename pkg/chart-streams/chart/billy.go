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

	billy "gopkg.in/src-d/go-billy.v4"

	"github.com/otaviof/chart-streams/pkg/billyutil"
)

type billyChartBuilder struct {
	Filesystem billy.Filesystem
	ChartPath  *string
	ChartName  *string
	CommitTime *time.Time
}

var _ ChartBuilder = &billyChartBuilder{}

// NewBillyChartBuilder builds charts with content stored in billy filesystem.
func NewBillyChartBuilder(fs billy.Filesystem) ChartBuilder {
	return &billyChartBuilder{Filesystem: fs}
}

func (cb *billyChartBuilder) SetChartName(n string) ChartBuilder {
	cb.ChartName = &n
	return cb
}

func (cb *billyChartBuilder) SetChartPath(p string) ChartBuilder {
	cb.ChartPath = &p
	return cb
}

func (cb *billyChartBuilder) SetCommitTime(t time.Time) ChartBuilder {
	cb.CommitTime = &t
	return cb
}

func (cb *billyChartBuilder) Build() (*Package, error) {

	b := bytes.NewBuffer([]byte{})
	bw := bufio.NewWriter(b)
	gzw := gzip.NewWriter(bw)
	tw := tar.NewWriter(gzw)

	walkErr := billyutil.Walk(
		cb.Filesystem,
		*cb.ChartPath,
		func(fs billy.Filesystem, path string, fileInfo os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			archivePath := fs.Join(*cb.ChartName, strings.TrimPrefix(path, *cb.ChartPath))

			header, err := tar.FileInfoHeader(fileInfo, fileInfo.Name())
			if err != nil {
				return err
			}

			header.Name = archivePath

			if fileInfo.IsDir() {
				header.Mode = int64(fileInfo.Mode())
				header.ModTime = *cb.CommitTime
				header.Size = fileInfo.Size()
				return tw.WriteHeader(header)
			}

			if !fileInfo.Mode().IsRegular() {
				return nil
			}

			f, openErr := fs.Open(path)
			if openErr != nil {
				return openErr
			}
			defer f.Close()

			b, err := ioutil.ReadAll(f)
			if err != nil {
				return err
			}

			header.Mode = int64(fileInfo.Mode())
			header.ModTime = *cb.CommitTime
			header.Size = int64(len(b))

			if err := tw.WriteHeader(header); err != nil {
				return err
			}

			_, err = tw.Write(b)
			return err
		})

	tw.Close()
	gzw.Close()
	bw.Flush()

	if walkErr != nil {
		return nil, walkErr
	}

	p := &Package{bytesBuffer: b}

	return p, nil
}
