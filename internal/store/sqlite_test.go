package store

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

func newTestStore(t *testing.T, seedDemo bool) *Store {
	t.Helper()

	store, err := Bootstrap(filepath.Join(t.TempDir(), "cleartify.db"), seedDemo)
	if err != nil {
		t.Fatalf("expected test store bootstrap: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return store
}

func openRawStore(t *testing.T) *Store {
	t.Helper()

	store, err := Open(filepath.Join(t.TempDir(), "raw.db"))
	if err != nil {
		t.Fatalf("expected raw sqlite store: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return store
}

func mustExec(t *testing.T, db *sql.DB, statement string, args ...any) {
	t.Helper()

	if _, err := db.Exec(statement, args...); err != nil {
		t.Fatalf("expected exec to succeed for %q: %v", statement, err)
	}
}

func TestOpenAppliesSQLitePragmas(t *testing.T) {
	t.Parallel()

	store, err := Open(filepath.Join(t.TempDir(), "cleartify.db"))
	if err != nil {
		t.Fatalf("expected sqlite open: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	var foreignKeys int
	if err := store.DB.QueryRow(`PRAGMA foreign_keys;`).Scan(&foreignKeys); err != nil {
		t.Fatalf("expected foreign keys pragma: %v", err)
	}
	if foreignKeys != 1 {
		t.Fatalf("expected foreign keys enabled, got %d", foreignKeys)
	}

	var busyTimeout int
	if err := store.DB.QueryRow(`PRAGMA busy_timeout;`).Scan(&busyTimeout); err != nil {
		t.Fatalf("expected busy timeout pragma: %v", err)
	}
	if busyTimeout != 5000 {
		t.Fatalf("expected busy timeout 5000, got %d", busyTimeout)
	}
}

func TestBootstrapAndClose(t *testing.T) {
	t.Parallel()

	store := newTestStore(t, false)
	if store.DB == nil {
		t.Fatal("expected db to be initialized")
	}

	count, err := countRows(context.Background(), store.DB, "signals")
	if err != nil {
		t.Fatalf("expected signal count query: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected empty signal table before seeding, got %d", count)
	}
}

func TestBootstrapReturnsOpenError(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(filePath, []byte("x"), 0o644); err != nil {
		t.Fatalf("expected file fixture: %v", err)
	}

	if _, err := Bootstrap(filepath.Join(filePath, "db.sqlite"), false); err == nil {
		t.Fatal("expected bootstrap open error")
	}
}

func TestBootstrapReturnsSeedError(t *testing.T) {
	originalStatements := phase0MigrationStatements
	t.Cleanup(func() { phase0MigrationStatements = originalStatements })

	phase0MigrationStatements = []string{
		`CREATE TABLE IF NOT EXISTS migrations (
			name TEXT PRIMARY KEY,
			applied_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS work_objects (
			id TEXT PRIMARY KEY,
			cluster_id TEXT,
			title TEXT NOT NULL,
			summary TEXT NOT NULL,
			goal TEXT NOT NULL,
			status TEXT NOT NULL,
			priority TEXT NOT NULL,
			owner_type TEXT NOT NULL,
			owner_id TEXT NOT NULL,
			prototype_url TEXT,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);`,
	}

	if _, err := Bootstrap(filepath.Join(t.TempDir(), "seed-error.db"), true); err == nil {
		t.Fatal("expected bootstrap seed error")
	}
}

func TestOpenReturnsErrorsForInvalidInputs(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "not-a-directory")
	mustWriteFile := func(path string) {
		t.Helper()
		if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
			t.Fatalf("expected file fixture: %v", err)
		}
	}

	mustWriteFile(filePath)
	if _, err := Open(filepath.Join(filePath, "db.sqlite")); err == nil {
		t.Fatal("expected mkdir error")
	}

	originalDriver := sqliteDriverName
	t.Cleanup(func() { sqliteDriverName = originalDriver })
	sqliteDriverName = "missing-driver"
	if _, err := Open(filepath.Join(t.TempDir(), "driver.db")); err == nil {
		t.Fatal("expected sql.Open error")
	}
}

func TestOpenReturnsPragmaErrors(t *testing.T) {
	originalPragmas := sqlitePragmas
	t.Cleanup(func() { sqlitePragmas = originalPragmas })

	sqlitePragmas = []string{"SELECT * FROM definitely_missing_table;"}
	if _, err := Open(filepath.Join(t.TempDir(), "pragma.db")); err == nil {
		t.Fatal("expected pragma error")
	}
}

func TestStoreHelperFunctionsHandleEdgeCases(t *testing.T) {
	t.Parallel()

	if err := (*Store)(nil).Close(); err != nil {
		t.Fatalf("expected nil store close to succeed: %v", err)
	}

	store := openRawStore(t)
	if _, err := countRows(context.Background(), store.DB, "missing"); err == nil {
		t.Fatal("expected countRows to fail for a missing table")
	}

	value, err := decodeJSONMap(sql.NullString{})
	if err != nil {
		t.Fatalf("expected empty metadata to decode: %v", err)
	}
	if value != nil {
		t.Fatalf("expected nil metadata, got %#v", value)
	}

	if _, err := decodeJSONMap(sql.NullString{Valid: true, String: "{"}); err == nil {
		t.Fatal("expected invalid metadata json error")
	}
}
