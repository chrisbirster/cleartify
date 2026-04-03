package workosclient

import (
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
		RedirectURI: "http://localhost:3000/auth/callback",
		State:       "signed-state",
		ScreenHint:  "sign-in",
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
	if parsed.Scheme != "https" || parsed.Host != "api.workos.com" {
		t.Fatalf("expected hosted workos auth url, got %q", rawURL)
	}
	if parsed.Query().Get("client_id") != "client-id" {
		t.Fatalf("expected client id query param, got %q", parsed.Query().Get("client_id"))
	}
	if parsed.Query().Get("screen_hint") != "sign-in" {
		t.Fatalf("expected screen hint query param, got %q", parsed.Query().Get("screen_hint"))
	}
}

func TestSDKClientLogoutURLUsesHostedWorkOSURL(t *testing.T) {
	t.Parallel()

	client := New("api-key", "client-id")
	rawURL, err := client.LogoutURL("session_123", "http://localhost:4173/")
	if err != nil {
		t.Fatalf("expected logout url: %v", err)
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("expected parseable logout url: %v", err)
	}
	if parsed.Scheme != "https" || parsed.Host != "api.workos.com" {
		t.Fatalf("expected hosted workos logout url, got %q", rawURL)
	}
	if parsed.Query().Get("session_id") != "session_123" {
		t.Fatalf("expected session id query param, got %q", parsed.Query().Get("session_id"))
	}
}

func TestSDKClientAuthorizationURLRejectsUnconfiguredClient(t *testing.T) {
	t.Parallel()

	if _, err := New("", "").AuthorizationURL(AuthorizationURLOpts{}); err == nil {
		t.Fatal("expected unconfigured client error")
	}
}

func TestFakeClientDefaults(t *testing.T) {
	t.Parallel()

	client := &FakeClient{ConfiguredValue: true}
	if !client.Configured() {
		t.Fatal("expected fake client to be configured")
	}

	rawURL, err := client.AuthorizationURL(AuthorizationURLOpts{
		RedirectURI: "http://localhost:3000/auth/callback",
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

	client.AuthorizationURLError = &url.Error{Op: "parse", URL: "bad", Err: url.InvalidHostError("boom")}
	if _, err := client.AuthorizationURL(AuthorizationURLOpts{}); err == nil {
		t.Fatal("expected fake auth url error")
	}
}
