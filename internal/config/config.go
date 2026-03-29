package config

import (
	"fmt"
	"strconv"
	"strings"
)

type Config struct {
	Port                 string
	DatabasePath         string
	AppPublicURL         string
	AppMarketingURL      string
	WorkOSAPIKey         string
	WorkOSClientID       string
	WorkOSCookiePassword string
	SeedDemo             bool
}

func Load(getenv func(string) string) (Config, error) {
	port := firstNonEmpty(getenv("PORT"), "8000")
	if _, err := strconv.Atoi(port); err != nil {
		return Config{}, fmt.Errorf("invalid PORT %q", port)
	}

	seedDemo, err := parseBool(firstNonEmpty(getenv("SEED_DEMO"), "true"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid SEED_DEMO: %w", err)
	}

	return Config{
		Port:                 port,
		DatabasePath:         firstNonEmpty(getenv("DATABASE_PATH"), ".tasktmp/dev.sqlite.db"),
		AppPublicURL:         firstNonEmpty(getenv("APP_PUBLIC_URL"), "http://localhost:3000"),
		AppMarketingURL:      firstNonEmpty(getenv("APP_MARKETING_URL"), "http://localhost:4173"),
		WorkOSAPIKey:         strings.TrimSpace(getenv("WORKOS_API_KEY")),
		WorkOSClientID:       strings.TrimSpace(getenv("WORKOS_CLIENT_ID")),
		WorkOSCookiePassword: strings.TrimSpace(getenv("WORKOS_COOKIE_PASSWORD")),
		SeedDemo:             seedDemo,
	}, nil
}

func (c Config) Addr() string {
	return ":" + c.Port
}

func (c Config) WorkOSConfigured() bool {
	return c.WorkOSAPIKey != "" && c.WorkOSClientID != "" && c.WorkOSCookiePassword != ""
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func parseBool(raw string) (bool, error) {
	value, err := strconv.ParseBool(strings.TrimSpace(raw))
	if err != nil {
		return false, err
	}
	return value, nil
}
