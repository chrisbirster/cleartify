package config

import "testing"

func TestLoadUsesDefaults(t *testing.T) {
	t.Parallel()

	cfg, err := Load(func(string) string { return "" })
	if err != nil {
		t.Fatalf("expected defaults to load: %v", err)
	}

	if cfg.Port != "8000" {
		t.Fatalf("expected default port 8000, got %q", cfg.Port)
	}
	if cfg.DatabasePath != ".tasktmp/dev.sqlite.db" {
		t.Fatalf("expected default database path, got %q", cfg.DatabasePath)
	}
	if !cfg.SeedDemo {
		t.Fatal("expected seed demo to default to true")
	}
	if cfg.AuthMode != "workos" {
		t.Fatalf("expected default auth mode workos, got %q", cfg.AuthMode)
	}
	if cfg.WorkOSRedirectURI != "http://localhost:3000/auth/callback" {
		t.Fatalf("expected default redirect uri, got %q", cfg.WorkOSRedirectURI)
	}
}

func TestLoadOverridesAndWorkOSConfigured(t *testing.T) {
	t.Parallel()

	values := map[string]string{
		"PORT":                   "9001",
		"DATABASE_PATH":          "/tmp/cleartify.db",
		"APP_PUBLIC_URL":         "https://app.example.com",
		"APP_MARKETING_URL":      "https://example.com",
		"AUTH_MODE":              "demo",
		"WORKOS_API_KEY":         "api-key",
		"WORKOS_CLIENT_ID":       "client-id",
		"WORKOS_REDIRECT_URI":    "https://app.example.com/auth/callback",
		"WORKOS_COOKIE_PASSWORD": "cookie-secret",
		"SEED_DEMO":              "false",
	}

	cfg, err := Load(func(key string) string { return values[key] })
	if err != nil {
		t.Fatalf("expected config to load: %v", err)
	}

	if cfg.Addr() != ":9001" {
		t.Fatalf("expected addr :9001, got %q", cfg.Addr())
	}
	if !cfg.WorkOSConfigured() {
		t.Fatal("expected workos to be configured")
	}
	if !cfg.DemoAuthEnabled() {
		t.Fatal("expected explicit demo auth mode")
	}
	if cfg.SeedDemo {
		t.Fatal("expected seed demo to be false")
	}
	if cfg.WorkOSRedirectURI != "https://app.example.com/auth/callback" {
		t.Fatalf("expected override redirect uri, got %q", cfg.WorkOSRedirectURI)
	}
}

func TestLoadRejectsInvalidValues(t *testing.T) {
	t.Parallel()

	if _, err := Load(func(key string) string {
		if key == "PORT" {
			return "abc"
		}
		return ""
	}); err == nil {
		t.Fatal("expected invalid port error")
	}

	if _, err := Load(func(key string) string {
		if key == "AUTH_MODE" {
			return "magic"
		}
		return ""
	}); err == nil {
		t.Fatal("expected invalid auth mode error")
	}

	if _, err := Load(func(key string) string {
		if key == "SEED_DEMO" {
			return "maybe"
		}
		return ""
	}); err == nil {
		t.Fatal("expected invalid bool error")
	}
}

func TestFirstNonEmptyHandlesEmptyInputs(t *testing.T) {
	t.Parallel()

	if value := firstNonEmpty("", "  ", "value"); value != "value" {
		t.Fatalf("expected first non-empty value, got %q", value)
	}
	if value := firstNonEmpty("", " "); value != "" {
		t.Fatalf("expected empty result, got %q", value)
	}
}
