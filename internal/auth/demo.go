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
	workspace := *demoWorkspace()
	workspace.Mode = "live"

	return domain.AuthSession{
		Authenticated:   true,
		User:            demoUser,
		Workspace:       &workspace,
		NeedsOnboarding: false,
	}, nil
}

func (s *DemoAuthService) LoginURL(context.Context, *http.Request) (string, error) {
	return "/", nil
}

func (s *DemoAuthService) Authenticate(context.Context, *http.Request) (CallbackResult, error) {
	return CallbackResult{
		SessionID: "ses_demo",
		ReturnTo:  "/",
	}, nil
}

func (s *DemoAuthService) LogoutURL(context.Context, *http.Request) (string, error) {
	return "/", nil
}

func (s *DemoAuthService) CompleteOnboarding(context.Context, *http.Request) (domain.AuthSession, error) {
	return s.CurrentSession(context.Background(), nil)
}

func (s *DemoAuthService) SelectWorkspaceMode(context.Context, *http.Request, string) (domain.AuthSession, error) {
	return s.CurrentSession(context.Background(), nil)
}
