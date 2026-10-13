package api

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chrisbirster/cleartify/internal/database"
	"github.com/chrisbirster/cleartify/internal/waitlist"
)

func TestWaitlistEndpointPersistsSignupAndConsent(t *testing.T) {
	ctx := context.Background()
	store, err := database.Open(ctx, database.Config{Path: ":memory:"})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer store.Close()

	if err := store.Migrate(ctx); err != nil {
		t.Fatalf("migrate database: %v", err)
	}

	handler := New(waitlist.NewService(store), "http://localhost:3000")
	body := []byte(`{"name":"Chris","email":"Chris@example.com","phone":"(717) 555-1212","emailUpdates":true,"acceptPolicies":true}`)
	req := httptest.NewRequest(http.MethodPost, "/api/waitlist", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "http://localhost:3000")
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%s", res.Code, http.StatusCreated, res.Body.String())
	}
	if got := res.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Fatalf("CORS origin = %q", got)
	}

	var entries int
	if err := store.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM waitlist_entries`).Scan(&entries); err != nil {
		t.Fatalf("count waitlist entries: %v", err)
	}
	if entries != 1 {
		t.Fatalf("waitlist entries = %d, want 1", entries)
	}

	var email, phone string
	var updates int
	if err := store.DB.QueryRowContext(ctx,
		`SELECT email_normalized, phone_e164, email_updates FROM waitlist_entries LIMIT 1`,
	).Scan(&email, &phone, &updates); err != nil {
		t.Fatalf("read waitlist entry: %v", err)
	}
	if email != "chris@example.com" || phone != "+17175551212" || updates != 1 {
		t.Fatalf("stored entry = (%q, %q, %d)", email, phone, updates)
	}

	var consents int
	if err := store.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM consents`).Scan(&consents); err != nil {
		t.Fatalf("count consents: %v", err)
	}
	var acceptedVersion string
	if err := store.DB.QueryRowContext(ctx, `SELECT policy_version FROM policy_acceptances LIMIT 1`).Scan(&acceptedVersion); err != nil { t.Fatalf("read policy acceptance: %v",err) }
	if acceptedVersion != waitlist.CurrentPolicyVersion { t.Fatalf("policy version=%q",acceptedVersion) }
	if consents != 1 {
		t.Fatalf("consents = %d, want 1", consents)
	}
}

func TestWaitlistEndpointRejectsInvalidPhone(t *testing.T) {
	ctx := context.Background()
	store, err := database.Open(ctx, database.Config{Path: ":memory:"})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer store.Close()

	if err := store.Migrate(ctx); err != nil {
		t.Fatalf("migrate database: %v", err)
	}

	handler := New(waitlist.NewService(store), "")
	body := []byte(`{"name":"Chris","email":"chris@example.com","phone":"call-me-maybe","emailUpdates":false,"acceptPolicies":true}`)
	req := httptest.NewRequest(http.MethodPost, "/api/waitlist", bytes.NewReader(body))
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body=%s", res.Code, http.StatusBadRequest, res.Body.String())
	}
}
