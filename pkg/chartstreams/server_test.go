package chartstreams

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	helmchart "helm.sh/helm/v3/pkg/chart"
	helmrepo "helm.sh/helm/v3/pkg/repo"
	"sigs.k8s.io/yaml"
)

func TestServer(t *testing.T) {
	cfg := &Config{ListenAddr: "127.0.0.1:8080"}
	p := NewFakeChartProvider(cfg)
	s := NewServer(cfg, p)
	g := s.SetupRoutes()

	t.Run("IndexHandler", func(t *testing.T) {
		w := httptest.NewRecorder()

		req, err := http.NewRequest("GET", "/index.yaml", nil)
		assert.NoError(t, err)

		g.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)

		b := w.Body.Bytes()
		assert.True(t, len(b) > 10)

		indexFile := &helmrepo.IndexFile{}
		err = yaml.Unmarshal(b, indexFile)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(indexFile.Entries))
		assert.Equal(t, helmchart.APIVersionV1, indexFile.APIVersion)
	})

	t.Run("DirectLinkHandler", func(t *testing.T) {
		w := httptest.NewRecorder()

		req, err := http.NewRequest("GET", "/chart/chart/0.0.1", nil)
		assert.NoError(t, err)

		g.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)

		b := w.Body.Bytes()
		assert.True(t, len(b) > 0)
	})
}
