package auth

import (
	"context"
	"net/http"
	"time"

	"cleartify/internal/domain"
)

const SessionCookieName = "cleartify_session"

type CallbackResult struct {
	SessionID string
	ReturnTo  string
}

type SessionRecord struct {
	ID                   string
	User                 domain.User
	Organization         *domain.Organization
	Workspace            *domain.Workspace
	NeedsOnboarding      bool
	RefreshToken         string
	WorkOSSessionID      string
	OrganizationID       string
	AccessTokenExpiresAt *time.Time
}

type SessionStore interface {
	UpsertAuthUser(ctx context.Context, user domain.User) error
	LookupAuthSession(ctx context.Context, sessionID string) (SessionRecord, error)
	UpsertAuthSession(ctx context.Context, session SessionRecord) error
	DeleteAuthSession(ctx context.Context, sessionID string) error
	MarkUserOnboardingComplete(ctx context.Context, userID string, completedAt time.Time) error
}

type Service interface {
	CurrentSession(ctx context.Context, r *http.Request) (domain.AuthSession, error)
	LoginURL(ctx context.Context, r *http.Request) (string, error)
	Authenticate(ctx context.Context, r *http.Request) (CallbackResult, error)
	LogoutURL(ctx context.Context, r *http.Request) (string, error)
	CompleteOnboarding(ctx context.Context, r *http.Request) (domain.AuthSession, error)
}
