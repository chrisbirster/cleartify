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
	SocialDefaultFollowEnabled bool
	SocialAccountRegistrationEnabled bool
	AgentRegistrationEnabled bool
	AtprotoOAuthEnabled bool
	AtprotoPublicOrigin string
	AtprotoSessionKey string
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
		SocialDefaultFollowEnabled: strings.EqualFold(strings.TrimSpace(os.Getenv("SOCIAL_DEFAULT_FOLLOW_ENABLED")),"true"),
		SocialAccountRegistrationEnabled: strings.EqualFold(strings.TrimSpace(os.Getenv("SOCIAL_ACCOUNT_REGISTRATION_ENABLED")),"true"),
		AgentRegistrationEnabled: strings.EqualFold(strings.TrimSpace(os.Getenv("AGENT_REGISTRATION_ENABLED")),"true"),
		AtprotoOAuthEnabled: strings.EqualFold(strings.TrimSpace(os.Getenv("ATPROTO_OAUTH_ENABLED")),"true"),
		AtprotoPublicOrigin: strings.TrimSpace(os.Getenv("ATPROTO_PUBLIC_ORIGIN")),
		AtprotoSessionKey: strings.TrimSpace(os.Getenv("ATPROTO_SESSION_KEY_B64")),
	}
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
