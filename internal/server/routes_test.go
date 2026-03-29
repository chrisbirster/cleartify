package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRoutesRegisterKnownAPIEndpoints(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)

	for _, path := range []string{
		"/api/health",
		"/api/auth/session",
		"/api/signals",
		"/api/clusters",
		"/api/work-objects",
	} {
		request := httptest.NewRequest(http.MethodGet, path, nil)
		recorder := httptest.NewRecorder()
		server.ServeHTTP(recorder, request)
		if recorder.Code == http.StatusNotFound {
			t.Fatalf("expected route %s to be registered", path)
		}
	}
}

func TestRoutesRegisterClusterWorkObjectCreateEndpoint(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)

	request := httptest.NewRequest(http.MethodPost, "/api/clusters/clu_1/create-work-object", nil)
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, request)
	if recorder.Code == http.StatusNotFound {
		t.Fatal("expected cluster create-work-object route to be registered")
	}
}

func TestClusterWorkObjectCreateRejectsWrongMethod(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)

	request := httptest.NewRequest(http.MethodGet, "/api/clusters/clu_1/create-work-object", nil)
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405 for wrong method, got %d", recorder.Code)
	}
}
