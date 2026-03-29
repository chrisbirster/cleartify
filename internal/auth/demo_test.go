package auth

import (
	"context"
	"net/http/httptest"
	"testing"
)

func TestDemoAuthServiceReturnsAuthenticatedSession(t *testing.T) {
	t.Parallel()

	service := NewDemoAuthService()
	session, err := service.CurrentSession(context.Background(), httptest.NewRequest("GET", "/api/auth/session", nil))
	if err != nil {
		t.Fatalf("expected demo session: %v", err)
	}

	if !session.Authenticated {
		t.Fatal("expected demo auth to be authenticated")
	}
	if session.User == nil || session.User.Email != "demo@cleartify.com" {
		t.Fatalf("expected demo user, got %#v", session.User)
	}
}
