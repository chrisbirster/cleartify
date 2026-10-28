package clear

import (
	"context"
	"errors"
	"github.com/chrisbirster/cleartify/internal/database"
	"testing"
)

func TestApprovalNeverExecutesAndCannotReplay(t *testing.T) {
	ctx := context.Background()
	db, err := database.Open(ctx, database.Config{Path: ":memory:"})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err = db.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	_, err = db.DB.ExecContext(ctx, `INSERT INTO users(id,display_name,primary_email,created_at,updated_at) VALUES('u1','Demo','a@example.com','now','now')`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.DB.ExecContext(ctx, `INSERT INTO published_protocols(id,creator_did,slug,version,title,summary,license,spec_json,status,created_at) VALUES('p1','did:plc:ewvi7nxzyoun6zhxrhs64oiz','weekly-reset',1,'Weekly','Weekly reset','MIT','{}','published','now')`)
	if err != nil {
		t.Fatal(err)
	}
	s := Service{DB: db.DB}
	if _, err = s.Request(ctx, "u1", "p1"); !errors.Is(err, ErrNotPublished) {
		t.Fatalf("uninstalled protocol allowed: %v", err)
	}
	_, err = db.DB.ExecContext(ctx, `INSERT INTO protocol_installations(id,user_id,protocol_key,version,source,created_at,publication_id,spec_sha256) VALUES('install','u1','p1',1,'alpha','now','p1','44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a')`)
	if err != nil {
		t.Fatal(err)
	}
	id, err := s.Request(ctx, "u1", "p1")
	if err != nil {
		t.Fatal(err)
	}
	if err = s.Decide(ctx, "u1", id, true); err != nil {
		t.Fatal(err)
	}
	if err = s.Decide(ctx, "u1", id, true); !errors.Is(err, ErrNotPending) {
		t.Fatalf("reapproval allowed: %v", err)
	}
	var status string
	if err = db.DB.QueryRowContext(ctx, "SELECT status FROM clear_run_requests WHERE id=?", id).Scan(&status); err != nil || status != "approved" {
		t.Fatalf("status %s %v", status, err)
	}
}
