package store

import (
	"context"
	"database/sql"
	"errors"
	"strings"
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
		mustExec(t, store.DB, `CREATE TABLE work_objects (id TEXT, cluster_id TEXT, title TEXT, summary TEXT, goal TEXT, status TEXT, priority TEXT, owner_type TEXT, owner_id TEXT, prototype_url TEXT, prototype_note TEXT, release_ready_at TEXT, created_at TEXT, updated_at TEXT);`)
		mustExec(t, store.DB, `INSERT INTO work_objects (id, cluster_id, title, summary, goal, status, priority, owner_type, owner_id, created_at, updated_at) VALUES (NULL, 'clu_1', 'Work', 'Summary', 'Goal', 'planning', 'high', 'user', 'usr_1', '2026-03-01T09:00:00Z', '2026-03-01T09:30:00Z');`)
		if _, err := store.ListWorkObjects(context.Background()); err == nil {
			t.Fatal("expected scan error")
		}
	})

	t.Run("invalid work object created_at", func(t *testing.T) {
		store := openRawStore(t)
		mustExec(t, store.DB, `CREATE TABLE work_objects (id TEXT, cluster_id TEXT, title TEXT, summary TEXT, goal TEXT, status TEXT, priority TEXT, owner_type TEXT, owner_id TEXT, prototype_url TEXT, prototype_note TEXT, release_ready_at TEXT, created_at TEXT, updated_at TEXT);`)
		mustExec(t, store.DB, `INSERT INTO work_objects (id, cluster_id, title, summary, goal, status, priority, owner_type, owner_id, created_at, updated_at) VALUES ('wo_1', 'clu_1', 'Work', 'Summary', 'Goal', 'planning', 'high', 'user', 'usr_1', 'bad', '2026-03-01T09:30:00Z');`)
		if _, err := store.ListWorkObjects(context.Background()); err == nil {
			t.Fatal("expected created_at error")
		}
	})

	t.Run("invalid work object updated_at", func(t *testing.T) {
		store := openRawStore(t)
		mustExec(t, store.DB, `CREATE TABLE work_objects (id TEXT, cluster_id TEXT, title TEXT, summary TEXT, goal TEXT, status TEXT, priority TEXT, owner_type TEXT, owner_id TEXT, prototype_url TEXT, prototype_note TEXT, release_ready_at TEXT, created_at TEXT, updated_at TEXT);`)
		mustExec(t, store.DB, `INSERT INTO work_objects (id, cluster_id, title, summary, goal, status, priority, owner_type, owner_id, created_at, updated_at) VALUES ('wo_1', 'clu_1', 'Work', 'Summary', 'Goal', 'planning', 'high', 'user', 'usr_1', '2026-03-01T09:00:00Z', 'bad');`)
		if _, err := store.ListWorkObjects(context.Background()); err == nil {
			t.Fatal("expected updated_at error")
		}
	})

	t.Run("decisions query error", func(t *testing.T) {
		store := openRawStore(t)
		mustExec(t, store.DB, `CREATE TABLE work_objects (id TEXT, cluster_id TEXT, title TEXT, summary TEXT, goal TEXT, status TEXT, priority TEXT, owner_type TEXT, owner_id TEXT, prototype_url TEXT, prototype_note TEXT, release_ready_at TEXT, created_at TEXT, updated_at TEXT);`)
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
		mustExec(t, store.DB, `CREATE TABLE work_objects (id TEXT, cluster_id TEXT, title TEXT, summary TEXT, goal TEXT, status TEXT, priority TEXT, owner_type TEXT, owner_id TEXT, prototype_url TEXT, prototype_note TEXT, release_ready_at TEXT, created_at TEXT, updated_at TEXT);`)
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

func TestGenerateImplementationArtifactForWorkObjectPersistsArtifact(t *testing.T) {
	store := newTestStore(t, true)
	originalArtifactNow := artifactNow
	originalArtifactID := artifactIDFromWorkObject
	t.Cleanup(func() {
		artifactNow = originalArtifactNow
		artifactIDFromWorkObject = originalArtifactID
	})

	if _, err := store.GenerateDraftPlanForWorkObject(context.Background(), "wo_onboarding_refresh"); err != nil {
		t.Fatalf("expected draft plan generation: %v", err)
	}

	createdAt := time.Date(2026, 3, 31, 13, 30, 0, 0, time.UTC)
	artifactNow = func() time.Time { return createdAt }
	artifactIDFromWorkObject = func(workObjectID string, _ time.Time) string {
		return "art_impl_brief"
	}

	artifact, err := store.GenerateImplementationArtifactForWorkObject(context.Background(), "wo_onboarding_refresh")
	if err != nil {
		t.Fatalf("expected implementation artifact generation: %v", err)
	}

	if artifact.ID != "art_impl_brief" {
		t.Fatalf("expected deterministic artifact id, got %q", artifact.ID)
	}
	if artifact.ArtifactType != "implementation_brief" {
		t.Fatalf("expected implementation_brief artifact type, got %q", artifact.ArtifactType)
	}
	if artifact.Title != "Guided onboarding refresh implementation brief" {
		t.Fatalf("expected derived artifact title, got %q", artifact.Title)
	}
	if !strings.Contains(artifact.Body, "Overview") ||
		!strings.Contains(artifact.Body, "Open questions") ||
		!strings.Contains(artifact.Body, "Onboarding activation clarity (medium severity)") ||
		!strings.Contains(artifact.Body, "https://example.com/prototype/onboarding-refresh") ||
		!strings.Contains(artifact.Body, "Which prototype screens or states must ship in the first implementation pass?") {
		t.Fatalf("expected derived artifact body to include persisted context, got %q", artifact.Body)
	}

	workObjects, err := store.ListWorkObjects(context.Background())
	if err != nil {
		t.Fatalf("expected work object list after artifact generation: %v", err)
	}

	for _, entry := range workObjects {
		if entry.ID == "wo_onboarding_refresh" {
			var found bool
			for _, persisted := range entry.Artifacts {
				if persisted.ID == "art_impl_brief" {
					found = true
				}
			}
			if !found {
				t.Fatal("expected generated implementation artifact to persist on the work object")
			}
		}
	}
}

func TestGenerateImplementationArtifactForWorkObjectReturnsNotFound(t *testing.T) {
	t.Parallel()

	store := newTestStore(t, true)
	if _, err := store.GenerateImplementationArtifactForWorkObject(context.Background(), "missing"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestGenerateImplementationArtifactForWorkObjectReturnsDraftPlanNotFound(t *testing.T) {
	t.Parallel()

	store := newTestStore(t, true)
	if _, err := store.GenerateImplementationArtifactForWorkObject(context.Background(), "wo_mobile_perf"); !errors.Is(err, ErrDraftPlanNotFound) {
		t.Fatalf("expected missing draft plan error, got %v", err)
	}
}

func TestGenerateTaskBreakdownArtifactForWorkObjectPersistsArtifact(t *testing.T) {
	store := newTestStore(t, true)
	originalArtifactNow := artifactNow
	originalArtifactID := artifactIDFromWorkObject
	t.Cleanup(func() {
		artifactNow = originalArtifactNow
		artifactIDFromWorkObject = originalArtifactID
	})

	if _, err := store.GenerateDraftPlanForWorkObject(context.Background(), "wo_onboarding_refresh"); err != nil {
		t.Fatalf("expected draft plan generation: %v", err)
	}

	createdAt := time.Date(2026, 3, 31, 14, 0, 0, 0, time.UTC)
	artifactNow = func() time.Time { return createdAt }
	artifactIDFromWorkObject = func(workObjectID string, _ time.Time) string {
		return "art_task_breakdown"
	}

	artifact, err := store.GenerateTaskBreakdownArtifactForWorkObject(context.Background(), "wo_onboarding_refresh")
	if err != nil {
		t.Fatalf("expected task-breakdown artifact generation: %v", err)
	}

	if artifact.ID != "art_task_breakdown" {
		t.Fatalf("expected deterministic artifact id, got %q", artifact.ID)
	}
	if artifact.ArtifactType != "task_breakdown" {
		t.Fatalf("expected task_breakdown artifact type, got %q", artifact.ArtifactType)
	}
	if artifact.Title != "Guided onboarding refresh task breakdown" {
		t.Fatalf("expected derived artifact title, got %q", artifact.Title)
	}
	if !strings.Contains(artifact.Body, "Execution tasks") ||
		!strings.Contains(artifact.Body, "1. Review the Work Object goal for") ||
		!strings.Contains(artifact.Body, "Linked Cluster: Onboarding activation clarity (medium severity)") ||
		!strings.Contains(artifact.Body, "Prototype link: https://example.com/prototype/onboarding-refresh") ||
		!strings.Contains(artifact.Body, "Open questions to resolve") {
		t.Fatalf("expected derived task breakdown body to include persisted context, got %q", artifact.Body)
	}

	workObjects, err := store.ListWorkObjects(context.Background())
	if err != nil {
		t.Fatalf("expected work object list after artifact generation: %v", err)
	}

	for _, entry := range workObjects {
		if entry.ID == "wo_onboarding_refresh" {
			var found bool
			for _, persisted := range entry.Artifacts {
				if persisted.ID == "art_task_breakdown" {
					found = true
				}
			}
			if !found {
				t.Fatal("expected generated task-breakdown artifact to persist on the work object")
			}
		}
	}
}

func TestGenerateTaskBreakdownArtifactForWorkObjectReturnsNotFound(t *testing.T) {
	t.Parallel()

	store := newTestStore(t, true)
	if _, err := store.GenerateTaskBreakdownArtifactForWorkObject(context.Background(), "missing"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestGenerateTaskBreakdownArtifactForWorkObjectReturnsDraftPlanNotFound(t *testing.T) {
	t.Parallel()

	store := newTestStore(t, true)
	if _, err := store.GenerateTaskBreakdownArtifactForWorkObject(context.Background(), "wo_mobile_perf"); !errors.Is(err, ErrDraftPlanNotFound) {
		t.Fatalf("expected missing draft plan error, got %v", err)
	}
}

func TestUpdateWorkObjectStatusPersistsChanges(t *testing.T) {
	store := newTestStore(t, true)
	originalNow := workObjectNow
	t.Cleanup(func() { workObjectNow = originalNow })

	updatedAt := time.Date(2026, 3, 29, 16, 15, 0, 0, time.UTC)
	workObjectNow = func() time.Time { return updatedAt }

	workObject, err := store.UpdateWorkObjectStatus(context.Background(), "wo_mobile_perf", "in_progress")
	if err != nil {
		t.Fatalf("expected work object status update: %v", err)
	}

	if workObject.Status != "in_progress" {
		t.Fatalf("expected updated status, got %q", workObject.Status)
	}
	if !workObject.UpdatedAt.Equal(updatedAt) {
		t.Fatalf("expected updated timestamp %s, got %s", updatedAt, workObject.UpdatedAt)
	}

	var persistedStatus string
	var persistedUpdatedAt string
	if err := store.DB.QueryRow(`SELECT status, updated_at FROM work_objects WHERE id = ?`, "wo_mobile_perf").Scan(&persistedStatus, &persistedUpdatedAt); err != nil {
		t.Fatalf("expected persisted work object status: %v", err)
	}
	if persistedStatus != "in_progress" {
		t.Fatalf("expected persisted in_progress status, got %q", persistedStatus)
	}
	if persistedUpdatedAt != updatedAt.Format(time.RFC3339) {
		t.Fatalf("expected persisted updated_at %q, got %q", updatedAt.Format(time.RFC3339), persistedUpdatedAt)
	}

	workObjects, err := store.ListWorkObjects(context.Background())
	if err != nil {
		t.Fatalf("expected work object list after update: %v", err)
	}

	for _, entry := range workObjects {
		if entry.ID == "wo_mobile_perf" && entry.Status != "in_progress" {
			t.Fatalf("expected list view to return updated status, got %q", entry.Status)
		}
	}
}

func TestUpdateWorkObjectStatusRejectsInvalidStatus(t *testing.T) {
	t.Parallel()

	store := newTestStore(t, true)
	if _, err := store.UpdateWorkObjectStatus(context.Background(), "wo_mobile_perf", "review"); !errors.Is(err, ErrInvalidWorkObjectStatus) {
		t.Fatalf("expected invalid status error, got %v", err)
	}
}

func TestUpdateWorkObjectStatusReturnsNotFound(t *testing.T) {
	t.Parallel()

	store := newTestStore(t, true)
	if _, err := store.UpdateWorkObjectStatus(context.Background(), "missing", "ready"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestUpdateWorkObjectCoreFieldsPersistsChanges(t *testing.T) {
	store := newTestStore(t, true)
	originalNow := workObjectNow
	t.Cleanup(func() { workObjectNow = originalNow })

	updatedAt := time.Date(2026, 3, 29, 17, 15, 0, 0, time.UTC)
	workObjectNow = func() time.Time { return updatedAt }

	workObject, err := store.UpdateWorkObjectCoreFields(
		context.Background(),
		"wo_mobile_perf",
		"Faster mobile dashboard rendering",
		"Keep mobile analytics responsive without flattening the product surface.",
		"Reduce perceived latency and keep the dashboard feel crisp on mobile.",
	)
	if err != nil {
		t.Fatalf("expected work object core field update: %v", err)
	}

	if workObject.Title != "Faster mobile dashboard rendering" {
		t.Fatalf("expected updated title, got %q", workObject.Title)
	}
	if workObject.Summary != "Keep mobile analytics responsive without flattening the product surface." {
		t.Fatalf("expected updated summary, got %q", workObject.Summary)
	}
	if workObject.Goal != "Reduce perceived latency and keep the dashboard feel crisp on mobile." {
		t.Fatalf("expected updated goal, got %q", workObject.Goal)
	}
	if !workObject.UpdatedAt.Equal(updatedAt) {
		t.Fatalf("expected updated timestamp %s, got %s", updatedAt, workObject.UpdatedAt)
	}

	var persistedTitle, persistedSummary, persistedGoal, persistedUpdatedAt string
	if err := store.DB.QueryRow(
		`SELECT title, summary, goal, updated_at FROM work_objects WHERE id = ?`,
		"wo_mobile_perf",
	).Scan(&persistedTitle, &persistedSummary, &persistedGoal, &persistedUpdatedAt); err != nil {
		t.Fatalf("expected persisted work object core fields: %v", err)
	}
	if persistedTitle != "Faster mobile dashboard rendering" || persistedSummary != "Keep mobile analytics responsive without flattening the product surface." || persistedGoal != "Reduce perceived latency and keep the dashboard feel crisp on mobile." {
		t.Fatalf("expected persisted core fields, got %q %q %q", persistedTitle, persistedSummary, persistedGoal)
	}
	if persistedUpdatedAt != updatedAt.Format(time.RFC3339) {
		t.Fatalf("expected persisted updated_at %q, got %q", updatedAt.Format(time.RFC3339), persistedUpdatedAt)
	}

	workObjects, err := store.ListWorkObjects(context.Background())
	if err != nil {
		t.Fatalf("expected work object list after update: %v", err)
	}
	for _, entry := range workObjects {
		if entry.ID == "wo_mobile_perf" {
			if entry.Title != "Faster mobile dashboard rendering" || entry.Summary != "Keep mobile analytics responsive without flattening the product surface." || entry.Goal != "Reduce perceived latency and keep the dashboard feel crisp on mobile." {
				t.Fatalf("expected list query to return updated core fields, got %#v", entry)
			}
		}
	}
}

func TestUpdateWorkObjectCoreFieldsRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	store := newTestStore(t, true)
	if _, err := store.UpdateWorkObjectCoreFields(context.Background(), "wo_mobile_perf", "   ", "Valid summary", "Valid goal"); !errors.Is(err, ErrInvalidWorkObjectCoreFields) {
		t.Fatalf("expected invalid core fields error, got %v", err)
	}
}

func TestUpdateWorkObjectCoreFieldsReturnsNotFound(t *testing.T) {
	t.Parallel()

	store := newTestStore(t, true)
	if _, err := store.UpdateWorkObjectCoreFields(context.Background(), "missing", "Valid title", "Valid summary", "Valid goal"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestUpdateWorkObjectPrototypePersistsChanges(t *testing.T) {
	store := newTestStore(t, true)
	originalNow := workObjectNow
	t.Cleanup(func() { workObjectNow = originalNow })

	updatedAt := time.Date(2026, 3, 29, 17, 20, 0, 0, time.UTC)
	workObjectNow = func() time.Time { return updatedAt }

	workObject, err := store.UpdateWorkObjectPrototype(
		context.Background(),
		"wo_mobile_perf",
		"https://example.com/prototype/mobile-perf-v2",
		"Updated prototype note for the first Prototype-to-Plan pass.",
	)
	if err != nil {
		t.Fatalf("expected work object prototype update: %v", err)
	}

	if workObject.PrototypeURL != "https://example.com/prototype/mobile-perf-v2" {
		t.Fatalf("expected updated prototype url, got %q", workObject.PrototypeURL)
	}
	if workObject.PrototypeNote != "Updated prototype note for the first Prototype-to-Plan pass." {
		t.Fatalf("expected updated prototype note, got %q", workObject.PrototypeNote)
	}
	if !workObject.UpdatedAt.Equal(updatedAt) {
		t.Fatalf("expected updated timestamp %s, got %s", updatedAt, workObject.UpdatedAt)
	}

	var persistedURL string
	var persistedNote string
	var persistedUpdatedAt string
	if err := store.DB.QueryRow(`SELECT prototype_url, COALESCE(prototype_note, ''), updated_at FROM work_objects WHERE id = ?`, "wo_mobile_perf").Scan(&persistedURL, &persistedNote, &persistedUpdatedAt); err != nil {
		t.Fatalf("expected persisted work object prototype: %v", err)
	}
	if persistedURL != "https://example.com/prototype/mobile-perf-v2" {
		t.Fatalf("expected persisted prototype url, got %q", persistedURL)
	}
	if persistedNote != "Updated prototype note for the first Prototype-to-Plan pass." {
		t.Fatalf("expected persisted prototype note, got %q", persistedNote)
	}
	if persistedUpdatedAt != updatedAt.Format(time.RFC3339) {
		t.Fatalf("expected persisted updated_at %q, got %q", updatedAt.Format(time.RFC3339), persistedUpdatedAt)
	}
}

func TestUpdateWorkObjectPrototypeRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	store := newTestStore(t, true)
	if _, err := store.UpdateWorkObjectPrototype(context.Background(), "wo_mobile_perf", "   ", "   "); !errors.Is(err, ErrInvalidWorkObjectPrototypeFields) {
		t.Fatalf("expected invalid prototype error, got %v", err)
	}
}

func TestUpdateWorkObjectPrototypeReturnsNotFound(t *testing.T) {
	t.Parallel()

	store := newTestStore(t, true)
	if _, err := store.UpdateWorkObjectPrototype(context.Background(), "missing", "https://example.com/prototype/missing", "Prototype note"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestUpdateWorkObjectPriorityPersistsChanges(t *testing.T) {
	store := newTestStore(t, true)
	originalNow := workObjectNow
	t.Cleanup(func() { workObjectNow = originalNow })

	updatedAt := time.Date(2026, 3, 29, 16, 45, 0, 0, time.UTC)
	workObjectNow = func() time.Time { return updatedAt }

	workObject, err := store.UpdateWorkObjectPriority(context.Background(), "wo_mobile_perf", "urgent")
	if err != nil {
		t.Fatalf("expected work object priority update: %v", err)
	}

	if workObject.Priority != "urgent" {
		t.Fatalf("expected updated priority, got %q", workObject.Priority)
	}
	if !workObject.UpdatedAt.Equal(updatedAt) {
		t.Fatalf("expected updated timestamp %s, got %s", updatedAt, workObject.UpdatedAt)
	}

	var persistedPriority string
	var persistedUpdatedAt string
	if err := store.DB.QueryRow(`SELECT priority, updated_at FROM work_objects WHERE id = ?`, "wo_mobile_perf").Scan(&persistedPriority, &persistedUpdatedAt); err != nil {
		t.Fatalf("expected persisted work object priority: %v", err)
	}
	if persistedPriority != "urgent" {
		t.Fatalf("expected persisted urgent priority, got %q", persistedPriority)
	}
	if persistedUpdatedAt != updatedAt.Format(time.RFC3339) {
		t.Fatalf("expected persisted updated_at %q, got %q", updatedAt.Format(time.RFC3339), persistedUpdatedAt)
	}

	workObjects, err := store.ListWorkObjects(context.Background())
	if err != nil {
		t.Fatalf("expected work object list after update: %v", err)
	}

	for _, entry := range workObjects {
		if entry.ID == "wo_mobile_perf" && entry.Priority != "urgent" {
			t.Fatalf("expected list view to return updated priority, got %q", entry.Priority)
		}
	}
}

func TestUpdateWorkObjectPriorityRejectsInvalidPriority(t *testing.T) {
	t.Parallel()

	store := newTestStore(t, true)
	if _, err := store.UpdateWorkObjectPriority(context.Background(), "wo_mobile_perf", "critical"); !errors.Is(err, ErrInvalidWorkObjectPriority) {
		t.Fatalf("expected invalid priority error, got %v", err)
	}
}

func TestUpdateWorkObjectPriorityReturnsNotFound(t *testing.T) {
	t.Parallel()

	store := newTestStore(t, true)
	if _, err := store.UpdateWorkObjectPriority(context.Background(), "missing", "high"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestMarkWorkObjectReadyForReleasePersistsChanges(t *testing.T) {
	store := newTestStore(t, true)
	originalNow := workObjectNow
	t.Cleanup(func() { workObjectNow = originalNow })

	if _, err := store.GenerateDraftPlanForWorkObject(context.Background(), "wo_onboarding_refresh"); err != nil {
		t.Fatalf("expected draft plan generation: %v", err)
	}

	releaseReadyAt := time.Date(2026, 3, 31, 16, 10, 0, 0, time.UTC)
	workObjectNow = func() time.Time { return releaseReadyAt }

	workObject, err := store.MarkWorkObjectReadyForRelease(context.Background(), "wo_onboarding_refresh")
	if err != nil {
		t.Fatalf("expected mark ready for release: %v", err)
	}

	if workObject.ReleaseReadyAt == nil {
		t.Fatal("expected release ready timestamp to be set")
	}
	if !workObject.ReleaseReadyAt.Equal(releaseReadyAt) {
		t.Fatalf("expected release ready timestamp %s, got %s", releaseReadyAt, workObject.ReleaseReadyAt)
	}
	if !workObject.UpdatedAt.Equal(releaseReadyAt) {
		t.Fatalf("expected updated timestamp %s, got %s", releaseReadyAt, workObject.UpdatedAt)
	}

	var persistedReleaseReadyAt string
	var persistedUpdatedAt string
	if err := store.DB.QueryRow(
		`SELECT COALESCE(release_ready_at, ''), updated_at FROM work_objects WHERE id = ?`,
		"wo_onboarding_refresh",
	).Scan(&persistedReleaseReadyAt, &persistedUpdatedAt); err != nil {
		t.Fatalf("expected persisted work object release-ready state: %v", err)
	}
	if persistedReleaseReadyAt != releaseReadyAt.Format(time.RFC3339) {
		t.Fatalf("expected persisted release_ready_at %q, got %q", releaseReadyAt.Format(time.RFC3339), persistedReleaseReadyAt)
	}
	if persistedUpdatedAt != releaseReadyAt.Format(time.RFC3339) {
		t.Fatalf("expected persisted updated_at %q, got %q", releaseReadyAt.Format(time.RFC3339), persistedUpdatedAt)
	}

	workObjects, err := store.ListWorkObjects(context.Background())
	if err != nil {
		t.Fatalf("expected work object list after review update: %v", err)
	}
	for _, entry := range workObjects {
		if entry.ID == "wo_onboarding_refresh" && entry.ReleaseReadyAt == nil {
			t.Fatal("expected list view to return persisted release-ready state")
		}
	}
}

func TestMarkWorkObjectReadyForReleaseReturnsNotFound(t *testing.T) {
	t.Parallel()

	store := newTestStore(t, true)
	if _, err := store.MarkWorkObjectReadyForRelease(context.Background(), "missing"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestMarkWorkObjectReadyForReleaseRequiresDraftPlan(t *testing.T) {
	t.Parallel()

	store := newTestStore(t, true)
	if _, err := store.MarkWorkObjectReadyForRelease(context.Background(), "wo_mobile_perf"); !errors.Is(err, ErrDraftPlanNotFound) {
		t.Fatalf("expected draft plan error, got %v", err)
	}
}

func TestCreateDecisionForWorkObjectPersistsDecision(t *testing.T) {
	store := newTestStore(t, true)
	originalNow := decisionNow
	originalID := decisionIDFromWorkObject
	t.Cleanup(func() {
		decisionNow = originalNow
		decisionIDFromWorkObject = originalID
	})

	createdAt := time.Date(2026, 3, 29, 18, 5, 0, 0, time.UTC)
	decisionNow = func() time.Time { return createdAt }
	decisionIDFromWorkObject = func(workObjectID string, _ time.Time) string {
		return "dec_new_decision"
	}

	decision, err := store.CreateDecisionForWorkObject(
		context.Background(),
		"wo_mobile_perf",
		"Keep the first rollout tightly scoped",
		"Capture the smallest high-signal product move first.",
		"That keeps execution moving while the linked Cluster context is still fresh.",
	)
	if err != nil {
		t.Fatalf("expected decision creation: %v", err)
	}

	if decision.ID != "dec_new_decision" {
		t.Fatalf("expected deterministic decision id, got %q", decision.ID)
	}
	if decision.WorkObjectID != "wo_mobile_perf" {
		t.Fatalf("expected work object link, got %q", decision.WorkObjectID)
	}
	if !decision.CreatedAt.Equal(createdAt) {
		t.Fatalf("expected created time %s, got %s", createdAt, decision.CreatedAt)
	}

	var persistedTitle, persistedSummary, persistedRationale, persistedCreatedAt string
	if err := store.DB.QueryRow(
		`SELECT title, summary, rationale, created_at FROM decisions WHERE id = ?`,
		"dec_new_decision",
	).Scan(&persistedTitle, &persistedSummary, &persistedRationale, &persistedCreatedAt); err != nil {
		t.Fatalf("expected persisted decision: %v", err)
	}
	if persistedTitle != "Keep the first rollout tightly scoped" {
		t.Fatalf("expected persisted title, got %q", persistedTitle)
	}
	if persistedSummary != "Capture the smallest high-signal product move first." {
		t.Fatalf("expected persisted summary, got %q", persistedSummary)
	}
	if persistedRationale != "That keeps execution moving while the linked Cluster context is still fresh." {
		t.Fatalf("expected persisted rationale, got %q", persistedRationale)
	}
	if persistedCreatedAt != createdAt.Format(time.RFC3339) {
		t.Fatalf("expected persisted created_at %q, got %q", createdAt.Format(time.RFC3339), persistedCreatedAt)
	}

	workObject, err := store.lookupWorkObject(context.Background(), "wo_mobile_perf")
	if err != nil {
		t.Fatalf("expected work object lookup after decision create: %v", err)
	}
	found := false
	for _, entry := range workObject.Decisions {
		if entry.ID == "dec_new_decision" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected new decision to appear in work object detail data")
	}
}

func TestCreateDecisionForWorkObjectRejectsInvalidFields(t *testing.T) {
	t.Parallel()

	store := newTestStore(t, true)
	if _, err := store.CreateDecisionForWorkObject(context.Background(), "wo_mobile_perf", "  ", "Valid summary", "Valid rationale"); !errors.Is(err, ErrInvalidDecisionFields) {
		t.Fatalf("expected invalid decision error, got %v", err)
	}
}

func TestCreateDecisionForWorkObjectReturnsNotFound(t *testing.T) {
	t.Parallel()

	store := newTestStore(t, true)
	if _, err := store.CreateDecisionForWorkObject(context.Background(), "missing", "Valid title", "Valid summary", "Valid rationale"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestCreateArtifactForWorkObjectPersistsArtifact(t *testing.T) {
	store := newTestStore(t, true)
	originalNow := artifactNow
	originalID := artifactIDFromWorkObject
	t.Cleanup(func() {
		artifactNow = originalNow
		artifactIDFromWorkObject = originalID
	})

	createdAt := time.Date(2026, 3, 29, 19, 5, 0, 0, time.UTC)
	artifactNow = func() time.Time { return createdAt }
	artifactIDFromWorkObject = func(workObjectID string, _ time.Time) string {
		return "art_new_artifact"
	}

	artifact, err := store.CreateArtifactForWorkObject(
		context.Background(),
		"wo_mobile_perf",
		"note",
		"Mobile loading state note",
		"Document the lightweight loading treatment for the first pass.",
	)
	if err != nil {
		t.Fatalf("expected artifact creation: %v", err)
	}

	if artifact.ID != "art_new_artifact" {
		t.Fatalf("expected deterministic artifact id, got %q", artifact.ID)
	}
	if artifact.WorkObjectID != "wo_mobile_perf" {
		t.Fatalf("expected work object link, got %q", artifact.WorkObjectID)
	}
	if artifact.ArtifactType != "note" {
		t.Fatalf("expected artifact type note, got %q", artifact.ArtifactType)
	}
	if !artifact.CreatedAt.Equal(createdAt) {
		t.Fatalf("expected created time %s, got %s", createdAt, artifact.CreatedAt)
	}

	var persistedType, persistedTitle, persistedBody, persistedCreatedAt string
	if err := store.DB.QueryRow(
		`SELECT artifact_type, title, body, created_at FROM artifacts WHERE id = ?`,
		"art_new_artifact",
	).Scan(&persistedType, &persistedTitle, &persistedBody, &persistedCreatedAt); err != nil {
		t.Fatalf("expected persisted artifact: %v", err)
	}
	if persistedType != "note" {
		t.Fatalf("expected persisted type, got %q", persistedType)
	}
	if persistedTitle != "Mobile loading state note" {
		t.Fatalf("expected persisted title, got %q", persistedTitle)
	}
	if persistedBody != "Document the lightweight loading treatment for the first pass." {
		t.Fatalf("expected persisted body, got %q", persistedBody)
	}
	if persistedCreatedAt != createdAt.Format(time.RFC3339) {
		t.Fatalf("expected persisted created_at %q, got %q", createdAt.Format(time.RFC3339), persistedCreatedAt)
	}

	workObject, err := store.lookupWorkObject(context.Background(), "wo_mobile_perf")
	if err != nil {
		t.Fatalf("expected work object lookup after artifact create: %v", err)
	}
	found := false
	for _, entry := range workObject.Artifacts {
		if entry.ID == "art_new_artifact" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected new artifact to appear in work object detail data")
	}
}

func TestCreateArtifactForWorkObjectRejectsInvalidFields(t *testing.T) {
	t.Parallel()

	store := newTestStore(t, true)
	if _, err := store.CreateArtifactForWorkObject(context.Background(), "wo_mobile_perf", "  ", "Valid title", "Valid body"); !errors.Is(err, ErrInvalidArtifactFields) {
		t.Fatalf("expected invalid artifact error, got %v", err)
	}
}

func TestCreateArtifactForWorkObjectReturnsNotFound(t *testing.T) {
	t.Parallel()

	store := newTestStore(t, true)
	if _, err := store.CreateArtifactForWorkObject(context.Background(), "missing", "note", "Valid title", "Valid body"); !errors.Is(err, sql.ErrNoRows) {
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

func TestIsValidWorkObjectStatus(t *testing.T) {
	t.Parallel()

	for _, status := range []string{"planning", "ready", "in_progress", "in_review", "shipped"} {
		if !isValidWorkObjectStatus(status) {
			t.Fatalf("expected %q to be valid", status)
		}
	}
	if isValidWorkObjectStatus("review") {
		t.Fatal("expected legacy review status to be invalid")
	}
}
