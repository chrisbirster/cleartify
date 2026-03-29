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

func TestHandleSignals(t *testing.T) {
	t.Parallel()

	server, err := New(
		config.Config{AppMarketingURL: "http://localhost:4173"},
		&fakeStore{signals: []domain.Signal{{ID: "sig_1", Title: "Demo signal"}}},
		auth.NewDemoAuthService(),
		newTestServer(t).frontend,
	)
	if err != nil {
		t.Fatalf("expected server build: %v", err)
	}

	recorder := httptest.NewRecorder()
	server.handleSignals(recorder, httptest.NewRequest(http.MethodGet, "/api/signals", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), `"id":"sig_1"`) {
		t.Fatalf("expected signal id in response, got %q", recorder.Body.String())
	}
}

func TestHandleSignalsFailure(t *testing.T) {
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
	server.handleSignals(recorder, httptest.NewRequest(http.MethodGet, "/api/signals", nil))
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", recorder.Code)
	}
}
