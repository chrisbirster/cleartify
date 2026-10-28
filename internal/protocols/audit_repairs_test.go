package protocols

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/chrisbirster/cleartify/internal/database"
	"github.com/chrisbirster/cleartify/internal/security"
	"testing"
)

func TestGuideAndRemixManifestsAcrossValidationFlagsAndCreatorSync(t *testing.T) {
	ctx := context.Background()
	db, e := database.Open(ctx, database.Config{Path: ":memory:"})
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	if e = db.Migrate(ctx); e != nil {
		t.Fatal(e)
	}
	tokens, _ := security.NewTokens("test-guide-compatibility-secret-000000")
	pushes := 0
	drafts := DraftService{DB: db.DB, Tokens: tokens, GuideStepsEnabled: true, Push: func(context.Context) error { pushes++; return nil }}
	parent := ""
	for version := 1; version <= 2; version++ {
		spec := `{"steps":[{"operation":"guide","text":"Review next actions."}]}`
		if version == 2 {
			spec = `{"steps":[{"operation":"guide","text":"Review next actions."}],"remixOf":"` + parent + `","attribution":"Original creator"}`
		}
		id, e := drafts.Create(ctx, "did:plc:creator", DraftInput{Slug: "guide", Version: version, Title: "Guide", License: "CC-BY-4.0", Spec: json.RawMessage(spec)})
		if e != nil {
			t.Fatal(e)
		}
		for _, visibility := range []string{"private", "public"} {
			if e = drafts.PublishDraft(ctx, "did:plc:creator", id, visibility, "creator original provenance"); e != nil {
				t.Fatal(e)
			}
		}
		for _, strict := range []bool{false, true} {
			for _, analysis := range []bool{false, true} {
				flow := InstallService{Store: db, StrictPermissions: strict, PermissionDiffEnabled: analysis}
				m, e := flow.Preview(ctx, id)
				if e != nil {
					t.Fatalf("version %d strict=%v analysis=%v: %v", version, strict, analysis, e)
				}
				if m.Permissions == nil || len(m.Permissions) != 0 {
					t.Fatalf("noncanonical empty permissions: %+v", m)
				}
				if version == 2 && analysis && (m.PermissionChanges == nil || m.PermissionChanges.Changed) {
					t.Fatalf("empty previous permissions diff: %+v", m)
				}
				req, e := flow.Request(ctx, "did:plc:reader", id, m.ManifestSHA256)
				if e != nil {
					t.Fatal(e)
				}
				if e = flow.Decide(ctx, req, true); e != nil {
					t.Fatal(e)
				}
			}
		}
		parent = id
	}
	if pushes != 6 {
		t.Fatalf("creator writes not pushed: %d", pushes)
	}
	m, e := (InstallService{Store: db}).Preview(ctx, parent)
	if e != nil {
		t.Fatal(e)
	}
	heldFlow := InstallService{Store: db, HoldVisibility: true}
	holdRequest, e := heldFlow.Request(ctx, "did:plc:held-reader", parent, m.ManifestSHA256)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = db.DB.ExecContext(ctx, `INSERT INTO social_profiles(did,handle,display_name,kind,claimed,public_visible,updated_at) VALUES('did:plc:creator','creator','Creator','person',1,1,'now')`); e != nil {
		t.Fatal(e)
	}
	if _, e = db.DB.ExecContext(ctx, `INSERT INTO social_impersonation_holds(subject_did,reason,created_at) VALUES('did:plc:creator','reports','now')`); e != nil {
		t.Fatal(e)
	}
	if _, e = heldFlow.Preview(ctx, parent); !errors.Is(e, ErrInstallUnavailable) {
		t.Fatal("held creator previewed", e)
	}
	if e = heldFlow.Decide(ctx, holdRequest, true); !errors.Is(e, ErrInstallDecision) {
		t.Fatal("held creator approved", e)
	}
	if _, e = db.DB.ExecContext(ctx, `UPDATE social_impersonation_holds SET released_at='now'`); e != nil {
		t.Fatal(e)
	}
	if e = heldFlow.Decide(ctx, holdRequest, true); e != nil {
		t.Fatal("released creator inaccessible", e)
	}
	pending, e := (InstallService{Store: db}).Request(ctx, "did:plc:blocked-reader", parent, m.ManifestSHA256)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = db.DB.ExecContext(ctx, `INSERT INTO social_blocks(actor_did,subject_did,created_at) VALUES('did:plc:creator','did:plc:blocked-reader','now')`); e != nil {
		t.Fatal(e)
	}
	if _, e = (InstallService{Store: db}).Request(ctx, "did:plc:blocked-reader", parent, m.ManifestSHA256); !errors.Is(e, ErrInstallUnavailable) {
		t.Fatal("blocked installation request accepted", e)
	}
	if e = (InstallService{Store: db}).Decide(ctx, pending, true); !errors.Is(e, ErrInstallDecision) {
		t.Fatal("block did not revoke pending approval", e)
	}
	takedowns := Takedowns{DB: db.DB, Push: drafts.Push}
	id, e := takedowns.Request(ctx, "did:plc:creator", parent, "rights_review")
	if e != nil {
		t.Fatal(e)
	}
	if e = takedowns.Decide(ctx, id, "hidden"); e != nil {
		t.Fatal(e)
	}
	if pushes != 8 {
		t.Fatalf("takedown writes not pushed: %d", pushes)
	}
	drafts.Push = func(context.Context) error { return errors.New("injected sync failure") }
	if _, e = drafts.Create(ctx, "did:plc:creator", DraftInput{Slug: "failed-sync", Version: 1, Title: "Guide", License: "CC-BY-4.0", Spec: json.RawMessage(`{"steps":[]}`)}); e == nil {
		t.Fatal("sync failure reported success")
	}
}

func TestCreatorSyncFailureRetriesDoNotDuplicateVersionsOrAudit(t *testing.T) {
	ctx := context.Background()
	db, e := database.Open(ctx, database.Config{Path: ":memory:"})
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	if e = db.Migrate(ctx); e != nil {
		t.Fatal(e)
	}
	tokens, _ := security.NewTokens("test-creator-sync-retry-secret-000000")
	fail := true
	push := func(context.Context) error {
		if fail {
			return errors.New("injected sync failure")
		}
		return nil
	}
	drafts := DraftService{DB: db.DB, Tokens: tokens, Push: push}
	input := DraftInput{Slug: "retry", Version: 1, Title: "Retry", License: "MIT", Spec: json.RawMessage(`{"steps":[]}`)}
	if _, e = drafts.Create(ctx, "did:plc:creator", input); e == nil {
		t.Fatal("failed push acknowledged")
	}
	fail = false
	id, e := drafts.Create(ctx, "did:plc:creator", input)
	if e != nil {
		t.Fatal("create recovery", e)
	}
	for _, target := range []string{"private", "public"} {
		fail = true
		if e = drafts.PublishDraft(ctx, "did:plc:creator", id, target, "creator original"); e == nil {
			t.Fatal("failed publication push acknowledged")
		}
		fail = false
		if e = drafts.PublishDraft(ctx, "did:plc:creator", id, target, "creator original"); e != nil {
			t.Fatal("publication recovery", e)
		}
	}
	takedowns := Takedowns{DB: db.DB, Push: push}
	fail = true
	if _, e = takedowns.Request(ctx, "did:plc:creator", id, "rights_review"); e == nil {
		t.Fatal("failed takedown push acknowledged")
	}
	fail = false
	request, e := takedowns.Request(ctx, "did:plc:creator", id, "rights_review")
	if e != nil {
		t.Fatal("request recovery", e)
	}
	fail = true
	if e = takedowns.Decide(ctx, request, "hidden"); e == nil {
		t.Fatal("failed decision push acknowledged")
	}
	fail = false
	if e = takedowns.Decide(ctx, request, "hidden"); e != nil {
		t.Fatal("decision recovery", e)
	}
	for table, want := range map[string]int{"published_protocols": 1, "creator_publication_audit": 2, "creator_takedown_requests": 1, "creator_takedown_audit": 2} {
		var n int
		if e = db.DB.QueryRowContext(ctx, "SELECT count(*) FROM "+table).Scan(&n); e != nil || n != want {
			t.Fatalf("duplicate or missing %s: %d %v", table, n, e)
		}
	}
}
