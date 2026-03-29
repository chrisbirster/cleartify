package store

import (
	"context"
	"testing"
)

func TestListClustersReturnsNestedSignals(t *testing.T) {
	t.Parallel()

	store := newTestStore(t, true)
	clusters, err := store.ListClusters(context.Background())
	if err != nil {
		t.Fatalf("expected cluster list: %v", err)
	}

	if len(clusters) != 3 {
		t.Fatalf("expected 3 clusters, got %d", len(clusters))
	}
	if len(clusters[0].Signals) == 0 {
		t.Fatalf("expected nested signals, got %#v", clusters[0])
	}
	if clusters[0].UpdatedAt.IsZero() {
		t.Fatalf("expected updated_at to be decoded, got %#v", clusters[0])
	}
}

func TestListClustersReturnsErrorsForBadData(t *testing.T) {
	t.Run("closed database", func(t *testing.T) {
		store := openRawStore(t)
		if err := store.Close(); err != nil {
			t.Fatalf("expected close: %v", err)
		}
		if _, err := store.ListClusters(context.Background()); err == nil {
			t.Fatal("expected query error")
		}
	})

	t.Run("scan error", func(t *testing.T) {
		store := openRawStore(t)
		mustExec(t, store.DB, `CREATE TABLE clusters (id TEXT, title TEXT, summary TEXT, severity TEXT, status TEXT, created_at TEXT, updated_at TEXT);`)
		mustExec(t, store.DB, `INSERT INTO clusters (id, title, summary, severity, status, created_at, updated_at) VALUES (NULL, 'Cluster', 'Summary', 'high', 'active', '2026-03-01T09:00:00Z', '2026-03-01T09:30:00Z');`)
		if _, err := store.ListClusters(context.Background()); err == nil {
			t.Fatal("expected scan error")
		}
	})

	t.Run("invalid cluster created_at", func(t *testing.T) {
		store := openRawStore(t)
		mustExec(t, store.DB, `CREATE TABLE clusters (id TEXT, title TEXT, summary TEXT, severity TEXT, status TEXT, created_at TEXT, updated_at TEXT);`)
		mustExec(t, store.DB, `INSERT INTO clusters (id, title, summary, severity, status, created_at, updated_at) VALUES ('clu_1', 'Cluster', 'Summary', 'high', 'active', 'bad', '2026-03-01T09:30:00Z');`)
		if _, err := store.ListClusters(context.Background()); err == nil {
			t.Fatal("expected created_at error")
		}
	})

	t.Run("invalid cluster updated_at", func(t *testing.T) {
		store := openRawStore(t)
		mustExec(t, store.DB, `CREATE TABLE clusters (id TEXT, title TEXT, summary TEXT, severity TEXT, status TEXT, created_at TEXT, updated_at TEXT);`)
		mustExec(t, store.DB, `INSERT INTO clusters (id, title, summary, severity, status, created_at, updated_at) VALUES ('clu_1', 'Cluster', 'Summary', 'high', 'active', '2026-03-01T09:00:00Z', 'bad');`)
		if _, err := store.ListClusters(context.Background()); err == nil {
			t.Fatal("expected updated_at error")
		}
	})

	t.Run("nested signals query error", func(t *testing.T) {
		store := openRawStore(t)
		mustExec(t, store.DB, `CREATE TABLE clusters (id TEXT, title TEXT, summary TEXT, severity TEXT, status TEXT, created_at TEXT, updated_at TEXT);`)
		mustExec(t, store.DB, `INSERT INTO clusters (id, title, summary, severity, status, created_at, updated_at) VALUES ('clu_1', 'Cluster', 'Summary', 'high', 'active', '2026-03-01T09:00:00Z', '2026-03-01T09:30:00Z');`)
		if _, err := store.ListClusters(context.Background()); err == nil {
			t.Fatal("expected nested signal query error")
		}
	})

	t.Run("nested signal scan error", func(t *testing.T) {
		store := openRawStore(t)
		mustExec(t, store.DB, `CREATE TABLE signals (id TEXT, source_type TEXT, source_id TEXT, title TEXT, body TEXT, url TEXT, metadata_json TEXT, created_at TEXT);`)
		mustExec(t, store.DB, `CREATE TABLE cluster_signals (cluster_id TEXT, signal_id TEXT);`)
		mustExec(t, store.DB, `INSERT INTO signals (id, source_type, source_id, title, body, created_at) VALUES ('sig_1', NULL, 's1', 'Signal', 'Body', '2026-03-01T09:00:00Z');`)
		mustExec(t, store.DB, `INSERT INTO cluster_signals (cluster_id, signal_id) VALUES ('clu_1', 'sig_1');`)
		if _, err := store.listSignalsForCluster(context.Background(), "clu_1"); err == nil {
			t.Fatal("expected nested signal scan error")
		}
	})

	t.Run("nested signal metadata error", func(t *testing.T) {
		store := openRawStore(t)
		mustExec(t, store.DB, `CREATE TABLE signals (id TEXT, source_type TEXT, source_id TEXT, title TEXT, body TEXT, url TEXT, metadata_json TEXT, created_at TEXT);`)
		mustExec(t, store.DB, `CREATE TABLE cluster_signals (cluster_id TEXT, signal_id TEXT);`)
		mustExec(t, store.DB, `INSERT INTO signals (id, source_type, source_id, title, body, metadata_json, created_at) VALUES ('sig_1', 'feedback', 's1', 'Signal', 'Body', '{', '2026-03-01T09:00:00Z');`)
		mustExec(t, store.DB, `INSERT INTO cluster_signals (cluster_id, signal_id) VALUES ('clu_1', 'sig_1');`)
		if _, err := store.listSignalsForCluster(context.Background(), "clu_1"); err == nil {
			t.Fatal("expected nested signal metadata error")
		}
	})

	t.Run("nested signal timestamp error", func(t *testing.T) {
		store := openRawStore(t)
		mustExec(t, store.DB, `CREATE TABLE signals (id TEXT, source_type TEXT, source_id TEXT, title TEXT, body TEXT, url TEXT, metadata_json TEXT, created_at TEXT);`)
		mustExec(t, store.DB, `CREATE TABLE cluster_signals (cluster_id TEXT, signal_id TEXT);`)
		mustExec(t, store.DB, `INSERT INTO signals (id, source_type, source_id, title, body, created_at) VALUES ('sig_1', 'feedback', 's1', 'Signal', 'Body', 'bad');`)
		mustExec(t, store.DB, `INSERT INTO cluster_signals (cluster_id, signal_id) VALUES ('clu_1', 'sig_1');`)
		if _, err := store.listSignalsForCluster(context.Background(), "clu_1"); err == nil {
			t.Fatal("expected nested signal timestamp error")
		}
	})
}
