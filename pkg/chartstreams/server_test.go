package chartstreams

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-github/v39/github"
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
		assert.Equal(t, http.StatusOK, w.Code)

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
		assert.Equal(t, http.StatusOK, w.Code)

		b := w.Body.Bytes()
		assert.True(t, len(b) > 0)
	})
}

func TestServer_GitHubPullTriggerHandler(t *testing.T) {
	t.Run("GitHubWebhookSecret specified", func(t *testing.T) {
		cfg := &Config{ListenAddr: "127.0.0.1:8080", GitHubWebhookSecret: "<SECRET>"}
		p := NewFakeChartProvider(cfg)
		s := NewServer(cfg, p)
		g := s.SetupRoutes()

		w := httptest.NewRecorder()

		evt := &github.PushEvent{
			Ref: github.String("master"),
		}
		evtBytes, err := json.Marshal(evt)
		assert.NoError(t, err)

		evtReader := bytes.NewReader(evtBytes)
		req, err := http.NewRequest("POST", "/api/webhooks/github", evtReader)
		assert.NoError(t, err)

		g.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("GitHubWebhookSecret not specified", func(t *testing.T) {
		cfg := &Config{ListenAddr: "127.0.0.1:8080"}
		p := NewFakeChartProvider(cfg)
		s := NewServer(cfg, p)
		g := s.SetupRoutes()

		t.Run("valid event should return OK", func(t *testing.T) {
			w := httptest.NewRecorder()

			evt := &github.PushEvent{
				Ref: github.String("master"),
			}
			evtBytes, _ := json.Marshal(evt)

			evtReader := bytes.NewReader(evtBytes)
			req, err := http.NewRequest("POST", "/api/webhooks/github", evtReader)
			assert.NoError(t, err)

			g.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code)
		})

		t.Run("invalid JSON payload should return BadRequest", func(t *testing.T) {
			w := httptest.NewRecorder()

			req, err := http.NewRequest("POST", "/api/webhooks/github",
				bytes.NewReader([]byte("{invalid JSON payload")))
			assert.NoError(t, err)

			g.ServeHTTP(w, req)
			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	})
}
