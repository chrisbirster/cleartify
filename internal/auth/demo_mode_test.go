package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"cleartify/internal/domain"
)

func TestDemoModeAuthServiceLoginCurrentSessionAndOnboarding(t *testing.T) {
	t.Parallel()

	store := &fakeSessionStore{}
	service := NewDemoModeAuthService("http://localhost:3000", "http://localhost:4173", "", store)

	loginRequest := httptest.NewRequest("GET", "/login?returnTo=%2Fwork%2Fwo_onboarding_refresh", nil)
	loginURL, err := service.LoginURL(context.Background(), loginRequest)
	if err != nil {
		t.Fatalf("expected demo login url: %v", err)
	}
	if loginURL == "" {
		t.Fatal("expected demo login url")
	}

	callbackRequest := httptest.NewRequest("GET", loginURL, nil)
	callbackResult, err := service.Authenticate(context.Background(), callbackRequest)
	if err != nil {
		t.Fatalf("expected demo callback to succeed: %v", err)
	}
	if callbackResult.ReturnTo != "/work/wo_onboarding_refresh" {
		t.Fatalf("expected callback return path, got %q", callbackResult.ReturnTo)
	}

	store.session = SessionRecord{
		ID:              callbackResult.SessionID,
		User:            domain.User{ID: demoUser.ID, Email: demoUser.Email, DisplayName: demoUser.DisplayName},
		Workspace:       demoWorkspace(),
		NeedsOnboarding: true,
	}

	sessionRequest := httptest.NewRequest(http.MethodGet, "/api/auth/session", nil)
	sessionRequest.AddCookie(&http.Cookie{Name: SessionCookieName, Value: callbackResult.SessionID})

	session, err := service.CurrentSession(context.Background(), sessionRequest)
	if err != nil {
		t.Fatalf("expected demo current session: %v", err)
	}
	if !session.Authenticated || !session.NeedsOnboarding {
		t.Fatalf("expected authenticated onboarding session, got %#v", session)
	}

	completedSession, err := service.CompleteOnboarding(context.Background(), sessionRequest)
	if err != nil {
		t.Fatalf("expected onboarding completion: %v", err)
	}
	if completedSession.NeedsOnboarding {
		t.Fatal("expected onboarding to be complete")
	}
}

func TestUnavailableAuthServiceReturnsUnauthenticatedAndUnavailable(t *testing.T) {
	t.Parallel()

	service := NewUnavailableAuthService("http://localhost:3000", "http://localhost:4173")

	session, err := service.CurrentSession(context.Background(), httptest.NewRequest(http.MethodGet, "/api/auth/session", nil))
	if err != nil {
		t.Fatalf("expected unauthenticated session: %v", err)
	}
	if session.Authenticated {
		t.Fatal("expected unauthenticated session")
	}

	if _, err := service.LoginURL(context.Background(), httptest.NewRequest(http.MethodGet, "/login", nil)); err == nil {
		t.Fatal("expected unavailable auth error")
	}
	if _, err := service.CompleteOnboarding(context.Background(), httptest.NewRequest(http.MethodPost, "/api/me/onboarding/complete", nil)); err == nil {
		t.Fatal("expected unauthenticated onboarding error")
	}
}
