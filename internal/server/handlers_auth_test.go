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

func TestHandleLoginRedirectsToHostedAuth(t *testing.T) {
	t.Parallel()

	server, err := New(
		config.Config{AppPublicURL: "http://localhost:3000", AppMarketingURL: "http://localhost:4173"},
		&fakeStore{},
		&fakeAuthService{loginURL: "https://workos.example/authorize"},
		newTestServer(t).frontend,
	)
	if err != nil {
		t.Fatalf("expected server build: %v", err)
	}

	recorder := httptest.NewRecorder()
	server.handleLogin(recorder, httptest.NewRequest(http.MethodGet, "/login", nil))
	if recorder.Code != http.StatusFound {
		t.Fatalf("expected 302, got %d", recorder.Code)
	}
	if location := recorder.Header().Get("Location"); location != "https://workos.example/authorize" {
		t.Fatalf("expected hosted login redirect, got %q", location)
	}
}

func TestHandleAuthCallbackSetsSessionCookie(t *testing.T) {
	t.Parallel()

	server, err := New(
		config.Config{AppPublicURL: "http://localhost:3000", AppMarketingURL: "http://localhost:4173"},
		&fakeStore{},
		&fakeAuthService{callbackResult: auth.CallbackResult{SessionID: "session_123", ReturnTo: "/work/wo_1"}},
		newTestServer(t).frontend,
	)
	if err != nil {
		t.Fatalf("expected server build: %v", err)
	}

	recorder := httptest.NewRecorder()
	server.handleAuthCallback(recorder, httptest.NewRequest(http.MethodGet, "/auth/callback?code=abc&state=signed", nil))
	if recorder.Code != http.StatusFound {
		t.Fatalf("expected 302, got %d", recorder.Code)
	}
	if location := recorder.Header().Get("Location"); location != "/work/wo_1" {
		t.Fatalf("expected callback redirect, got %q", location)
	}
	cookie := recorder.Result().Cookies()[0]
	if cookie.Name != auth.SessionCookieName || cookie.Value != "session_123" {
		t.Fatalf("expected session cookie, got %#v", cookie)
	}
}

func TestHandleAuthCallbackFailureRedirectsToAuthError(t *testing.T) {
	t.Parallel()

	server, err := New(
		config.Config{AppPublicURL: "http://localhost:3000", AppMarketingURL: "http://localhost:4173"},
		&fakeStore{},
		&fakeAuthService{err: auth.ErrInvalidCallback},
		newTestServer(t).frontend,
	)
	if err != nil {
		t.Fatalf("expected server build: %v", err)
	}

	recorder := httptest.NewRecorder()
	server.handleAuthCallback(recorder, httptest.NewRequest(http.MethodGet, "/auth/callback?state=bad", nil))
	if recorder.Code != http.StatusFound {
		t.Fatalf("expected 302, got %d", recorder.Code)
	}
	if location := recorder.Header().Get("Location"); !strings.Contains(location, "authError=auth_failed") {
		t.Fatalf("expected auth failed redirect, got %q", location)
	}
}

func TestHandleLogoutClearsSessionCookie(t *testing.T) {
	t.Parallel()

	server, err := New(
		config.Config{AppPublicURL: "http://localhost:3000", AppMarketingURL: "http://localhost:4173"},
		&fakeStore{},
		&fakeAuthService{logoutURL: "https://workos.example/logout"},
		newTestServer(t).frontend,
	)
	if err != nil {
		t.Fatalf("expected server build: %v", err)
	}

	recorder := httptest.NewRecorder()
	server.handleLogout(recorder, httptest.NewRequest(http.MethodGet, "/logout", nil))
	if recorder.Code != http.StatusFound {
		t.Fatalf("expected 302, got %d", recorder.Code)
	}
	if location := recorder.Header().Get("Location"); location != "https://workos.example/logout" {
		t.Fatalf("expected logout redirect, got %q", location)
	}
	if cookie := recorder.Result().Cookies()[0]; cookie.MaxAge != -1 {
		t.Fatalf("expected cleared session cookie, got %#v", cookie)
	}
}

func TestHandleCompleteOnboarding(t *testing.T) {
	t.Parallel()

	server, err := New(
		config.Config{AppPublicURL: "http://localhost:3000", AppMarketingURL: "http://localhost:4173"},
		&fakeStore{},
		&fakeAuthService{
			completedSession: domain.AuthSession{
				Authenticated:   true,
				NeedsOnboarding: false,
				User:            &domain.User{ID: "usr_demo", Email: "demo@cleartify.com", DisplayName: "Demo User"},
			},
		},
		newTestServer(t).frontend,
	)
	if err != nil {
		t.Fatalf("expected server build: %v", err)
	}

	recorder := httptest.NewRecorder()
	server.handleCompleteOnboarding(recorder, httptest.NewRequest(http.MethodPost, "/api/me/onboarding/complete", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), `"needsOnboarding":false`) {
		t.Fatalf("expected onboarding completion response, got %q", recorder.Body.String())
	}
}

func TestHandleCompleteOnboardingRequiresAuth(t *testing.T) {
	t.Parallel()

	server, err := New(
		config.Config{AppPublicURL: "http://localhost:3000", AppMarketingURL: "http://localhost:4173"},
		&fakeStore{},
		&fakeAuthService{err: auth.ErrUnauthenticated},
		newTestServer(t).frontend,
	)
	if err != nil {
		t.Fatalf("expected server build: %v", err)
	}

	recorder := httptest.NewRecorder()
	server.handleCompleteOnboarding(recorder, httptest.NewRequest(http.MethodPost, "/api/me/onboarding/complete", nil))
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", recorder.Code)
	}
}

func TestHandleWorkspaceMode(t *testing.T) {
	t.Parallel()

	server, err := New(
		config.Config{AppPublicURL: "http://localhost:3000", AppMarketingURL: "http://localhost:4173"},
		&fakeStore{},
		&fakeAuthService{
			selectedWorkspaceSession: domain.AuthSession{
				Authenticated:   true,
				NeedsOnboarding: false,
				User:            &domain.User{ID: "usr_demo", Email: "demo@cleartify.com", DisplayName: "Demo User"},
				Workspace:       &domain.Workspace{ID: "wrk_demo", Name: "Cleartify Demo Team", Slug: "cleartify-demo-team", Mode: "sample"},
			},
		},
		newTestServer(t).frontend,
	)
	if err != nil {
		t.Fatalf("expected server build: %v", err)
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/workspace-mode", strings.NewReader(`{"mode":"sample"}`))
	server.handleWorkspaceMode(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), `"mode":"sample"`) {
		t.Fatalf("expected sample mode in response, got %q", recorder.Body.String())
	}
}

func TestHandleWorkspaceModeRejectsInvalidMode(t *testing.T) {
	t.Parallel()

	server, err := New(
		config.Config{AppPublicURL: "http://localhost:3000", AppMarketingURL: "http://localhost:4173"},
		&fakeStore{},
		&fakeAuthService{err: auth.ErrInvalidWorkspaceMode},
		newTestServer(t).frontend,
	)
	if err != nil {
		t.Fatalf("expected server build: %v", err)
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/workspace-mode", strings.NewReader(`{"mode":"invalid"}`))
	server.handleWorkspaceMode(recorder, request)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
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
