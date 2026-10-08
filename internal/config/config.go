package config

import (
	"os"
	"strings"
)

type Config struct {
	Port           string
	AppOrigin      string
	DatabasePath   string
	TursoRemoteURL string
	TursoAuthToken string
}

func Load() Config {
	return Config{
		Port:           env("PORT", "8080"),
		AppOrigin:      strings.TrimRight(env("APP_ORIGIN", "http://localhost:3000"), "/"),
		DatabasePath:   env("DATABASE_PATH", "./data/cleartify.db"),
		TursoRemoteURL: strings.TrimSpace(os.Getenv("TURSO_REMOTE_URL")),
		TursoAuthToken: strings.TrimSpace(os.Getenv("TURSO_AUTH_TOKEN")),
	}
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
