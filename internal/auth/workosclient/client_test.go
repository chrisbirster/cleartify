package workosclient

import (
	"errors"
	"net/url"
	"testing"
)

func TestSDKClientConfiguredAndClientID(t *testing.T) {
	t.Parallel()

	client := New("api-key", "client-id")
	if !client.Configured() {
		t.Fatal("expected client to be configured")
	}
	if client.ClientID() != "client-id" {
		t.Fatalf("expected client id, got %q", client.ClientID())
	}

	if New("", "client-id").Configured() {
		t.Fatal("expected missing api key to leave client unconfigured")
	}
}

func TestSDKClientAuthorizationURL(t *testing.T) {
	t.Parallel()

	client := New("api-key", "client-id")
	rawURL, err := client.AuthorizationURL(AuthorizationURLOpts{
		RedirectURI: "http://localhost:3000/api/auth/callback",
		State:       "signed-state",
	})
	if err != nil {
		t.Fatalf("expected authorization url: %v", err)
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("expected parseable url: %v", err)
	}
	if parsed.Query().Get("provider") != "authkit" {
		t.Fatalf("expected authkit provider, got %q", parsed.Query().Get("provider"))
	}
	if parsed.Query().Get("client_id") != "client-id" {
		t.Fatalf("expected client id query param, got %q", parsed.Query().Get("client_id"))
	}
}

func TestSDKClientAuthorizationURLRejectsUnconfiguredClient(t *testing.T) {
	t.Parallel()

	if _, err := New("", "").AuthorizationURL(AuthorizationURLOpts{}); err == nil {
		t.Fatal("expected unconfigured client error")
	}
}

func TestSDKClientAuthorizationURLRejectsInvalidBaseURL(t *testing.T) {
	originalBaseURL := authorizationBaseURL
	t.Cleanup(func() { authorizationBaseURL = originalBaseURL })

	authorizationBaseURL = "://bad"
	if _, err := New("api-key", "client-id").AuthorizationURL(AuthorizationURLOpts{}); err == nil {
		t.Fatal("expected invalid base url error")
	}
}

func TestFakeClientDefaults(t *testing.T) {
	t.Parallel()

	client := &FakeClient{ConfiguredValue: true}
	if !client.Configured() {
		t.Fatal("expected fake client to be configured")
	}

	rawURL, err := client.AuthorizationURL(AuthorizationURLOpts{
		RedirectURI: "http://localhost:3000/api/auth/callback",
		State:       "signed-state",
	})
	if err != nil {
		t.Fatalf("expected default auth url: %v", err)
	}
	if rawURL == "" {
		t.Fatal("expected non-empty auth url")
	}
}

func TestFakeClientUsesExplicitValuesAndErrors(t *testing.T) {
	t.Parallel()

	client := &FakeClient{
		ConfiguredValue:       true,
		ClientIDValue:         "explicit-client",
		AuthorizationURLValue: "https://example.com/auth",
	}
	if client.ClientID() != "explicit-client" {
		t.Fatalf("expected explicit client id, got %q", client.ClientID())
	}
	rawURL, err := client.AuthorizationURL(AuthorizationURLOpts{})
	if err != nil {
		t.Fatalf("expected explicit url: %v", err)
	}
	if rawURL != "https://example.com/auth" {
		t.Fatalf("expected explicit auth url, got %q", rawURL)
	}

	client.AuthorizationURLError = &url.Error{Op: "parse", URL: "bad", Err: errTest}
	if _, err := client.AuthorizationURL(AuthorizationURLOpts{}); err == nil {
		t.Fatal("expected fake auth url error")
	}
}

var errTest = errors.New("boom")
