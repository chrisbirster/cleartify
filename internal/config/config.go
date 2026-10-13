package config

import (
	"os"
	"strings"
)

type Config struct {
 StrictProtocolPermissionsEnabled bool
 PrivateAlphaRoutesEnabled bool
 ClearSessionsEnabled bool
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
	ProtocolInstallFlowEnabled bool
	SocialProfileActionsEnabled bool
	AgentRegistrationEnabled bool
	AtprotoOAuthEnabled bool
	AtprotoPublicOrigin string
	AtprotoSessionKey string
}

func Load() Config {
	appOrigin := strings.TrimRight(env("APP_ORIGIN", "http://localhost:3000"), "/")
	return Config{
 StrictProtocolPermissionsEnabled: strings.EqualFold(strings.TrimSpace(os.Getenv("STRICT_PROTOCOL_PERMISSIONS_ENABLED")),"true"),
 PrivateAlphaRoutesEnabled: strings.EqualFold(strings.TrimSpace(os.Getenv("PRIVATE_ALPHA_ROUTES_ENABLED")),"true"),
 ClearSessionsEnabled: strings.EqualFold(strings.TrimSpace(os.Getenv("CLEAR_SESSIONS_ENABLED")),"true"),
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
		ProtocolInstallFlowEnabled: strings.EqualFold(strings.TrimSpace(os.Getenv("PROTOCOL_INSTALL_FLOW_ENABLED")),"true"),
		SocialProfileActionsEnabled: strings.EqualFold(strings.TrimSpace(os.Getenv("SOCIAL_PROFILE_ACTIONS_ENABLED")),"true"),
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
