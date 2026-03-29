package server

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"cleartify/internal/config"
	"cleartify/internal/domain"
)

func TestHandleSession(t *testing.T) {
	t.Parallel()

	server, err := New(
		config.Config{AppMarketingURL: "http://localhost:4173"},
		&fakeStore{},
		&fakeAuthService{session: domain.AuthSession{Authenticated: true}},
		newTestServer(t).frontend,
	)
	if err != nil {
		t.Fatalf("expected server build: %v", err)
	}

	recorder := httptest.NewRecorder()
	server.handleSession(recorder, httptest.NewRequest(http.MethodGet, "/api/auth/session", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
}

func TestHandleSessionFailure(t *testing.T) {
	t.Parallel()

	server, err := New(
		config.Config{AppMarketingURL: "http://localhost:4173"},
		&fakeStore{},
		&fakeAuthService{err: errors.New("boom")},
		newTestServer(t).frontend,
	)
	if err != nil {
		t.Fatalf("expected server build: %v", err)
	}

	recorder := httptest.NewRecorder()
	server.handleSession(recorder, httptest.NewRequest(http.MethodGet, "/api/auth/session", nil))
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), `"code":"session_failed"`) {
		t.Fatalf("expected session error code, got %q", recorder.Body.String())
	}
}
