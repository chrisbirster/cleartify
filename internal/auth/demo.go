package auth

import (
	"context"
	"net/http"

	"cleartify/internal/domain"
)

var demoUser = &domain.User{
	ID:          "usr_demo",
	Email:       "demo@cleartify.com",
	DisplayName: "Demo User",
}

type DemoAuthService struct{}

func NewDemoAuthService() *DemoAuthService {
	return &DemoAuthService{}
}

func (s *DemoAuthService) CurrentSession(context.Context, *http.Request) (domain.AuthSession, error) {
	return domain.AuthSession{
		Authenticated: true,
		User:          demoUser,
	}, nil
}
