package auth

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"cleartify/internal/auth/workosclient"
	"cleartify/internal/domain"
)

type fakeSessionStore struct {
	user              domain.User
	session           SessionRecord
	sessionErr        error
	upsertedUser      domain.User
	upsertedSession   SessionRecord
	deletedSessionID  string
	completedUserID   string
	completedAt       time.Time
	workspaceModeID   string
	workspaceMode     string
	modeSelectedAt    time.Time
	sampleDataEnsured bool
}

func (f *fakeSessionStore) UpsertAuthUser(_ context.Context, user domain.User) error {
	f.upsertedUser = user
	f.user = user
	return nil
}

func (f *fakeSessionStore) LookupAuthSession(context.Context, string) (SessionRecord, error) {
	if f.sessionErr != nil {
		return SessionRecord{}, f.sessionErr
	}
	return f.session, nil
}

func (f *fakeSessionStore) UpsertAuthSession(_ context.Context, session SessionRecord) error {
	f.upsertedSession = session
	f.session = session
	return nil
}

func (f *fakeSessionStore) DeleteAuthSession(_ context.Context, sessionID string) error {
	f.deletedSessionID = sessionID
	return nil
}

func (f *fakeSessionStore) MarkUserOnboardingComplete(_ context.Context, userID string, completedAt time.Time) error {
	f.completedUserID = userID
	f.completedAt = completedAt
	if f.session.User.ID == userID {
		f.session.NeedsOnboarding = false
	}
	return nil
}

func (f *fakeSessionStore) SetWorkspaceMode(_ context.Context, workspaceID string, mode string, selectedAt time.Time) error {
	f.workspaceModeID = workspaceID
	f.workspaceMode = mode
	f.modeSelectedAt = selectedAt
	if f.session.Workspace != nil && f.session.Workspace.ID == workspaceID {
		f.session.Workspace.Mode = mode
	}
	return nil
}

func (f *fakeSessionStore) EnsureSampleData() error {
	f.sampleDataEnsured = true
	return nil
}

func TestValidateWorkOSConfig(t *testing.T) {
	t.Parallel()

	if err := ValidateWorkOSConfig("api", "client", "http://localhost:3000/auth/callback", "cookie"); err != nil {
		t.Fatalf("expected valid config: %v", err)
	}
	if err := ValidateWorkOSConfig("", "client", "http://localhost:3000/auth/callback", "cookie"); err == nil {
		t.Fatal("expected api key validation error")
	}
	if err := ValidateWorkOSConfig("api", "", "http://localhost:3000/auth/callback", "cookie"); err == nil {
		t.Fatal("expected client id validation error")
	}
	if err := ValidateWorkOSConfig("api", "client", "", "cookie"); err == nil {
		t.Fatal("expected redirect uri validation error")
	}
	if err := ValidateWorkOSConfig("api", "client", "http://localhost:3000/auth/callback", ""); err == nil {
		t.Fatal("expected cookie password validation error")
	}
}

func TestNewWorkOSAuthServiceRejectsMissingStoreAndUnconfiguredClient(t *testing.T) {
	t.Parallel()

	if _, err := NewWorkOSAuthService("api", "client", "http://localhost:3000/auth/callback", "http://localhost:3000", "http://localhost:4173", "cookie", nil, nil); err == nil {
		t.Fatal("expected missing session store error")
	}
	if _, err := NewWorkOSAuthService(
		"api",
		"client",
		"http://localhost:3000/auth/callback",
		"http://localhost:3000",
		"http://localhost:4173",
		"cookie",
		&fakeSessionStore{},
		&workosclient.FakeClient{},
	); err == nil {
		t.Fatal("expected unconfigured client error")
	}
}

func TestWorkOSAuthServiceLoginURLUsesHostedAuthKit(t *testing.T) {
	t.Parallel()

	service, err := NewWorkOSAuthService(
		"api",
		"client",
		"http://localhost:3000/auth/callback",
		"http://localhost:3000",
		"http://localhost:4173",
		"cookie",
		&fakeSessionStore{},
		&workosclient.FakeClient{ConfiguredValue: true},
	)
	if err != nil {
		t.Fatalf("expected workos service to build: %v", err)
	}

	request := httptest.NewRequest("GET", "/login?returnTo=%2Freview%2Fwo_1&screenHint=sign-up", nil)
	loginURL, err := service.LoginURL(context.Background(), request)
	if err != nil {
		t.Fatalf("expected login url: %v", err)
	}

	if loginURL == "" || loginURL == "/" {
		t.Fatalf("expected hosted authkit url, got %q", loginURL)
	}
}

func TestWorkOSAuthServiceAuthenticateCreatesPersistedSession(t *testing.T) {
	t.Parallel()

	store := &fakeSessionStore{}
	client := &workosclient.FakeClient{
		ConfiguredValue: true,
		AuthenticateResultValue: workosclient.AuthenticateResult{
			UserID:         "user_123",
			Email:          "user@example.com",
			DisplayName:    "Test User",
			OrganizationID: "org_123",
			AccessToken:    testAccessToken("session_123", time.Date(2026, 4, 1, 17, 0, 0, 0, time.UTC)),
			RefreshToken:   "refresh_token",
		},
		OrganizationValue: workosclient.Organization{
			ID:   "org_123",
			Name: "Acme Product",
		},
	}

	service, err := NewWorkOSAuthService(
		"api",
		"client",
		"http://localhost:3000/auth/callback",
		"http://localhost:3000",
		"http://localhost:4173",
		"cookie",
		store,
		client,
	)
	if err != nil {
		t.Fatalf("expected workos service to build: %v", err)
	}

	state, err := service.stateCodec.Sign(StatePayload{Flow: "workos-authkit", ReturnTo: "/plans/wo_1"})
	if err != nil {
		t.Fatalf("expected state to sign: %v", err)
	}

	request := httptest.NewRequest("GET", "/auth/callback?code=abc123&state="+state, nil)
	request.RemoteAddr = "127.0.0.1:12345"
	request.Header.Set("User-Agent", "Cleartify/Test")

	result, err := service.Authenticate(context.Background(), request)
	if err != nil {
		t.Fatalf("expected callback auth to succeed: %v", err)
	}

	if result.SessionID == "" {
		t.Fatal("expected local session id")
	}
	if result.ReturnTo != "/plans/wo_1" {
		t.Fatalf("expected callback returnTo, got %q", result.ReturnTo)
	}
	if store.upsertedUser.Email != "user@example.com" {
		t.Fatalf("expected user upsert, got %#v", store.upsertedUser)
	}
	if store.upsertedSession.WorkOSSessionID != "session_123" {
		t.Fatalf("expected workos session id, got %#v", store.upsertedSession)
	}
	if store.upsertedSession.Workspace == nil || store.upsertedSession.Workspace.Name != "Acme Product" {
		t.Fatalf("expected workspace context, got %#v", store.upsertedSession.Workspace)
	}
}

func TestWorkOSAuthServiceCurrentSessionRefreshesExpiringTokens(t *testing.T) {
	t.Parallel()

	expiresSoon := time.Now().Add(-1 * time.Minute).UTC()
	store := &fakeSessionStore{
		session: SessionRecord{
			ID:                   "local_session",
			User:                 domain.User{ID: "user_123", Email: "user@example.com", DisplayName: "Test User"},
			Workspace:            &domain.Workspace{ID: "org_123", Name: "Acme Product", Slug: "acme-product"},
			NeedsOnboarding:      true,
			RefreshToken:         "refresh_token",
			WorkOSSessionID:      "session_123",
			OrganizationID:       "org_123",
			AccessTokenExpiresAt: &expiresSoon,
		},
	}
	client := &workosclient.FakeClient{
		ConfiguredValue: true,
		RefreshResultValue: workosclient.RefreshResult{
			AccessToken:  testAccessToken("session_456", time.Now().Add(30*time.Minute)),
			RefreshToken: "next_refresh_token",
		},
	}

	service, err := NewWorkOSAuthService(
		"api",
		"client",
		"http://localhost:3000/auth/callback",
		"http://localhost:3000",
		"http://localhost:4173",
		"cookie",
		store,
		client,
	)
	if err != nil {
		t.Fatalf("expected workos service to build: %v", err)
	}

	request := httptest.NewRequest("GET", "/api/auth/session", nil)
	request.AddCookie(&http.Cookie{Name: SessionCookieName, Value: "local_session"})
	request.RemoteAddr = "127.0.0.1:12345"

	session, err := service.CurrentSession(context.Background(), request)
	if err != nil {
		t.Fatalf("expected current session to succeed: %v", err)
	}
	if !session.Authenticated {
		t.Fatal("expected authenticated session")
	}
	if store.upsertedSession.RefreshToken != "next_refresh_token" {
		t.Fatalf("expected refreshed token to persist, got %#v", store.upsertedSession)
	}
}

func TestWorkOSAuthServiceCurrentSessionReturnsUnauthenticatedWhenMissing(t *testing.T) {
	t.Parallel()

	service, err := NewWorkOSAuthService(
		"api",
		"client",
		"http://localhost:3000/auth/callback",
		"http://localhost:3000",
		"http://localhost:4173",
		"cookie",
		&fakeSessionStore{sessionErr: sql.ErrNoRows},
		&workosclient.FakeClient{ConfiguredValue: true},
	)
	if err != nil {
		t.Fatalf("expected workos service to build: %v", err)
	}

	session, err := service.CurrentSession(context.Background(), httptest.NewRequest("GET", "/api/auth/session", nil))
	if err != nil {
		t.Fatalf("expected unauthenticated session, got error: %v", err)
	}
	if session.Authenticated {
		t.Fatal("expected unauthenticated session")
	}
}

func TestWorkOSAuthServiceCompleteOnboardingAndLogout(t *testing.T) {
	t.Parallel()

	store := &fakeSessionStore{
		session: SessionRecord{
			ID:              "local_session",
			User:            domain.User{ID: "user_123", Email: "user@example.com", DisplayName: "Test User"},
			Workspace:       &domain.Workspace{ID: "org_123", Name: "Acme Product", Slug: "acme-product"},
			NeedsOnboarding: true,
			WorkOSSessionID: "session_123",
		},
	}
	client := &workosclient.FakeClient{
		ConfiguredValue: true,
		LogoutURLValue:  "https://workos.example/logout",
	}

	service, err := NewWorkOSAuthService(
		"api",
		"client",
		"http://localhost:3000/auth/callback",
		"http://localhost:3000",
		"http://localhost:4173",
		"cookie",
		store,
		client,
	)
	if err != nil {
		t.Fatalf("expected workos service to build: %v", err)
	}

	request := httptest.NewRequest("POST", "/api/me/onboarding/complete", nil)
	request.AddCookie(&http.Cookie{Name: SessionCookieName, Value: "local_session"})

	session, err := service.CompleteOnboarding(context.Background(), request)
	if err != nil {
		t.Fatalf("expected onboarding completion to succeed: %v", err)
	}
	if session.NeedsOnboarding {
		t.Fatal("expected onboarding to be complete")
	}
	if store.completedUserID != "user_123" {
		t.Fatalf("expected onboarding persistence, got %q", store.completedUserID)
	}

	logoutURL, err := service.LogoutURL(context.Background(), request)
	if err != nil {
		t.Fatalf("expected logout url: %v", err)
	}
	if logoutURL != "https://workos.example/logout" {
		t.Fatalf("expected hosted logout url, got %q", logoutURL)
	}
	if store.deletedSessionID != "local_session" {
		t.Fatalf("expected local session delete, got %q", store.deletedSessionID)
	}
}

func testAccessToken(sessionID string, expiresAt time.Time) string {
	payload, _ := json.Marshal(map[string]any{
		"sid": sessionID,
		"exp": expiresAt.UTC().Unix(),
	})
	return "header." + base64.RawURLEncoding.EncodeToString(payload) + ".signature"
}
