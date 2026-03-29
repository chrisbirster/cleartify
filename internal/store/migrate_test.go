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
		`CREATE TABLE IF NOT EXISTS work_objects (id TEXT PRIMARY KEY, title TEXT, summary TEXT, goal TEXT, status TEXT, priority TEXT, owner_type TEXT, owner_id TEXT, prototype_url TEXT, created_at TEXT, updated_at TEXT);`,
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
