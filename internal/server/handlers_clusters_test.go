package server

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"cleartify/internal/auth"
	"cleartify/internal/config"
	"cleartify/internal/domain"
)

func TestHandleClusters(t *testing.T) {
	t.Parallel()

	server, err := New(
		config.Config{AppMarketingURL: "http://localhost:4173"},
		&fakeStore{clusters: []domain.Cluster{{ID: "clu_1", Title: "Demo cluster"}}},
		auth.NewDemoAuthService(),
		newTestServer(t).frontend,
	)
	if err != nil {
		t.Fatalf("expected server build: %v", err)
	}

	recorder := httptest.NewRecorder()
	server.handleClusters(recorder, httptest.NewRequest(http.MethodGet, "/api/clusters", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), `"id":"clu_1"`) {
		t.Fatalf("expected cluster id in response, got %q", recorder.Body.String())
	}
}

func TestHandleClustersFailure(t *testing.T) {
	t.Parallel()

	server, err := New(
		config.Config{AppMarketingURL: "http://localhost:4173"},
		&fakeStore{err: errors.New("boom")},
		auth.NewDemoAuthService(),
		newTestServer(t).frontend,
	)
	if err != nil {
		t.Fatalf("expected server build: %v", err)
	}

	recorder := httptest.NewRecorder()
	server.handleClusters(recorder, httptest.NewRequest(http.MethodGet, "/api/clusters", nil))
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", recorder.Code)
	}
}
