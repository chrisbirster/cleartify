package store

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"
)

func TestListWorkObjectsReturnsDecisionsAndArtifacts(t *testing.T) {
	t.Parallel()

	store := newTestStore(t, true)
	workObjects, err := store.ListWorkObjects(context.Background())
	if err != nil {
		t.Fatalf("expected work object list: %v", err)
	}

	if len(workObjects) != 2 {
		t.Fatalf("expected 2 work objects, got %d", len(workObjects))
	}
	if len(workObjects[0].Decisions) == 0 {
		t.Fatalf("expected nested decisions, got %#v", workObjects[0])
	}
	if len(workObjects[0].Artifacts) == 0 {
		t.Fatalf("expected nested artifacts, got %#v", workObjects[0])
	}
	if workObjects[0].ClusterID == "" {
		t.Fatalf("expected cluster id to be populated, got %#v", workObjects[0])
	}
}

func TestListWorkObjectsReturnsErrorsForBadData(t *testing.T) {
	t.Run("closed database", func(t *testing.T) {
		store := openRawStore(t)
		if err := store.Close(); err != nil {
			t.Fatalf("expected close: %v", err)
		}
		if _, err := store.ListWorkObjects(context.Background()); err == nil {
			t.Fatal("expected query error")
		}
	})

	t.Run("scan error", func(t *testing.T) {
		store := openRawStore(t)
		mustExec(t, store.DB, `CREATE TABLE work_objects (id TEXT, cluster_id TEXT, title TEXT, summary TEXT, goal TEXT, status TEXT, priority TEXT, owner_type TEXT, owner_id TEXT, prototype_url TEXT, created_at TEXT, updated_at TEXT);`)
		mustExec(t, store.DB, `INSERT INTO work_objects (id, cluster_id, title, summary, goal, status, priority, owner_type, owner_id, created_at, updated_at) VALUES (NULL, 'clu_1', 'Work', 'Summary', 'Goal', 'planning', 'high', 'user', 'usr_1', '2026-03-01T09:00:00Z', '2026-03-01T09:30:00Z');`)
		if _, err := store.ListWorkObjects(context.Background()); err == nil {
			t.Fatal("expected scan error")
		}
	})

	t.Run("invalid work object created_at", func(t *testing.T) {
		store := openRawStore(t)
		mustExec(t, store.DB, `CREATE TABLE work_objects (id TEXT, cluster_id TEXT, title TEXT, summary TEXT, goal TEXT, status TEXT, priority TEXT, owner_type TEXT, owner_id TEXT, prototype_url TEXT, created_at TEXT, updated_at TEXT);`)
		mustExec(t, store.DB, `INSERT INTO work_objects (id, cluster_id, title, summary, goal, status, priority, owner_type, owner_id, created_at, updated_at) VALUES ('wo_1', 'clu_1', 'Work', 'Summary', 'Goal', 'planning', 'high', 'user', 'usr_1', 'bad', '2026-03-01T09:30:00Z');`)
		if _, err := store.ListWorkObjects(context.Background()); err == nil {
			t.Fatal("expected created_at error")
		}
	})

	t.Run("invalid work object updated_at", func(t *testing.T) {
		store := openRawStore(t)
		mustExec(t, store.DB, `CREATE TABLE work_objects (id TEXT, cluster_id TEXT, title TEXT, summary TEXT, goal TEXT, status TEXT, priority TEXT, owner_type TEXT, owner_id TEXT, prototype_url TEXT, created_at TEXT, updated_at TEXT);`)
		mustExec(t, store.DB, `INSERT INTO work_objects (id, cluster_id, title, summary, goal, status, priority, owner_type, owner_id, created_at, updated_at) VALUES ('wo_1', 'clu_1', 'Work', 'Summary', 'Goal', 'planning', 'high', 'user', 'usr_1', '2026-03-01T09:00:00Z', 'bad');`)
		if _, err := store.ListWorkObjects(context.Background()); err == nil {
			t.Fatal("expected updated_at error")
		}
	})

	t.Run("decisions query error", func(t *testing.T) {
		store := openRawStore(t)
		mustExec(t, store.DB, `CREATE TABLE work_objects (id TEXT, cluster_id TEXT, title TEXT, summary TEXT, goal TEXT, status TEXT, priority TEXT, owner_type TEXT, owner_id TEXT, prototype_url TEXT, created_at TEXT, updated_at TEXT);`)
		mustExec(t, store.DB, `INSERT INTO work_objects (id, cluster_id, title, summary, goal, status, priority, owner_type, owner_id, created_at, updated_at) VALUES ('wo_1', 'clu_1', 'Work', 'Summary', 'Goal', 'planning', 'high', 'user', 'usr_1', '2026-03-01T09:00:00Z', '2026-03-01T09:30:00Z');`)
		if _, err := store.ListWorkObjects(context.Background()); err == nil {
			t.Fatal("expected decisions query error")
		}
	})

	t.Run("decision scan error", func(t *testing.T) {
		store := openRawStore(t)
		mustExec(t, store.DB, `CREATE TABLE decisions (id TEXT, work_object_id TEXT, title TEXT, summary TEXT, rationale TEXT, created_at TEXT);`)
		mustExec(t, store.DB, `INSERT INTO decisions (id, work_object_id, title, summary, rationale, created_at) VALUES (NULL, 'wo_1', 'Decision', 'Summary', 'Rationale', '2026-03-01T09:00:00Z');`)
		if _, err := store.listDecisionsForWorkObject(context.Background(), "wo_1"); err == nil {
			t.Fatal("expected decision scan error")
		}
	})

	t.Run("decision timestamp error", func(t *testing.T) {
		store := openRawStore(t)
		mustExec(t, store.DB, `CREATE TABLE decisions (id TEXT, work_object_id TEXT, title TEXT, summary TEXT, rationale TEXT, created_at TEXT);`)
		mustExec(t, store.DB, `INSERT INTO decisions (id, work_object_id, title, summary, rationale, created_at) VALUES ('dec_1', 'wo_1', 'Decision', 'Summary', 'Rationale', 'bad');`)
		if _, err := store.listDecisionsForWorkObject(context.Background(), "wo_1"); err == nil {
			t.Fatal("expected decision timestamp error")
		}
	})

	t.Run("artifacts query error", func(t *testing.T) {
		store := openRawStore(t)
		mustExec(t, store.DB, `CREATE TABLE work_objects (id TEXT, cluster_id TEXT, title TEXT, summary TEXT, goal TEXT, status TEXT, priority TEXT, owner_type TEXT, owner_id TEXT, prototype_url TEXT, created_at TEXT, updated_at TEXT);`)
		mustExec(t, store.DB, `CREATE TABLE decisions (id TEXT, work_object_id TEXT, title TEXT, summary TEXT, rationale TEXT, created_at TEXT);`)
		mustExec(t, store.DB, `INSERT INTO work_objects (id, cluster_id, title, summary, goal, status, priority, owner_type, owner_id, created_at, updated_at) VALUES ('wo_1', 'clu_1', 'Work', 'Summary', 'Goal', 'planning', 'high', 'user', 'usr_1', '2026-03-01T09:00:00Z', '2026-03-01T09:30:00Z');`)
		if _, err := store.ListWorkObjects(context.Background()); err == nil {
			t.Fatal("expected artifacts query error")
		}
	})

	t.Run("artifact scan error", func(t *testing.T) {
		store := openRawStore(t)
		mustExec(t, store.DB, `CREATE TABLE artifacts (id TEXT, work_object_id TEXT, artifact_type TEXT, title TEXT, body TEXT, metadata_json TEXT, created_at TEXT);`)
		mustExec(t, store.DB, `INSERT INTO artifacts (id, work_object_id, artifact_type, title, body, created_at) VALUES (NULL, 'wo_1', 'plan', 'Artifact', 'Body', '2026-03-01T09:00:00Z');`)
		if _, err := store.listArtifactsForWorkObject(context.Background(), "wo_1"); err == nil {
			t.Fatal("expected artifact scan error")
		}
	})

	t.Run("artifact metadata error", func(t *testing.T) {
		store := openRawStore(t)
		mustExec(t, store.DB, `CREATE TABLE artifacts (id TEXT, work_object_id TEXT, artifact_type TEXT, title TEXT, body TEXT, metadata_json TEXT, created_at TEXT);`)
		mustExec(t, store.DB, `INSERT INTO artifacts (id, work_object_id, artifact_type, title, body, metadata_json, created_at) VALUES ('art_1', 'wo_1', 'plan', 'Artifact', 'Body', '{', '2026-03-01T09:00:00Z');`)
		if _, err := store.listArtifactsForWorkObject(context.Background(), "wo_1"); err == nil {
			t.Fatal("expected artifact metadata error")
		}
	})

	t.Run("artifact timestamp error", func(t *testing.T) {
		store := openRawStore(t)
		mustExec(t, store.DB, `CREATE TABLE artifacts (id TEXT, work_object_id TEXT, artifact_type TEXT, title TEXT, body TEXT, metadata_json TEXT, created_at TEXT);`)
		mustExec(t, store.DB, `INSERT INTO artifacts (id, work_object_id, artifact_type, title, body, created_at) VALUES ('art_1', 'wo_1', 'plan', 'Artifact', 'Body', 'bad');`)
		if _, err := store.listArtifactsForWorkObject(context.Background(), "wo_1"); err == nil {
			t.Fatal("expected artifact timestamp error")
		}
	})
}

func TestCreateWorkObjectFromClusterPersistsClusterLink(t *testing.T) {
	store := newTestStore(t, true)
	originalNow := workObjectNow
	originalID := workObjectIDFromCluster
	t.Cleanup(func() {
		workObjectNow = originalNow
		workObjectIDFromCluster = originalID
	})

	createdAt := time.Date(2026, 3, 29, 14, 30, 0, 0, time.UTC)
	workObjectNow = func() time.Time { return createdAt }
	workObjectIDFromCluster = func(clusterID string, _ time.Time) string {
		return "wo_new_cluster"
	}

	workObject, err := store.CreateWorkObjectFromCluster(context.Background(), "clu_product_memory")
	if err != nil {
		t.Fatalf("expected work object creation: %v", err)
	}

	if workObject.ID != "wo_new_cluster" {
		t.Fatalf("expected deterministic work object id, got %q", workObject.ID)
	}
	if workObject.ClusterID != "clu_product_memory" {
		t.Fatalf("expected cluster link, got %q", workObject.ClusterID)
	}
	if workObject.Status != "planning" {
		t.Fatalf("expected planning status, got %q", workObject.Status)
	}
	if workObject.Priority != "medium" {
		t.Fatalf("expected medium priority from medium severity, got %q", workObject.Priority)
	}

	workObjects, err := store.ListWorkObjects(context.Background())
	if err != nil {
		t.Fatalf("expected work object list after create: %v", err)
	}

	var found bool
	for _, entry := range workObjects {
		if entry.ID == "wo_new_cluster" {
			found = true
			if entry.ClusterID != "clu_product_memory" {
				t.Fatalf("expected persisted cluster link, got %q", entry.ClusterID)
			}
		}
	}
	if !found {
		t.Fatal("expected created work object to be persisted")
	}
}

func TestCreateWorkObjectFromClusterReturnsNotFound(t *testing.T) {
	t.Parallel()

	store := newTestStore(t, true)
	if _, err := store.CreateWorkObjectFromCluster(context.Background(), "missing"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestCreateWorkObjectFromClusterReturnsInsertErrors(t *testing.T) {
	t.Parallel()

	store := openRawStore(t)
	mustExec(t, store.DB, `CREATE TABLE clusters (id TEXT PRIMARY KEY, title TEXT, summary TEXT, severity TEXT, status TEXT, created_at TEXT, updated_at TEXT);`)
	mustExec(t, store.DB, `INSERT INTO clusters (id, title, summary, severity, status, created_at, updated_at) VALUES ('clu_1', 'Inbox', 'Summary', 'high', 'active', '2026-03-01T09:00:00Z', '2026-03-01T09:30:00Z');`)

	if _, err := store.CreateWorkObjectFromCluster(context.Background(), "clu_1"); err == nil {
		t.Fatal("expected insert error without work_objects table")
	}
}

func TestLookupClusterForWorkObjectReturnsLookupErrors(t *testing.T) {
	t.Run("query error", func(t *testing.T) {
		store := openRawStore(t)
		if _, err := store.lookupClusterForWorkObject(context.Background(), "clu_1"); err == nil {
			t.Fatal("expected lookup query error")
		}
	})

	t.Run("scan error", func(t *testing.T) {
		store := openRawStore(t)
		mustExec(t, store.DB, `CREATE TABLE clusters (id TEXT, title TEXT, summary TEXT, severity TEXT, status TEXT, created_at TEXT, updated_at TEXT);`)
		mustExec(t, store.DB, `INSERT INTO clusters (id, title, summary, severity, status, created_at, updated_at) VALUES ('clu_1', NULL, 'Summary', 'high', 'active', '2026-03-01T09:00:00Z', '2026-03-01T09:30:00Z');`)
		if _, err := store.lookupClusterForWorkObject(context.Background(), "clu_1"); err == nil {
			t.Fatal("expected lookup scan error")
		}
	})

	t.Run("invalid created_at", func(t *testing.T) {
		store := openRawStore(t)
		mustExec(t, store.DB, `CREATE TABLE clusters (id TEXT, title TEXT, summary TEXT, severity TEXT, status TEXT, created_at TEXT, updated_at TEXT);`)
		mustExec(t, store.DB, `INSERT INTO clusters (id, title, summary, severity, status, created_at, updated_at) VALUES ('clu_1', 'Cluster', 'Summary', 'high', 'active', 'bad', '2026-03-01T09:30:00Z');`)
		if _, err := store.lookupClusterForWorkObject(context.Background(), "clu_1"); err == nil {
			t.Fatal("expected cluster created_at decode error")
		}
	})

	t.Run("invalid updated_at", func(t *testing.T) {
		store := openRawStore(t)
		mustExec(t, store.DB, `CREATE TABLE clusters (id TEXT, title TEXT, summary TEXT, severity TEXT, status TEXT, created_at TEXT, updated_at TEXT);`)
		mustExec(t, store.DB, `INSERT INTO clusters (id, title, summary, severity, status, created_at, updated_at) VALUES ('clu_1', 'Cluster', 'Summary', 'high', 'active', '2026-03-01T09:00:00Z', 'bad');`)
		if _, err := store.lookupClusterForWorkObject(context.Background(), "clu_1"); err == nil {
			t.Fatal("expected cluster updated_at decode error")
		}
	})
}

func TestWorkObjectPriorityForSeverity(t *testing.T) {
	t.Parallel()

	if priority := workObjectPriorityForSeverity("high"); priority != "high" {
		t.Fatalf("expected high priority, got %q", priority)
	}
	if priority := workObjectPriorityForSeverity("low"); priority != "low" {
		t.Fatalf("expected low priority, got %q", priority)
	}
	if priority := workObjectPriorityForSeverity("anything-else"); priority != "medium" {
		t.Fatalf("expected default medium priority, got %q", priority)
	}
}
