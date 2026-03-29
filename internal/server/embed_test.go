package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

func TestHandleAppServesSPAAndAssets(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)

	for _, path := range []string{"/app", "/app/", "/app/inbox", "/app/work/wo_mobile_perf"} {
		recorder := httptest.NewRecorder()
		server.handleApp(recorder, httptest.NewRequest(http.MethodGet, path, nil))
		if recorder.Code != http.StatusOK {
			t.Fatalf("expected 200 for %s, got %d", path, recorder.Code)
		}
		if !strings.Contains(recorder.Body.String(), "app") {
			t.Fatalf("expected embedded index for %s, got %q", path, recorder.Body.String())
		}
	}

	recorder := httptest.NewRecorder()
	server.handleApp(recorder, httptest.NewRequest(http.MethodGet, "/app/assets/app.js", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected asset to be served, got %d", recorder.Code)
	}

	fileRecorder := httptest.NewRecorder()
	server.handleApp(fileRecorder, httptest.NewRequest(http.MethodGet, "/app/favicon.ico", nil))
	if fileRecorder.Code != http.StatusOK {
		t.Fatalf("expected extension asset to be served, got %d", fileRecorder.Code)
	}
}

func TestHandleRootRedirectsAndReturnsAPINotFound(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)

	recorder := httptest.NewRecorder()
	server.handleRoot(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	if recorder.Code != http.StatusTemporaryRedirect {
		t.Fatalf("expected marketing redirect, got %d", recorder.Code)
	}

	apiRecorder := httptest.NewRecorder()
	server.handleRoot(apiRecorder, httptest.NewRequest(http.MethodGet, "/api/missing", nil))
	if apiRecorder.Code != http.StatusNotFound {
		t.Fatalf("expected api 404, got %d", apiRecorder.Code)
	}

	notFoundRecorder := httptest.NewRecorder()
	server.handleRoot(notFoundRecorder, httptest.NewRequest(http.MethodPost, "/missing", nil))
	if notFoundRecorder.Code != http.StatusNotFound {
		t.Fatalf("expected non-get root 404, got %d", notFoundRecorder.Code)
	}
}

func TestServeIndexReturnsServerErrorWhenIndexIsMissing(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	server.frontend = fstest.MapFS{}

	recorder := httptest.NewRecorder()
	server.serveIndex(recorder)
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected missing index to return 500, got %d", recorder.Code)
	}
}
