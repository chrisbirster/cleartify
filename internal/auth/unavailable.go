package auth

import (
	"context"
	"net/http"
	"strings"

	"cleartify/internal/domain"
)

type UnavailableAuthService struct {
	appPublicURL string
	marketingURL string
}

func NewUnavailableAuthService(appPublicURL string, marketingURL string) *UnavailableAuthService {
	return &UnavailableAuthService{
		appPublicURL: strings.TrimRight(strings.TrimSpace(appPublicURL), "/"),
		marketingURL: strings.TrimRight(strings.TrimSpace(marketingURL), "/"),
	}
}

func (s *UnavailableAuthService) CurrentSession(context.Context, *http.Request) (domain.AuthSession, error) {
	return domain.AuthSession{Authenticated: false}, nil
}

func (s *UnavailableAuthService) LoginURL(context.Context, *http.Request) (string, error) {
	return authErrorURL(s.appPublicURL, "auth_unavailable"), ErrAuthUnavailable
}

func (s *UnavailableAuthService) Authenticate(context.Context, *http.Request) (CallbackResult, error) {
	return CallbackResult{}, ErrAuthUnavailable
}

func (s *UnavailableAuthService) LogoutURL(context.Context, *http.Request) (string, error) {
	return firstNonEmpty(s.marketingURL, s.appPublicURL, "/"), nil
}

func (s *UnavailableAuthService) CompleteOnboarding(context.Context, *http.Request) (domain.AuthSession, error) {
	return domain.AuthSession{}, ErrUnauthenticated
}
