package clear

import (
	"context"
	"errors"
	"github.com/chrisbirster/cleartify/internal/database"
	"testing"
)

func TestCanonicalInstallIdentityAndApprovalRevocation(t *testing.T) {
	ctx := context.Background()
	db, e := database.Open(ctx, database.Config{Path: ":memory:"})
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	if e = db.Migrate(ctx); e != nil {
		t.Fatal(e)
	}
	for _, q := range []string{
		`INSERT INTO users(id,display_name,primary_email,created_at,updated_at) VALUES('u','User','u@example.test','now','now')`,
		`INSERT INTO published_protocols(id,creator_did,slug,version,title,summary,license,spec_json,status,created_at) VALUES('p1','did:plc:a','same',1,'A','','MIT','{}','published','now'),('p2','did:plc:b','same',1,'B','','MIT','{}','published','now')`,
		`INSERT INTO protocol_installations(id,user_id,protocol_key,version,source,enabled,created_at) VALUES('legacy','u','same',1,'alpha',1,'now')`,
	} {
		if _, e = db.DB.ExecContext(ctx, q); e != nil {
			t.Fatal(e)
		}
	}
	svc := Service{DB: db.DB}
	for _, pub := range []string{"p1", "p2"} {
		if _, e = svc.Request(ctx, "u", pub); !errors.Is(e, ErrNotPublished) {
			t.Fatalf("legacy slug authorized %s: %v", pub, e)
		}
	}
	if _, e = db.DB.ExecContext(ctx, `UPDATE protocol_installations SET publication_id='p1',spec_sha256=? WHERE id='legacy'`, specDigest("{}")); e != nil {
		t.Fatal(e)
	}
	if _, e = svc.Request(ctx, "u", "p2"); !errors.Is(e, ErrNotPublished) {
		t.Fatal("other creator authorized", e)
	}
	id, e := svc.Request(ctx, "u", "p1")
	if e != nil {
		t.Fatal("canonical publication install unusable", e)
	}
	if _, e = db.DB.ExecContext(ctx, `UPDATE protocol_installations SET enabled=0`); e != nil {
		t.Fatal(e)
	}
	if e = svc.Decide(ctx, "u", id, true); !errors.Is(e, ErrNotPending) {
		t.Fatal("disabled install approved", e)
	}
	if _, e = db.DB.ExecContext(ctx, `UPDATE protocol_installations SET enabled=1`); e != nil {
		t.Fatal(e)
	}
	if _, e = db.DB.ExecContext(ctx, `UPDATE published_protocols SET spec_json='{"changed":true}' WHERE id='p1'`); e != nil {
		t.Fatal(e)
	}
	if e = svc.Decide(ctx, "u", id, true); !errors.Is(e, ErrNotPending) {
		t.Fatal("changed spec approved", e)
	}
	if _, e = svc.Request(ctx, "u", "p1"); !errors.Is(e, ErrNotPublished) {
		t.Fatal("changed installed spec requested", e)
	}
	if _, e = db.DB.ExecContext(ctx, `UPDATE published_protocols SET spec_json='{}' WHERE id='p1'`); e != nil {
		t.Fatal(e)
	}
	if e = svc.Decide(ctx, "u", id, true); e != nil {
		t.Fatal(e)
	}
}
