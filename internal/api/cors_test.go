package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCORSAllowsCredentialsAndMutationMethodsOnlyForConfiguredOrigin(t *testing.T) {
	for _, origin := range []string{"https://app.example.test", "https://evil.example.test", ""} {
		t.Run(origin, func(t *testing.T) {
			handler := New(Options{AppOrigin: "https://app.example.test"})
			req := httptest.NewRequest(http.MethodOptions, "/api/me/profile", nil)
			req.Header.Set("Origin", origin)
			req.Header.Set("Access-Control-Request-Method", http.MethodPut)
			req.Header.Set("Access-Control-Request-Headers", "Content-Type")
			res := httptest.NewRecorder()
			handler.ServeHTTP(res, req)
			if res.Code != http.StatusNoContent {
				t.Fatalf("preflight: %d", res.Code)
			}
			if !strings.Contains(res.Header().Get("Vary"), "Origin") {
				t.Fatal("CORS response must vary even for rejected origins")
			}
			if origin != "https://app.example.test" {
				if res.Header().Get("Access-Control-Allow-Origin") != "" || res.Header().Get("Access-Control-Allow-Credentials") != "" {
					t.Fatal("untrusted origin allowed credentials")
				}
				return
			}
			if res.Header().Get("Access-Control-Allow-Origin") != origin || res.Header().Get("Access-Control-Allow-Credentials") != "true" {
				t.Fatal("configured origin cannot use credentials")
			}
			methods := strings.Split(res.Header().Get("Access-Control-Allow-Methods"), ",")
			for _, want := range []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"} {
				found := false
				for _, method := range methods {
					if strings.TrimSpace(method) == want {
						found = true
					}
				}
				if !found {
					t.Fatalf("missing allowed method %s", want)
				}
			}
		})
	}
}
