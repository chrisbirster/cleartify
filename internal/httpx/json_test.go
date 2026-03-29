package httpx

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	WriteJSON(recorder, 201, map[string]string{"status": "ok"})

	if recorder.Code != 201 {
		t.Fatalf("expected status 201, got %d", recorder.Code)
	}
	if contentType := recorder.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("expected json content type, got %q", contentType)
	}
	if !strings.Contains(recorder.Body.String(), `"status":"ok"`) {
		t.Fatalf("expected response body to contain status, got %q", recorder.Body.String())
	}
}
