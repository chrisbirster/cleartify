package config

import (
	"os"
	"strings"
)

type Config struct {
	Port             string
	AppOrigin        string
	AppURL           string
	AppSecret        string
	AdminToken       string
	NotificationMode string
	DatabasePath     string
	TursoRemoteURL   string
	TursoAuthToken   string
}

func Load() Config {
	appOrigin := strings.TrimRight(env("APP_ORIGIN", "http://localhost:3000"), "/")
	return Config{
		Port:             env("PORT", "8080"),
		AppOrigin:        appOrigin,
		AppURL:           strings.TrimRight(env("APP_URL", appOrigin), "/"),
		AppSecret:        strings.TrimSpace(os.Getenv("APP_SECRET")),
		AdminToken:       strings.TrimSpace(os.Getenv("ADMIN_TOKEN")),
		NotificationMode: strings.ToLower(env("NOTIFICATION_MODE", "disabled")),
		DatabasePath:     env("DATABASE_PATH", "./data/cleartify.db"),
		TursoRemoteURL:   strings.TrimSpace(os.Getenv("TURSO_REMOTE_URL")),
		TursoAuthToken:   strings.TrimSpace(os.Getenv("TURSO_AUTH_TOKEN")),
	}
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
