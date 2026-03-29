package auth

import (
	"context"
	"net/http/httptest"
	"testing"

	"cleartify/internal/auth/workosclient"
)

func TestValidateWorkOSConfig(t *testing.T) {
	t.Parallel()

	if err := ValidateWorkOSConfig("api", "client", "cookie"); err != nil {
		t.Fatalf("expected valid config: %v", err)
	}
	if err := ValidateWorkOSConfig("", "client", "cookie"); err == nil {
		t.Fatal("expected api key validation error")
	}
	if err := ValidateWorkOSConfig("api", "", "cookie"); err == nil {
		t.Fatal("expected client id validation error")
	}
	if err := ValidateWorkOSConfig("api", "client", ""); err == nil {
		t.Fatal("expected cookie password validation error")
	}
}

func TestNewWorkOSAuthServiceAndCurrentSession(t *testing.T) {
	t.Parallel()

	service, err := NewWorkOSAuthService("api", "client", "cookie", &workosclient.FakeClient{
		ConfiguredValue: true,
	})
	if err != nil {
		t.Fatalf("expected workos service to build: %v", err)
	}

	session, err := service.CurrentSession(context.Background(), httptest.NewRequest("GET", "/api/auth/session", nil))
	if err != nil {
		t.Fatalf("expected current session to succeed: %v", err)
	}
	if session.Authenticated {
		t.Fatal("expected workos skeleton session to be unauthenticated")
	}
}

func TestNewWorkOSAuthServiceRejectsUnconfiguredClient(t *testing.T) {
	t.Parallel()

	if _, err := NewWorkOSAuthService("api", "client", "cookie", &workosclient.FakeClient{}); err == nil {
		t.Fatal("expected unconfigured client error")
	}
}

func TestNewWorkOSAuthServiceBuildsDefaultClientWhenNil(t *testing.T) {
	t.Parallel()

	service, err := NewWorkOSAuthService("api", "client", "cookie", nil)
	if err != nil {
		t.Fatalf("expected default client service: %v", err)
	}
	if service == nil {
		t.Fatal("expected non-nil service")
	}
}

func TestNewWorkOSAuthServiceReturnsValidationError(t *testing.T) {
	t.Parallel()

	if _, err := NewWorkOSAuthService("", "client", "cookie", nil); err == nil {
		t.Fatal("expected validation error")
	}
}
