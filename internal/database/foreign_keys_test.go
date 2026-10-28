package database

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
)

func TestForeignKeyEnforcement(t *testing.T) {
	ctx := context.Background()
	s, e := Open(ctx, Config{Path: ":memory:"})
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	var enabled int
	if e = s.DB.QueryRowContext(ctx, "PRAGMA foreign_keys").Scan(&enabled); e != nil || enabled != 1 {
		t.Fatalf("enabled=%d %v", enabled, e)
	}
	if _, e = s.DB.ExecContext(ctx, "CREATE TABLE parent(id TEXT PRIMARY KEY)"); e != nil {
		t.Fatal(e)
	}
	if _, e = s.DB.ExecContext(ctx, "CREATE TABLE child(parent_id TEXT REFERENCES parent(id))"); e != nil {
		t.Fatal(e)
	}
	if _, e = s.DB.ExecContext(ctx, "INSERT INTO child VALUES('missing')"); e == nil {
		t.Fatal("orphan accepted")
	}
}

func TestForeignKeysSurviveReplacementAndRejectHistoricalOrphans(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "keys.db")
	s, e := Open(ctx, Config{Path: path})
	if e != nil {
		t.Fatal(e)
	}
	if e = s.Migrate(ctx); e != nil {
		t.Fatal(e)
	}
	s.DB.SetMaxIdleConns(0)
	s.sourceDB.SetMaxIdleConns(0)
	for i := 0; i < 3; i++ {
		var enabled int
		if e = s.DB.QueryRowContext(ctx, "PRAGMA foreign_keys").Scan(&enabled); e != nil || enabled != 1 {
			t.Fatalf("replacement connection disabled: %d %v", enabled, e)
		}
	}
	if e = s.Close(); e != nil {
		t.Fatal(e)
	}
	raw, e := sql.Open("turso", path)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = raw.ExecContext(ctx, `INSERT INTO phone_identities(id,user_id,phone_e164,verified_at,created_at) VALUES('orphan','missing','+15550000001','now','now')`); e != nil {
		t.Fatal(e)
	}
	raw.Close()
	s, e = Open(ctx, Config{Path: path})
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	if e = s.Migrate(ctx); e == nil {
		t.Fatal("historical orphan database accepted")
	}
}
