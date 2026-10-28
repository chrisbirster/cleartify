package protocols

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/chrisbirster/cleartify/internal/database"
	"github.com/chrisbirster/cleartify/internal/security"
)

func TestTakedownAppliesAcrossDiscoveryAndInstallLifecycle(t *testing.T) {
	ctx := context.Background()
	db, err := database.Open(ctx, database.Config{Path: ":memory:"})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err = db.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	tokens, err := security.NewTokens("test-only-protocol-access-secret-000000")
	if err != nil {
		t.Fatal(err)
	}
	drafts := DraftService{DB: db.DB, Tokens: tokens, GuideStepsEnabled: true}
	owner := "did:plc:author"
	id, err := drafts.Create(ctx, owner, DraftInput{Slug: "guide", Version: 1, Title: "Guide", License: "MIT", Spec: json.RawMessage(`{"permissions":[],"steps":[{"operation":"guide","text":"Review next actions."}]}`)})
	if err != nil {
		t.Fatal(err)
	}
	installs := InstallService{Store: db}
	if err = drafts.PublishDraft(ctx, owner, id, "private", "original creator statement"); err != nil {
		t.Fatal(err)
	}
	if _, err = installs.Preview(ctx, id); !errors.Is(err, ErrInstallUnavailable) {
		t.Fatalf("private preview: %v", err)
	}
	if err = drafts.PublishDraft(ctx, owner, id, "unlisted", "original creator statement"); err != nil {
		t.Fatal(err)
	}
	if _, err = installs.Preview(ctx, id); err != nil {
		t.Fatalf("unlisted direct access: %v", err)
	}
	if results, err := (Reader{DB: db.DB}).Discover(ctx, 20); err != nil || len(results) != 0 {
		t.Fatalf("unlisted discovery: %v %v", results, err)
	}
	if err = drafts.PublishDraft(ctx, owner, id, "public", "original creator statement"); err != nil {
		t.Fatal(err)
	}
	manifest, err := installs.Preview(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	request, err := installs.Request(ctx, "did:plc:reader", id, manifest.ManifestSHA256)
	if err != nil {
		t.Fatal(err)
	}
	takedowns := Takedowns{DB: db.DB}
	takedown, err := takedowns.Request(ctx, owner, id, "rights_review")
	if err != nil {
		t.Fatal(err)
	}
	if err = takedowns.Decide(ctx, takedown, "hidden"); err != nil {
		t.Fatal(err)
	}
	for _, paid := range []bool{false, true} {
		for _, holds := range []bool{false, true} {
			results, err := (Reader{DB: db.DB, PaidEnforcement: paid, HoldVisibility: holds}).Discover(ctx, 20)
			if err != nil || len(results) != 0 {
				t.Fatalf("hidden discovery paid=%v holds=%v: %+v, %v", paid, holds, results, err)
			}
		}
		flow := InstallService{Store: db, PaidEnforcement: paid}
		if _, err := flow.Preview(ctx, id); !errors.Is(err, ErrInstallUnavailable) {
			t.Errorf("hidden preview paid=%v: %v", paid, err)
		}
		if _, err := flow.Request(ctx, "did:plc:reader", id, manifest.ManifestSHA256); !errors.Is(err, ErrInstallUnavailable) {
			t.Errorf("hidden request paid=%v: %v", paid, err)
		}
	}
	if err = installs.Decide(ctx, request, true); !errors.Is(err, ErrInstallDecision) {
		t.Fatalf("hidden approval: %v", err)
	}
	var status string
	if err = db.DB.QueryRowContext(ctx, "SELECT status FROM protocol_install_requests WHERE id=?", request).Scan(&status); err != nil || status != "pending_approval" {
		t.Fatalf("rejected approval changed state: %s, %v", status, err)
	}
	if err = takedowns.Decide(ctx, takedown, "restored"); err != nil {
		t.Fatal(err)
	}
	for _, paid := range []bool{false, true} {
		results, err := (Reader{DB: db.DB, PaidEnforcement: paid}).Discover(ctx, 20)
		if err != nil || len(results) != 1 {
			t.Fatalf("restored discovery: %v %v", results, err)
		}
	}
	if _, err = installs.Preview(ctx, id); err != nil {
		t.Fatalf("restored preview: %v", err)
	}
	if err = installs.Decide(ctx, request, true); err != nil {
		t.Fatalf("restored approval: %v", err)
	}
}
