package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAllowMethodRejectsUnexpectedMethods(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	handler := server.allowMethod(http.MethodGet, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	recorder := httptest.NewRecorder()
	handler(recorder, httptest.NewRequest(http.MethodPost, "/api/health", nil))

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", recorder.Code)
	}
}
