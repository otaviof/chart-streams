package chartstreams

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/json"
	"fmt"
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
		secret := []byte("0406d00c77334cdceafae")
		cfg := &Config{ListenAddr: "127.0.0.1:8080", GitHubWebhookSecret: string(secret)}
		p := NewFakeChartProvider(cfg)
		s := NewServer(cfg, p)
		g := s.SetupRoutes()

		w := httptest.NewRecorder()

		evt := &github.PushEvent{
			Ref: github.String("master"),
		}
		evtBytes, err := json.Marshal(evt)
		assert.NoError(t, err)

		signature := signBody(secret, evtBytes)
		bodyReader := bytes.NewReader(evtBytes)
		req, err := http.NewRequest("POST", "/api/webhooks/github", bodyReader)
		assert.NoError(t, err)

		// See https://github.com/isutton/chart-streams/blob/7ef719f38aef56e7fe9f3e72960ccea5d75c4f07/vendor/gopkg.in/rjz/githubhook.v0/githubhook.go#L70-L90
		//
		// TODO(isutton): sha1 should be substituted by sha256.
		//				  see https://docs.github.com/en/developers/webhooks-and-events/webhooks/securing-your-webhooks
		req.Header.Add("x-hub-signature", fmt.Sprintf("sha1=%x", string(signature)))
		req.Header.Add("x-github-event", "push")
		req.Header.Add("x-github-delivery", "<DELIVERY_ID>")

		g.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
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

// signBody returns the body signature with the body and the given secret.
//
// This implementation has been copied from githubhook.
//
// see https://github.com/isutton/chart-streams/blob/7ef719f38aef56e7fe9f3e72960ccea5d75c4f07/vendor/gopkg.in/rjz/githubhook.v0/githubhook.go#L43-L47
func signBody(secret, body []byte) []byte {
	computed := hmac.New(sha1.New, secret)
	computed.Write(body)
	return []byte(computed.Sum(nil))
}
