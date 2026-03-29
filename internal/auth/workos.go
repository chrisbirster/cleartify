package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"cleartify/internal/domain"
	"cleartify/internal/auth/workosclient"
)

type WorkOSAuthService struct {
	client workosclient.Client
}

func NewWorkOSAuthService(apiKey, clientID, cookiePassword string, client workosclient.Client) (*WorkOSAuthService, error) {
	if err := ValidateWorkOSConfig(apiKey, clientID, cookiePassword); err != nil {
		return nil, err
	}

	if client == nil {
		client = workosclient.New(apiKey, clientID)
	}
	if !client.Configured() {
		return nil, fmt.Errorf("workos client is not configured")
	}

	return &WorkOSAuthService{client: client}, nil
}

func ValidateWorkOSConfig(apiKey, clientID, cookiePassword string) error {
	switch {
	case strings.TrimSpace(apiKey) == "":
		return fmt.Errorf("WORKOS_API_KEY is required")
	case strings.TrimSpace(clientID) == "":
		return fmt.Errorf("WORKOS_CLIENT_ID is required")
	case strings.TrimSpace(cookiePassword) == "":
		return fmt.Errorf("WORKOS_COOKIE_PASSWORD is required")
	default:
		return nil
	}
}

func (s *WorkOSAuthService) CurrentSession(context.Context, *http.Request) (domain.AuthSession, error) {
	return domain.AuthSession{Authenticated: false}, nil
}
