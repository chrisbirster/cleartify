package server

import (
	"net/http"
	"strings"
	"time"

	"cleartify/internal/auth"
)

func (s *Server) sessionCookie(sessionID string) *http.Cookie {
	return &http.Cookie{
		Name:     auth.SessionCookieName,
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   strings.HasPrefix(strings.ToLower(strings.TrimSpace(s.config.AppPublicURL)), "https://"),
		MaxAge:   60 * 60 * 24 * 30,
	}
}

func (s *Server) expiredSessionCookie() *http.Cookie {
	return &http.Cookie{
		Name:     auth.SessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   strings.HasPrefix(strings.ToLower(strings.TrimSpace(s.config.AppPublicURL)), "https://"),
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	}
}

func authErrorURL(appPublicURL string, code string) string {
	if strings.TrimSpace(appPublicURL) == "" {
		return "/?authError=" + code
	}
	separator := "/"
	if strings.HasSuffix(appPublicURL, "/") {
		separator = ""
	}
	return appPublicURL + separator + "?authError=" + code
}
