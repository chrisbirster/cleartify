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
}

func TestLoadOverridesAndWorkOSConfigured(t *testing.T) {
	t.Parallel()

	values := map[string]string{
		"PORT":                   "9001",
		"DATABASE_PATH":          "/tmp/cleartify.db",
		"APP_PUBLIC_URL":         "https://app.example.com",
		"APP_MARKETING_URL":      "https://example.com",
		"WORKOS_API_KEY":         "api-key",
		"WORKOS_CLIENT_ID":       "client-id",
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
	if cfg.SeedDemo {
		t.Fatal("expected seed demo to be false")
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
