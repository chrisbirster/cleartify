package auth

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"cleartify/internal/auth/workosclient"
	"cleartify/internal/domain"
)

type WorkOSAuthService struct {
	appPublicURL string
	marketingURL string
	redirectURI  string
	client       workosclient.Client
	stateCodec   *SignedStateCodec
	sessionStore SessionStore
}

func NewWorkOSAuthService(
	apiKey string,
	clientID string,
	redirectURI string,
	appPublicURL string,
	marketingURL string,
	cookiePassword string,
	sessionStore SessionStore,
	client workosclient.Client,
) (*WorkOSAuthService, error) {
	if err := ValidateWorkOSConfig(apiKey, clientID, redirectURI, cookiePassword); err != nil {
		return nil, err
	}
	if sessionStore == nil {
		return nil, fmt.Errorf("session store is required")
	}

	if client == nil {
		client = workosclient.New(apiKey, clientID)
	}
	if !client.Configured() {
		return nil, fmt.Errorf("workos client is not configured")
	}

	return &WorkOSAuthService{
		appPublicURL: strings.TrimRight(strings.TrimSpace(appPublicURL), "/"),
		marketingURL: strings.TrimRight(strings.TrimSpace(marketingURL), "/"),
		redirectURI:  strings.TrimSpace(redirectURI),
		client:       client,
		stateCodec:   NewSignedStateCodec(cookiePassword),
		sessionStore: sessionStore,
	}, nil
}

func ValidateWorkOSConfig(apiKey, clientID, redirectURI, cookiePassword string) error {
	switch {
	case strings.TrimSpace(apiKey) == "":
		return fmt.Errorf("WORKOS_API_KEY is required")
	case strings.TrimSpace(clientID) == "":
		return fmt.Errorf("WORKOS_CLIENT_ID is required")
	case strings.TrimSpace(redirectURI) == "":
		return fmt.Errorf("WORKOS_REDIRECT_URI is required")
	case strings.TrimSpace(cookiePassword) == "":
		return fmt.Errorf("WORKOS_COOKIE_PASSWORD is required")
	default:
		return nil
	}
}

func (s *WorkOSAuthService) CurrentSession(ctx context.Context, r *http.Request) (domain.AuthSession, error) {
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

	if needsRefresh(record.AccessTokenExpiresAt) && strings.TrimSpace(record.RefreshToken) != "" {
		refreshed, err := s.client.AuthenticateWithRefreshToken(ctx, workosclient.AuthenticateWithRefreshTokenOpts{
			RefreshToken:   record.RefreshToken,
			OrganizationID: record.OrganizationID,
			IPAddress:      requestIP(r),
			UserAgent:      r.UserAgent(),
		})
		if err != nil {
			return domain.AuthSession{Authenticated: false}, nil
		}

		workOSSessionID, expiresAt, err := parseAccessTokenClaims(refreshed.AccessToken)
		if err == nil {
			record.RefreshToken = refreshed.RefreshToken
			record.WorkOSSessionID = firstNonEmpty(workOSSessionID, record.WorkOSSessionID)
			record.AccessTokenExpiresAt = expiresAt
			if err := s.sessionStore.UpsertAuthSession(ctx, record); err != nil {
				return domain.AuthSession{}, err
			}
		}
	}

	return authSessionFromRecord(record), nil
}

func (s *WorkOSAuthService) LoginURL(_ context.Context, r *http.Request) (string, error) {
	returnTo := sanitizeReturnTo(r.URL.Query().Get("returnTo"))
	screenHint := normalizeScreenHint(r.URL.Query().Get("screenHint"))
	state, err := s.stateCodec.Sign(StatePayload{
		Flow:     "workos-authkit",
		ReturnTo: returnTo,
	})
	if err != nil {
		return "", err
	}

	return s.client.AuthorizationURL(workosclient.AuthorizationURLOpts{
		RedirectURI: s.redirectURI,
		State:       state,
		ScreenHint:  screenHint,
	})
}

func (s *WorkOSAuthService) Authenticate(ctx context.Context, r *http.Request) (CallbackResult, error) {
	code := strings.TrimSpace(r.URL.Query().Get("code"))
	state := strings.TrimSpace(r.URL.Query().Get("state"))
	if code == "" || state == "" {
		return CallbackResult{}, ErrInvalidCallback
	}

	payload, err := s.stateCodec.Verify(state, authNow())
	if err != nil {
		return CallbackResult{}, ErrInvalidCallback
	}

	result, err := s.client.AuthenticateWithCode(ctx, workosclient.AuthenticateWithCodeOpts{
		Code:      code,
		IPAddress: requestIP(r),
		UserAgent: r.UserAgent(),
	})
	if err != nil {
		return CallbackResult{}, err
	}

	workOSSessionID, expiresAt, err := parseAccessTokenClaims(result.AccessToken)
	if err != nil {
		return CallbackResult{}, err
	}

	user := domain.User{
		ID:          result.UserID,
		Email:       result.Email,
		DisplayName: result.DisplayName,
	}
	if err := s.sessionStore.UpsertAuthUser(ctx, user); err != nil {
		return CallbackResult{}, err
	}

	record := SessionRecord{
		ID:                   randomHex(18),
		User:                 user,
		NeedsOnboarding:      true,
		RefreshToken:         result.RefreshToken,
		WorkOSSessionID:      workOSSessionID,
		OrganizationID:       result.OrganizationID,
		AccessTokenExpiresAt: expiresAt,
	}
	if strings.TrimSpace(result.OrganizationID) != "" {
		org, err := s.client.GetOrganization(ctx, result.OrganizationID)
		if err != nil {
			return CallbackResult{}, err
		}
		record.Organization = &domain.Organization{
			ID:   org.ID,
			Name: org.Name,
			Slug: slugifyName(firstNonEmpty(org.Name, org.ID)),
		}
		record.Workspace = &domain.Workspace{
			ID:   org.ID,
			Name: org.Name,
			Slug: slugifyName(firstNonEmpty(org.Name, org.ID)),
		}
	} else {
		record.Workspace = &domain.Workspace{
			ID:   "workspace_" + result.UserID,
			Name: result.DisplayName + "'s workspace",
			Slug: slugifyName(result.DisplayName + " workspace"),
		}
	}

	if err := s.sessionStore.UpsertAuthSession(ctx, record); err != nil {
		return CallbackResult{}, err
	}

	return CallbackResult{
		SessionID: record.ID,
		ReturnTo:  sanitizeReturnTo(payload.ReturnTo),
	}, nil
}

func (s *WorkOSAuthService) LogoutURL(ctx context.Context, r *http.Request) (string, error) {
	sessionID := sessionIDFromRequest(r)
	if sessionID == "" {
		return firstNonEmpty(s.marketingURL, s.appPublicURL, "/"), nil
	}

	record, err := s.sessionStore.LookupAuthSession(ctx, sessionID)
	if err != nil && err != sql.ErrNoRows {
		return "", err
	}
	if err := s.sessionStore.DeleteAuthSession(ctx, sessionID); err != nil {
		return "", err
	}
	if err == sql.ErrNoRows || strings.TrimSpace(record.WorkOSSessionID) == "" {
		return firstNonEmpty(s.marketingURL, s.appPublicURL, "/"), nil
	}

	return s.client.LogoutURL(record.WorkOSSessionID, firstNonEmpty(s.marketingURL, s.appPublicURL, "/"))
}

func (s *WorkOSAuthService) CompleteOnboarding(ctx context.Context, r *http.Request) (domain.AuthSession, error) {
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

func normalizeScreenHint(raw string) string {
	switch strings.TrimSpace(raw) {
	case "sign-up":
		return "sign-up"
	default:
		return "sign-in"
	}
}

func needsRefresh(expiresAt *time.Time) bool {
	if expiresAt == nil {
		return false
	}
	return authNow().Add(1 * time.Minute).After(*expiresAt)
}

func slugifyName(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return ""
	}

	var builder strings.Builder
	lastDash := false
	for _, char := range value {
		switch {
		case char >= 'a' && char <= 'z':
			builder.WriteRune(char)
			lastDash = false
		case char >= '0' && char <= '9':
			builder.WriteRune(char)
			lastDash = false
		default:
			if !lastDash {
				builder.WriteRune('-')
				lastDash = true
			}
		}
	}

	return strings.Trim(builder.String(), "-")
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
