package database

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
)

func TestVersionedMigrationUpgradesBaselineAndIsIdempotent(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, Config{Path: filepath.Join(t.TempDir(), "upgrade.db")})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	// Simulate a database created by the previous unversioned application.
	for _, statement := range schemaStatements {
		if _, err := db.DB.ExecContext(ctx, statement); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 2; i++ {
		if err := db.Migrate(ctx); err != nil {
			t.Fatal(err)
		}
	}
	var count int
	if err := db.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM schema_migrations WHERE version=1`).Scan(&count); err != nil || count != 1 {
		t.Fatalf("migration count=%d: %v", count, err)
	}
	if err := db.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM beta_invite_identities`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("unexpected identity backfill count=%d: %v", count, err)
	}
}

func TestMigrationFailureDoesNotRecordVersion(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, Config{Path: ":memory:"})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	// A conflicting preexisting trigger makes the migration fail after its CREATE TABLE.
	if _, err := db.DB.ExecContext(ctx, `CREATE TABLE sentinel(id INTEGER)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.DB.ExecContext(ctx, `CREATE TRIGGER beta_invite_identity_no_update BEFORE UPDATE ON sentinel BEGIN SELECT 1; END`); err != nil {
		t.Fatal(err)
	}
	if err := db.Migrate(ctx); err == nil {
		t.Fatal("expected migration failure")
	}
	var count int
	if err := db.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM schema_migrations`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("failed version recorded: %d, %v", count, err)
	}
	if err := db.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM sqlite_master WHERE name='beta_invite_identities'`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("partial migration survived: %d, %v", count, err)
	}
	if _, err := db.DB.ExecContext(ctx, `DROP TRIGGER beta_invite_identity_no_update`); err != nil {
		t.Fatal(err)
	}
	if err := db.Migrate(ctx); err != nil {
		t.Fatalf("retry migration: %v", err)
	}
}

func TestUpgradeDoesNotGuessLegacyInstallationOwnership(t *testing.T) {
	ctx := context.Background()
	db, e := Open(ctx, Config{Path: filepath.Join(t.TempDir(), "legacy.db")})
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	for _, statement := range schemaStatements {
		if _, e = db.DB.ExecContext(ctx, statement); e != nil {
			t.Fatal(e)
		}
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
	if e = db.Migrate(ctx); e != nil {
		t.Fatal(e)
	}
	var publication sql.NullString
	var digest string
	if e = db.DB.QueryRowContext(ctx, `SELECT publication_id,spec_sha256 FROM protocol_installations WHERE id='legacy'`).Scan(&publication, &digest); e != nil {
		t.Fatal(e)
	}
	if publication.Valid || digest != "" {
		t.Fatal("upgrade guessed authorization", publication, digest)
	}
	if e = db.Migrate(ctx); e != nil {
		t.Fatal("repeat upgrade", e)
	}
}
