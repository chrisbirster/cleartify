package store

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"cleartify/internal/auth"
	"cleartify/internal/domain"
)

func TestAuthSessionPersistenceAndOnboardingCompletion(t *testing.T) {
	t.Parallel()

	store := newTestStore(t, false)
	ctx := context.Background()

	if err := store.UpsertAuthUser(ctx, domain.User{
		ID:          "user_123",
		Email:       "user@example.com",
		DisplayName: "Test User",
	}); err != nil {
		t.Fatalf("expected auth user upsert: %v", err)
	}

	expiresAt := time.Date(2026, 4, 1, 17, 0, 0, 0, time.UTC)
	if err := store.UpsertAuthSession(ctx, auth.SessionRecord{
		ID:                   "session_123",
		User:                 domain.User{ID: "user_123", Email: "user@example.com", DisplayName: "Test User"},
		Workspace:            &domain.Workspace{ID: "org_123", Name: "Acme Product", Slug: "acme-product"},
		Organization:         &domain.Organization{ID: "org_123", Name: "Acme Product", Slug: "acme-product"},
		NeedsOnboarding:      true,
		RefreshToken:         "refresh_token",
		WorkOSSessionID:      "workos_session_123",
		OrganizationID:       "org_123",
		AccessTokenExpiresAt: &expiresAt,
	}); err != nil {
		t.Fatalf("expected auth session upsert: %v", err)
	}

	record, err := store.LookupAuthSession(ctx, "session_123")
	if err != nil {
		t.Fatalf("expected auth session lookup: %v", err)
	}
	if !record.NeedsOnboarding {
		t.Fatal("expected onboarding to be incomplete initially")
	}
	if record.Workspace == nil || record.Workspace.Name != "Acme Product" {
		t.Fatalf("expected workspace context, got %#v", record.Workspace)
	}

	if err := store.MarkUserOnboardingComplete(ctx, "user_123", time.Date(2026, 4, 1, 18, 0, 0, 0, time.UTC)); err != nil {
		t.Fatalf("expected onboarding completion to persist: %v", err)
	}

	record, err = store.LookupAuthSession(ctx, "session_123")
	if err != nil {
		t.Fatalf("expected auth session lookup after onboarding completion: %v", err)
	}
	if record.NeedsOnboarding {
		t.Fatal("expected onboarding completion to persist")
	}

	if err := store.DeleteAuthSession(ctx, "session_123"); err != nil {
		t.Fatalf("expected auth session delete: %v", err)
	}
	if _, err := store.LookupAuthSession(ctx, "session_123"); err != sql.ErrNoRows {
		t.Fatalf("expected deleted session to return sql.ErrNoRows, got %v", err)
	}
}
