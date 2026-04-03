package auth

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"cleartify/internal/domain"
)

var authNow = time.Now

var ErrUnauthenticated = errors.New("unauthenticated")
var ErrInvalidCallback = errors.New("invalid auth callback")
var ErrAuthUnavailable = errors.New("auth unavailable")

type accessTokenClaims struct {
	SessionID string `json:"sid"`
	ExpiresAt int64  `json:"exp"`
}

func sessionIDFromRequest(r *http.Request) string {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(cookie.Value)
}

func authSessionFromRecord(record SessionRecord) domain.AuthSession {
	return domain.AuthSession{
		Authenticated:   true,
		User:            &record.User,
		Organization:    record.Organization,
		Workspace:       record.Workspace,
		NeedsOnboarding: record.NeedsOnboarding,
	}
}

func sanitizeReturnTo(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "/"
	}
	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		return "/"
	}
	if !strings.HasPrefix(raw, "/") || strings.HasPrefix(raw, "//") {
		return "/"
	}
	return raw
}

func authErrorURL(appPublicURL string, code string) string {
	u, err := url.Parse(strings.TrimSpace(appPublicURL))
	if err != nil {
		return "/?authError=" + url.QueryEscape(code)
	}
	u.Path = "/"
	query := u.Query()
	query.Set("authError", strings.TrimSpace(code))
	u.RawQuery = query.Encode()
	return u.String()
}

func requestIP(r *http.Request) string {
	forwardedFor := strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-For"), ",")[0])
	if forwardedFor != "" {
		return forwardedFor
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil {
		return host
	}
	return strings.TrimSpace(r.RemoteAddr)
}

func parseAccessTokenClaims(token string) (string, *time.Time, error) {
	parts := strings.Split(strings.TrimSpace(token), ".")
	if len(parts) != 3 {
		return "", nil, fmt.Errorf("invalid access token")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", nil, err
	}

	var claims accessTokenClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return "", nil, err
	}

	var expiresAt *time.Time
	if claims.ExpiresAt > 0 {
		value := time.Unix(claims.ExpiresAt, 0).UTC()
		expiresAt = &value
	}

	return strings.TrimSpace(claims.SessionID), expiresAt, nil
}

func demoWorkspace() *domain.Workspace {
	return &domain.Workspace{
		ID:   "wrk_demo",
		Name: "Cleartify Demo Team",
		Slug: "cleartify-demo-team",
	}
}
