package auth

import (
	"context"
	"database/sql"
	"net/http"
	"strings"
	"time"

	"cleartify/internal/domain"
)

type DemoModeAuthService struct {
	appPublicURL string
	marketingURL string
	stateCodec   *SignedStateCodec
	sessionStore SessionStore
}

func NewDemoModeAuthService(appPublicURL string, marketingURL string, cookiePassword string, sessionStore SessionStore) *DemoModeAuthService {
	return &DemoModeAuthService{
		appPublicURL: strings.TrimRight(strings.TrimSpace(appPublicURL), "/"),
		marketingURL: strings.TrimRight(strings.TrimSpace(marketingURL), "/"),
		stateCodec:   NewSignedStateCodec(firstNonEmpty(cookiePassword, "cleartify-demo-mode-secret")),
		sessionStore: sessionStore,
	}
}

func (s *DemoModeAuthService) CurrentSession(ctx context.Context, r *http.Request) (domain.AuthSession, error) {
	sessionID := sessionIDFromRequest(r)
	if sessionID == "" {
		return domain.AuthSession{Authenticated: false}, nil
	}

	record, err := s.sessionStore.LookupAuthSession(ctx, sessionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.AuthSession{Authenticated: false}, nil
		}
		return domain.AuthSession{}, err
	}

	return authSessionFromRecord(record), nil
}

func (s *DemoModeAuthService) LoginURL(_ context.Context, r *http.Request) (string, error) {
	state, err := s.stateCodec.Sign(StatePayload{
		Flow:     "demo-auth",
		ReturnTo: sanitizeReturnTo(r.URL.Query().Get("returnTo")),
	})
	if err != nil {
		return "", err
	}
	return s.appPublicURL + "/auth/callback?demo=1&state=" + state, nil
}

func (s *DemoModeAuthService) Authenticate(ctx context.Context, r *http.Request) (CallbackResult, error) {
	state := strings.TrimSpace(r.URL.Query().Get("state"))
	if strings.TrimSpace(r.URL.Query().Get("demo")) != "1" || state == "" {
		return CallbackResult{}, ErrInvalidCallback
	}

	payload, err := s.stateCodec.Verify(state, authNow())
	if err != nil {
		return CallbackResult{}, ErrInvalidCallback
	}

	user := domain.User{
		ID:          demoUser.ID,
		Email:       demoUser.Email,
		DisplayName: demoUser.DisplayName,
	}
	if err := s.sessionStore.UpsertAuthUser(ctx, user); err != nil {
		return CallbackResult{}, err
	}

	record := SessionRecord{
		ID:              randomHex(18),
		User:            user,
		Workspace:       demoWorkspace(),
		NeedsOnboarding: true,
	}
	if err := s.sessionStore.UpsertAuthSession(ctx, record); err != nil {
		return CallbackResult{}, err
	}

	return CallbackResult{
		SessionID: record.ID,
		ReturnTo:  sanitizeReturnTo(payload.ReturnTo),
	}, nil
}

func (s *DemoModeAuthService) LogoutURL(ctx context.Context, r *http.Request) (string, error) {
	sessionID := sessionIDFromRequest(r)
	if sessionID != "" {
		if err := s.sessionStore.DeleteAuthSession(ctx, sessionID); err != nil && err != sql.ErrNoRows {
			return "", err
		}
	}
	return firstNonEmpty(s.marketingURL, s.appPublicURL, "/"), nil
}

func (s *DemoModeAuthService) CompleteOnboarding(ctx context.Context, r *http.Request) (domain.AuthSession, error) {
	sessionID := sessionIDFromRequest(r)
	if sessionID == "" {
		return domain.AuthSession{}, ErrUnauthenticated
	}

	record, err := s.sessionStore.LookupAuthSession(ctx, sessionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.AuthSession{}, ErrUnauthenticated
		}
		return domain.AuthSession{}, err
	}

	if !record.NeedsOnboarding {
		return authSessionFromRecord(record), nil
	}

	if err := s.sessionStore.MarkUserOnboardingComplete(ctx, record.User.ID, authNow().UTC().Truncate(time.Second)); err != nil {
		if err == sql.ErrNoRows {
			return domain.AuthSession{}, ErrUnauthenticated
		}
		return domain.AuthSession{}, err
	}

	record.NeedsOnboarding = false
	return authSessionFromRecord(record), nil
}
