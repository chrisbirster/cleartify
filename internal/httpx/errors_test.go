package httpx

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWriteError(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	WriteError(recorder, http.StatusBadRequest, "bad_request", "Something went wrong.")

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), `"code":"bad_request"`) {
		t.Fatalf("expected error code in body, got %q", recorder.Body.String())
	}
}

func TestMethodNotAllowed(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	MethodNotAllowed(recorder)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status 405, got %d", recorder.Code)
	}
}
