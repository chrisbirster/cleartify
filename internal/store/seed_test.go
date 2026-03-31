package store

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
)

func TestSeedDemoAddsDeterministicRecords(t *testing.T) {
	t.Parallel()

	store := newTestStore(t, true)
	ctx := context.Background()

	assertCount := func(table string, expected int) {
		t.Helper()

		count, err := countRows(ctx, store.DB, table)
		if err != nil {
			t.Fatalf("expected count for table %s: %v", table, err)
		}
		if count != expected {
			t.Fatalf("expected %d rows in %s, got %d", expected, table, count)
		}
	}

	assertCount("users", 1)
	assertCount("sessions", 1)
	assertCount("signals", 6)
	assertCount("clusters", 3)
	assertCount("cluster_signals", 6)
	assertCount("work_objects", 2)
	assertCount("draft_plans", 0)
	assertCount("decisions", 2)
	assertCount("artifacts", 2)

	rows, err := store.DB.Query(`SELECT id, COALESCE(cluster_id, '') FROM work_objects ORDER BY id ASC`)
	if err != nil {
		t.Fatalf("expected work object cluster link query: %v", err)
	}
	defer rows.Close()

	clusterLinks := map[string]string{}
	for rows.Next() {
		var id string
		var clusterID string
		if err := rows.Scan(&id, &clusterID); err != nil {
			t.Fatalf("expected work object cluster link scan: %v", err)
		}
		clusterLinks[id] = clusterID
	}

	if clusterLinks["wo_mobile_perf"] != "clu_mobile_perf" {
		t.Fatalf("expected wo_mobile_perf cluster link, got %#v", clusterLinks)
	}
	if clusterLinks["wo_onboarding_refresh"] != "clu_onboarding_refresh" {
		t.Fatalf("expected wo_onboarding_refresh cluster link, got %#v", clusterLinks)
	}

	var prototypeURL string
	var prototypeNote string
	if err := store.DB.QueryRow(`SELECT prototype_url, COALESCE(prototype_note, '') FROM work_objects WHERE id = ?`, "wo_onboarding_refresh").Scan(&prototypeURL, &prototypeNote); err != nil {
		t.Fatalf("expected work object prototype query: %v", err)
	}
	if prototypeURL == "" {
		t.Fatal("expected seeded prototype url")
	}
	if prototypeNote == "" {
		t.Fatal("expected seeded prototype note")
	}
}

func TestSeedDemoDoesNotReseedNonEmptyDatabase(t *testing.T) {
	t.Parallel()

	store := newTestStore(t, true)
	if err := store.SeedDemo(); err != nil {
		t.Fatalf("expected repeat seed to be ignored: %v", err)
	}

	count, err := countRows(context.Background(), store.DB, "signals")
	if err != nil {
		t.Fatalf("expected signal count: %v", err)
	}
	if count != 6 {
		t.Fatalf("expected 6 signals after reseed attempt, got %d", count)
	}
}

func TestBootstrapReturnsMigrateAndSeedErrors(t *testing.T) {
	originalStatements := phase0MigrationStatements
	t.Cleanup(func() { phase0MigrationStatements = originalStatements })

	phase0MigrationStatements = []string{"INVALID SQL"}
	if _, err := Bootstrap(filepath.Join(t.TempDir(), "migrate.db"), false); err == nil {
		t.Fatal("expected bootstrap migrate error")
	}

	phase0MigrationStatements = []string{
		`CREATE TABLE IF NOT EXISTS migrations (name TEXT PRIMARY KEY, applied_at TEXT NOT NULL);`,
	}
	if _, err := Bootstrap(filepath.Join(t.TempDir(), "seed.db"), true); err == nil {
		t.Fatal("expected bootstrap seed error")
	}
}

func TestSeedDemoReturnsCountAndBeginErrors(t *testing.T) {
	rawStore := openRawStore(t)
	if err := rawStore.SeedDemo(); err == nil {
		t.Fatal("expected countRows error when signals table is missing")
	}

	store := newTestStore(t, false)
	originalBeginTx := beginTx
	t.Cleanup(func() { beginTx = originalBeginTx })
	beginTx = func(*sql.DB, context.Context) (*sql.Tx, error) {
		return nil, errors.New("begin failed")
	}

	if err := store.SeedDemo(); err == nil {
		t.Fatal("expected begin tx error")
	}
}

func TestSeedDemoReturnsHelperErrors(t *testing.T) {
	t.Run("seedUser failure", func(t *testing.T) {
		store := openRawStore(t)
		mustExec(t, store.DB, `CREATE TABLE signals (id TEXT, source_type TEXT, source_id TEXT, title TEXT, body TEXT, url TEXT, metadata_json TEXT, created_at TEXT);`)
		if err := store.SeedDemo(); err == nil {
			t.Fatal("expected seed user error")
		}
	})

	t.Run("seedSignals failure", func(t *testing.T) {
		store := openRawStore(t)
		mustExec(t, store.DB, `CREATE TABLE signals (id TEXT, source_type TEXT, source_id TEXT, title TEXT, body TEXT, url TEXT, metadata_json TEXT, created_at TEXT);`)
		mustExec(t, store.DB, `CREATE TABLE users (id TEXT PRIMARY KEY, email TEXT, display_name TEXT, created_at TEXT);`)
		mustExec(t, store.DB, `CREATE TABLE sessions (id TEXT PRIMARY KEY, user_id TEXT, created_at TEXT);`)
		originalMarshal := jsonMarshal
		t.Cleanup(func() { jsonMarshal = originalMarshal })
		jsonMarshal = func(any) ([]byte, error) {
			return nil, errors.New("marshal failed")
		}

		if err := store.SeedDemo(); err == nil {
			t.Fatal("expected seed signals error")
		}
	})

	t.Run("seedClusters failure", func(t *testing.T) {
		store := openRawStore(t)
		mustExec(t, store.DB, `CREATE TABLE signals (id TEXT, source_type TEXT, source_id TEXT, title TEXT, body TEXT, url TEXT, metadata_json TEXT, created_at TEXT);`)
		mustExec(t, store.DB, `CREATE TABLE users (id TEXT PRIMARY KEY, email TEXT, display_name TEXT, created_at TEXT);`)
		mustExec(t, store.DB, `CREATE TABLE sessions (id TEXT PRIMARY KEY, user_id TEXT, created_at TEXT);`)
		mustExec(t, store.DB, `CREATE TABLE clusters (id TEXT PRIMARY KEY, title TEXT, summary TEXT, severity TEXT, status TEXT, created_at TEXT, updated_at TEXT);`)
		if err := store.SeedDemo(); err == nil {
			t.Fatal("expected seed clusters error")
		}
	})

	t.Run("seedWorkObjects failure", func(t *testing.T) {
		store := openRawStore(t)
		mustExec(t, store.DB, `CREATE TABLE signals (id TEXT, source_type TEXT, source_id TEXT, title TEXT, body TEXT, url TEXT, metadata_json TEXT, created_at TEXT);`)
		mustExec(t, store.DB, `CREATE TABLE users (id TEXT PRIMARY KEY, email TEXT, display_name TEXT, created_at TEXT);`)
		mustExec(t, store.DB, `CREATE TABLE sessions (id TEXT PRIMARY KEY, user_id TEXT, created_at TEXT);`)
		mustExec(t, store.DB, `CREATE TABLE clusters (id TEXT PRIMARY KEY, title TEXT, summary TEXT, severity TEXT, status TEXT, created_at TEXT, updated_at TEXT);`)
		mustExec(t, store.DB, `CREATE TABLE cluster_signals (cluster_id TEXT, signal_id TEXT);`)
		mustExec(t, store.DB, `CREATE TABLE work_objects (id TEXT PRIMARY KEY, cluster_id TEXT, title TEXT, summary TEXT, goal TEXT, status TEXT, priority TEXT, owner_type TEXT, owner_id TEXT, prototype_url TEXT, prototype_note TEXT, release_ready_at TEXT, created_at TEXT, updated_at TEXT);`)
		mustExec(t, store.DB, `CREATE TABLE decisions (id TEXT PRIMARY KEY, work_object_id TEXT, title TEXT, summary TEXT, rationale TEXT, created_at TEXT);`)
		if err := store.SeedDemo(); err == nil {
			t.Fatal("expected seed work objects error")
		}
	})
}

func TestSeedHelpersReturnErrors(t *testing.T) {
	t.Run("seedUser fails without users table", func(t *testing.T) {
		store := openRawStore(t)
		tx, err := store.DB.Begin()
		if err != nil {
			t.Fatalf("expected tx: %v", err)
		}
		defer tx.Rollback()

		if err := seedUser(tx); err == nil {
			t.Fatal("expected user insert error")
		}
	})

	t.Run("seedUser fails without sessions table", func(t *testing.T) {
		store := openRawStore(t)
		mustExec(t, store.DB, `CREATE TABLE users (id TEXT PRIMARY KEY, email TEXT, display_name TEXT, created_at TEXT);`)
		tx, err := store.DB.Begin()
		if err != nil {
			t.Fatalf("expected tx: %v", err)
		}
		defer tx.Rollback()

		if err := seedUser(tx); err == nil {
			t.Fatal("expected session insert error")
		}
	})

	t.Run("seedSignals fails when marshal fails", func(t *testing.T) {
		store := openRawStore(t)
		tx, err := store.DB.Begin()
		if err != nil {
			t.Fatalf("expected tx: %v", err)
		}
		defer tx.Rollback()

		originalMarshal := jsonMarshal
		t.Cleanup(func() { jsonMarshal = originalMarshal })
		jsonMarshal = func(any) ([]byte, error) {
			return nil, errors.New("marshal failed")
		}

		if err := seedSignals(tx); err == nil {
			t.Fatal("expected marshal error")
		}
	})

	t.Run("seedSignals fails without signals table", func(t *testing.T) {
		store := openRawStore(t)
		tx, err := store.DB.Begin()
		if err != nil {
			t.Fatalf("expected tx: %v", err)
		}
		defer tx.Rollback()

		if err := seedSignals(tx); err == nil {
			t.Fatal("expected signal insert error")
		}
	})

	t.Run("seedClusters fails without clusters table", func(t *testing.T) {
		store := openRawStore(t)
		tx, err := store.DB.Begin()
		if err != nil {
			t.Fatalf("expected tx: %v", err)
		}
		defer tx.Rollback()

		if err := seedClusters(tx); err == nil {
			t.Fatal("expected cluster insert error")
		}
	})

	t.Run("seedClusters fails without cluster_signals table", func(t *testing.T) {
		store := openRawStore(t)
		mustExec(t, store.DB, `CREATE TABLE clusters (id TEXT PRIMARY KEY, title TEXT, summary TEXT, severity TEXT, status TEXT, created_at TEXT, updated_at TEXT);`)
		tx, err := store.DB.Begin()
		if err != nil {
			t.Fatalf("expected tx: %v", err)
		}
		defer tx.Rollback()

		if err := seedClusters(tx); err == nil {
			t.Fatal("expected cluster signal insert error")
		}
	})

	t.Run("seedWorkObjects fails without work_objects table", func(t *testing.T) {
		store := openRawStore(t)
		tx, err := store.DB.Begin()
		if err != nil {
			t.Fatalf("expected tx: %v", err)
		}
		defer tx.Rollback()

		if err := seedWorkObjects(tx); err == nil {
			t.Fatal("expected work object insert error")
		}
	})

	t.Run("seedWorkObjects fails without decisions table", func(t *testing.T) {
		store := openRawStore(t)
		mustExec(t, store.DB, `CREATE TABLE work_objects (id TEXT PRIMARY KEY, cluster_id TEXT, title TEXT, summary TEXT, goal TEXT, status TEXT, priority TEXT, owner_type TEXT, owner_id TEXT, prototype_url TEXT, prototype_note TEXT, release_ready_at TEXT, created_at TEXT, updated_at TEXT);`)
		tx, err := store.DB.Begin()
		if err != nil {
			t.Fatalf("expected tx: %v", err)
		}
		defer tx.Rollback()

		if err := seedWorkObjects(tx); err == nil {
			t.Fatal("expected decision insert error")
		}
	})

	t.Run("seedWorkObjects fails when artifact metadata marshal fails", func(t *testing.T) {
		store := openRawStore(t)
		mustExec(t, store.DB, `CREATE TABLE work_objects (id TEXT PRIMARY KEY, cluster_id TEXT, title TEXT, summary TEXT, goal TEXT, status TEXT, priority TEXT, owner_type TEXT, owner_id TEXT, prototype_url TEXT, prototype_note TEXT, release_ready_at TEXT, created_at TEXT, updated_at TEXT);`)
		mustExec(t, store.DB, `CREATE TABLE decisions (id TEXT PRIMARY KEY, work_object_id TEXT, title TEXT, summary TEXT, rationale TEXT, created_at TEXT);`)
		tx, err := store.DB.Begin()
		if err != nil {
			t.Fatalf("expected tx: %v", err)
		}
		defer tx.Rollback()

		originalMarshal := jsonMarshal
		t.Cleanup(func() { jsonMarshal = originalMarshal })
		jsonMarshal = func(any) ([]byte, error) {
			return nil, errors.New("marshal failed")
		}

		if err := seedWorkObjects(tx); err == nil {
			t.Fatal("expected artifact marshal error")
		}
	})

	t.Run("seedWorkObjects fails without artifacts table", func(t *testing.T) {
		store := openRawStore(t)
		mustExec(t, store.DB, `CREATE TABLE work_objects (id TEXT PRIMARY KEY, cluster_id TEXT, title TEXT, summary TEXT, goal TEXT, status TEXT, priority TEXT, owner_type TEXT, owner_id TEXT, prototype_url TEXT, prototype_note TEXT, release_ready_at TEXT, created_at TEXT, updated_at TEXT);`)
		mustExec(t, store.DB, `CREATE TABLE decisions (id TEXT PRIMARY KEY, work_object_id TEXT, title TEXT, summary TEXT, rationale TEXT, created_at TEXT);`)
		tx, err := store.DB.Begin()
		if err != nil {
			t.Fatalf("expected tx: %v", err)
		}
		defer tx.Rollback()

		if err := seedWorkObjects(tx); err == nil {
			t.Fatal("expected artifact insert error")
		}
	})
}
