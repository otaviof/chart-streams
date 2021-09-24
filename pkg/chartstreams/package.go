package chartstreams

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"path"
	"time"

	log "github.com/sirupsen/logrus"
	helmchart "helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
)

// Package wraps the Helm Chart archive, which consists of chart files in a given directory packaged
// as a tarball, skipping ignored files in the process.
type Package struct {
	dir   string           // abs path to chart
	t     *time.Time       // modification time
	chart *helmchart.Chart // chart instance
	b     *bytes.Buffer    // package content
}

// Bytes return buffer bytes.
func (p *Package) Bytes() []byte {
	return p.b.Bytes()
}

// SHA256 returns the sha256 digest of package contents.
func (p *Package) SHA256() (string, error) {
	digest := sha256.New()
	_, err := digest.Write(p.b.Bytes())
	if err != nil {
		return "", err
	}
	sum := digest.Sum(nil)
	return fmt.Sprintf("%x", sum), nil
}

// skipChartFile checks if informed file is part of ignored list.
func (p *Package) skipChartFile(name string) bool {
	for _, f := range p.chart.Files {
		if f.Name == name {
			return true
		}
	}
	return false
}

// Build inspect chart files skipping ignored files, creating a tarball payload out of chart
// instance.
func (p *Package) Build() error {
	log.Infof("Creating tarball for chart '%s' (%s) from '%s'",
		p.chart.Name(), p.chart.Metadata.Version, p.dir)

	bw := bufio.NewWriter(p.b)
	gzw := gzip.NewWriter(bw)
	tw := tar.NewWriter(gzw)

	for _, f := range p.chart.Raw {
		if p.skipChartFile(f.Name) {
			log.Debugf("Skipping file '%s'", f.Name)
			continue
		}

		log.Debugf("Adding file to tarball '%s'...", f.Name)
		header := &tar.Header{
			Name:    path.Join(p.chart.Name(), f.Name),
			ModTime: *p.t,
			Mode:    int64(0644),
			Size:    int64(len(f.Data)),
		}
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		if _, err := tw.Write(f.Data); err != nil {
			return err
		}
	}

	if err := tw.Close(); err != nil {
		return err
	}
	if err := gzw.Close(); err != nil {
		return err
	}
	return bw.Flush()
}

// NewPackage instantiate a package by inspecting directory and loading the chart.
func NewPackage(dir string, t *time.Time) (*Package, error) {
	chart, err := loader.LoadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("loading chart from %q: %w", dir, err)
	}
	return &Package{
		dir:   dir,
		t:     t,
		chart: chart,
		b:     bytes.NewBuffer([]byte{}),
	}, nil
}
