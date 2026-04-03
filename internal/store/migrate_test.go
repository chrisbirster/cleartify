package store

import (
	"context"
	"database/sql"
	"errors"
	"testing"
)

func TestMigrateIsIdempotentAndCreatesTables(t *testing.T) {
	t.Parallel()

	store := newTestStore(t, false)
	if err := store.Migrate(); err != nil {
		t.Fatalf("expected second migrate pass to succeed: %v", err)
	}

	ctx := context.Background()
	for _, table := range []string{
		"migrations",
		"users",
		"sessions",
		"signals",
		"clusters",
		"cluster_signals",
		"work_objects",
		"draft_plans",
		"decisions",
		"artifacts",
	} {
		if _, err := countRows(ctx, store.DB, table); err != nil {
			t.Fatalf("expected table %s to exist: %v", table, err)
		}
	}

	hasClusterID, err := tableHasColumn(store.DB, "work_objects", "cluster_id")
	if err != nil {
		t.Fatalf("expected work_objects table info query: %v", err)
	}
	if !hasClusterID {
		t.Fatal("expected work_objects.cluster_id column to exist")
	}

	hasPrototypeNote, err := tableHasColumn(store.DB, "work_objects", "prototype_note")
	if err != nil {
		t.Fatalf("expected work_objects table info query: %v", err)
	}
	if !hasPrototypeNote {
		t.Fatal("expected work_objects.prototype_note column to exist")
	}

	hasReleaseReadyAt, err := tableHasColumn(store.DB, "work_objects", "release_ready_at")
	if err != nil {
		t.Fatalf("expected work_objects table info query: %v", err)
	}
	if !hasReleaseReadyAt {
		t.Fatal("expected work_objects.release_ready_at column to exist")
	}

	for _, columnCheck := range []struct {
		table  string
		column string
	}{
		{table: "users", column: "onboarding_completed_at"},
		{table: "sessions", column: "refresh_token"},
		{table: "sessions", column: "workos_session_id"},
		{table: "sessions", column: "organization_id"},
		{table: "sessions", column: "organization_name"},
		{table: "sessions", column: "access_token_expires_at"},
	} {
		hasColumn, err := tableHasColumn(store.DB, columnCheck.table, columnCheck.column)
		if err != nil {
			t.Fatalf("expected %s.%s table info query: %v", columnCheck.table, columnCheck.column, err)
		}
		if !hasColumn {
			t.Fatalf("expected %s.%s column to exist", columnCheck.table, columnCheck.column)
		}
	}
}

func TestMigrateReturnsStatementError(t *testing.T) {
	t.Parallel()

	store := openRawStore(t)
	if err := store.Close(); err != nil {
		t.Fatalf("expected close to succeed: %v", err)
	}

	if err := store.Migrate(); err == nil {
		t.Fatal("expected migration statement error")
	}
}

func TestMigrateReturnsRecordErrorWhenMigrationsTableIsMissing(t *testing.T) {
	originalStatements := phase0MigrationStatements
	t.Cleanup(func() { phase0MigrationStatements = originalStatements })

	phase0MigrationStatements = []string{
		`CREATE TABLE IF NOT EXISTS work_objects (id TEXT PRIMARY KEY, title TEXT, summary TEXT, goal TEXT, status TEXT, priority TEXT, owner_type TEXT, owner_id TEXT, prototype_url TEXT, prototype_note TEXT, release_ready_at TEXT, created_at TEXT, updated_at TEXT);`,
	}
	store := openRawStore(t)

	if err := store.Migrate(); err == nil {
		t.Fatal("expected record migration error")
	}
}

func TestEnsureNullableTextColumnReturnsTableInfoError(t *testing.T) {
	t.Parallel()

	store := openRawStore(t)
	if err := store.Close(); err != nil {
		t.Fatalf("expected close to succeed: %v", err)
	}
	if err := ensureNullableTextColumn(store.DB, "work_objects", "cluster_id"); err == nil {
		t.Fatal("expected query error")
	}
}

func TestEnsureNullableTextColumnAddsMissingColumn(t *testing.T) {
	t.Parallel()

	store := openRawStore(t)
	mustExec(t, store.DB, `CREATE TABLE work_objects (id TEXT PRIMARY KEY);`)
	mustExec(t, store.DB, `CREATE TABLE users (id TEXT PRIMARY KEY, email TEXT, display_name TEXT, created_at TEXT);`)
	mustExec(t, store.DB, `CREATE TABLE sessions (id TEXT PRIMARY KEY, user_id TEXT, created_at TEXT);`)

	if err := ensureNullableTextColumn(store.DB, "work_objects", "cluster_id"); err != nil {
		t.Fatalf("expected cluster_id column to be added: %v", err)
	}

	hasClusterID, err := tableHasColumn(store.DB, "work_objects", "cluster_id")
	if err != nil {
		t.Fatalf("expected work_objects table info query: %v", err)
	}
	if !hasClusterID {
		t.Fatal("expected work_objects.cluster_id column to exist")
	}

	if err := ensureNullableTextColumn(store.DB, "work_objects", "prototype_note"); err != nil {
		t.Fatalf("expected prototype_note column to be added: %v", err)
	}

	hasPrototypeNote, err := tableHasColumn(store.DB, "work_objects", "prototype_note")
	if err != nil {
		t.Fatalf("expected work_objects table info query: %v", err)
	}
	if !hasPrototypeNote {
		t.Fatal("expected work_objects.prototype_note column to exist")
	}

	if err := ensureNullableTextColumn(store.DB, "work_objects", "release_ready_at"); err != nil {
		t.Fatalf("expected release_ready_at column to be added: %v", err)
	}

	hasReleaseReadyAt, err := tableHasColumn(store.DB, "work_objects", "release_ready_at")
	if err != nil {
		t.Fatalf("expected work_objects table info query: %v", err)
	}
	if !hasReleaseReadyAt {
		t.Fatal("expected work_objects.release_ready_at column to exist")
	}

	if err := ensureNullableTextColumn(store.DB, "users", "onboarding_completed_at"); err != nil {
		t.Fatalf("expected users.onboarding_completed_at column to be added: %v", err)
	}
	hasOnboardingCompletedAt, err := tableHasColumn(store.DB, "users", "onboarding_completed_at")
	if err != nil {
		t.Fatalf("expected users table info query: %v", err)
	}
	if !hasOnboardingCompletedAt {
		t.Fatal("expected users.onboarding_completed_at column to exist")
	}

	for _, column := range []string{
		"refresh_token",
		"workos_session_id",
		"organization_id",
		"organization_name",
		"access_token_expires_at",
	} {
		if err := ensureNullableTextColumn(store.DB, "sessions", column); err != nil {
			t.Fatalf("expected sessions.%s column to be added: %v", column, err)
		}
		hasColumn, err := tableHasColumn(store.DB, "sessions", column)
		if err != nil {
			t.Fatalf("expected sessions table info query for %s: %v", column, err)
		}
		if !hasColumn {
			t.Fatalf("expected sessions.%s column to exist", column)
		}
	}
}

func TestMigrateReturnsPhase1RecordError(t *testing.T) {
	originalRecordMigration := recordMigration
	t.Cleanup(func() { recordMigration = originalRecordMigration })

	callCount := 0
	recordMigration = func(_ *sql.DB, name, _ string) error {
		callCount++
		if name == "phase1_work_objects_cluster_id" {
			return errors.New("phase1 failed")
		}
		return nil
	}

	store := openRawStore(t)
	if err := store.Migrate(); err == nil {
		t.Fatal("expected phase1 migration record error")
	}
	if callCount != 2 {
		t.Fatalf("expected 2 record migration calls, got %d", callCount)
	}
}

func TestMigrateReturnsPhase3RecordError(t *testing.T) {
	originalRecordMigration := recordMigration
	t.Cleanup(func() { recordMigration = originalRecordMigration })

	callCount := 0
	recordMigration = func(_ *sql.DB, name, _ string) error {
		callCount++
		if name == "phase3_work_objects_prototype_note" {
			return errors.New("phase3 failed")
		}
		return nil
	}

	store := openRawStore(t)
	if err := store.Migrate(); err == nil {
		t.Fatal("expected phase3 migration record error")
	}
	if callCount != 3 {
		t.Fatalf("expected 3 record migration calls, got %d", callCount)
	}
}

func TestMigrateReturnsPhase5RecordError(t *testing.T) {
	originalRecordMigration := recordMigration
	t.Cleanup(func() { recordMigration = originalRecordMigration })

	callCount := 0
	recordMigration = func(_ *sql.DB, name, _ string) error {
		callCount++
		if name == "phase5_work_objects_release_ready_at" {
			return errors.New("phase5 failed")
		}
		return nil
	}

	store := openRawStore(t)
	if err := store.Migrate(); err == nil {
		t.Fatal("expected phase5 migration record error")
	}
	if callCount != 4 {
		t.Fatalf("expected 4 record migration calls, got %d", callCount)
	}
}

func TestMigrateReturnsTrancheBRecordError(t *testing.T) {
	originalRecordMigration := recordMigration
	t.Cleanup(func() { recordMigration = originalRecordMigration })

	callCount := 0
	recordMigration = func(_ *sql.DB, name, _ string) error {
		callCount++
		if name == "tranche_b_users_onboarding_completed_at" {
			return errors.New("tranche b failed")
		}
		return nil
	}

	store := openRawStore(t)
	if err := store.Migrate(); err == nil {
		t.Fatal("expected tranche b migration record error")
	}
	if callCount != 5 {
		t.Fatalf("expected 5 record migration calls, got %d", callCount)
	}
}

func TestTableHasColumnReturnsScanError(t *testing.T) {
	store := openRawStore(t)
	originalQueryTableInfo := queryTableInfo
	t.Cleanup(func() { queryTableInfo = originalQueryTableInfo })

	queryTableInfo = func(db *sql.DB, _ string) (*sql.Rows, error) {
		return db.Query(`SELECT 1`)
	}

	if _, err := tableHasColumn(store.DB, "work_objects", "cluster_id"); err == nil {
		t.Fatal("expected scan error")
	}
}
