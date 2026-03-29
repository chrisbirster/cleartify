package store

import (
	"context"
	"testing"
)

func TestListSignalsReturnsSeededSignals(t *testing.T) {
	t.Parallel()

	store := newTestStore(t, true)
	signals, err := store.ListSignals(context.Background())
	if err != nil {
		t.Fatalf("expected signals list: %v", err)
	}

	if len(signals) != 6 {
		t.Fatalf("expected 6 signals, got %d", len(signals))
	}
	if signals[0].ID == "" || signals[0].CreatedAt.IsZero() {
		t.Fatalf("expected hydrated signal fields, got %#v", signals[0])
	}
}

func TestListSignalsReturnsErrorsForBadData(t *testing.T) {
	t.Run("closed database", func(t *testing.T) {
		store := openRawStore(t)
		if err := store.Close(); err != nil {
			t.Fatalf("expected close: %v", err)
		}
		if _, err := store.ListSignals(context.Background()); err == nil {
			t.Fatal("expected query error")
		}
	})

	t.Run("scan error", func(t *testing.T) {
		store := openRawStore(t)
		mustExec(t, store.DB, `CREATE TABLE signals (id TEXT, source_type TEXT, source_id TEXT, title TEXT, body TEXT, url TEXT, metadata_json TEXT, created_at TEXT);`)
		mustExec(t, store.DB, `INSERT INTO signals (id, source_type, source_id, title, body, created_at) VALUES (NULL, 'feedback', 's1', 'Signal', 'Body', '2026-03-01T09:00:00Z');`)
		if _, err := store.ListSignals(context.Background()); err == nil {
			t.Fatal("expected scan error")
		}
	})

	t.Run("invalid metadata", func(t *testing.T) {
		store := openRawStore(t)
		mustExec(t, store.DB, `CREATE TABLE signals (id TEXT, source_type TEXT, source_id TEXT, title TEXT, body TEXT, url TEXT, metadata_json TEXT, created_at TEXT);`)
		mustExec(t, store.DB, `INSERT INTO signals (id, source_type, source_id, title, body, metadata_json, created_at) VALUES ('sig_1', 'feedback', 's1', 'Signal', 'Body', '{', '2026-03-01T09:00:00Z');`)
		if _, err := store.ListSignals(context.Background()); err == nil {
			t.Fatal("expected metadata error")
		}
	})

	t.Run("invalid timestamp", func(t *testing.T) {
		store := openRawStore(t)
		mustExec(t, store.DB, `CREATE TABLE signals (id TEXT, source_type TEXT, source_id TEXT, title TEXT, body TEXT, url TEXT, metadata_json TEXT, created_at TEXT);`)
		mustExec(t, store.DB, `INSERT INTO signals (id, source_type, source_id, title, body, created_at) VALUES ('sig_1', 'feedback', 's1', 'Signal', 'Body', 'not-a-time');`)
		if _, err := store.ListSignals(context.Background()); err == nil {
			t.Fatal("expected timestamp error")
		}
	})
}
