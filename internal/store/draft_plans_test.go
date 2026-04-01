package store

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"cleartify/internal/domain"
)

func TestLookupDraftPlanForWorkObjectReturnsNilWhenMissing(t *testing.T) {
	t.Parallel()

	store := newTestStore(t, true)
	draftPlan, err := store.LookupDraftPlanForWorkObject(context.Background(), "wo_mobile_perf")
	if err != nil {
		t.Fatalf("expected missing draft plan lookup to succeed: %v", err)
	}
	if draftPlan != nil {
		t.Fatalf("expected nil draft plan, got %#v", draftPlan)
	}
}

func TestGenerateDraftPlanForWorkObjectPersistsDeterministicPlan(t *testing.T) {
	originalNow := draftPlanNow
	t.Cleanup(func() { draftPlanNow = originalNow })

	store := newTestStore(t, true)
	generatedAt := time.Date(2026, 3, 29, 21, 0, 0, 0, time.UTC)
	draftPlanNow = func() time.Time { return generatedAt }

	draftPlan, err := store.GenerateDraftPlanForWorkObject(context.Background(), "wo_onboarding_refresh")
	if err != nil {
		t.Fatalf("expected draft plan generation: %v", err)
	}

	if draftPlan.WorkObjectID != "wo_onboarding_refresh" {
		t.Fatalf("expected work object id, got %q", draftPlan.WorkObjectID)
	}
	if !strings.Contains(draftPlan.Overview, "Guided onboarding refresh") {
		t.Fatalf("expected overview to reference work object title, got %q", draftPlan.Overview)
	}
	if !strings.Contains(draftPlan.Overview, "attached prototype link and note") {
		t.Fatalf("expected overview to reference prototype context, got %q", draftPlan.Overview)
	}
	if !strings.Contains(draftPlan.Overview, "Linked Cluster context from \"Onboarding activation clarity\"") {
		t.Fatalf("expected overview to reference linked cluster, got %q", draftPlan.Overview)
	}
	if len(draftPlan.ImplementationSteps) != 4 {
		t.Fatalf("expected 4 implementation steps, got %d", len(draftPlan.ImplementationSteps))
	}
	if len(draftPlan.EdgeCases) != 3 {
		t.Fatalf("expected 3 edge cases, got %d", len(draftPlan.EdgeCases))
	}
	if len(draftPlan.OpenQuestions) != 3 {
		t.Fatalf("expected 3 open questions, got %d", len(draftPlan.OpenQuestions))
	}
	if !draftPlan.CreatedAt.Equal(generatedAt) || !draftPlan.UpdatedAt.Equal(generatedAt) {
		t.Fatalf("expected deterministic timestamps, got created=%s updated=%s", draftPlan.CreatedAt, draftPlan.UpdatedAt)
	}

	loadedPlan, err := store.LookupDraftPlanForWorkObject(context.Background(), "wo_onboarding_refresh")
	if err != nil {
		t.Fatalf("expected persisted draft plan lookup: %v", err)
	}
	if loadedPlan == nil {
		t.Fatal("expected persisted draft plan")
	}
	if loadedPlan.Overview != draftPlan.Overview {
		t.Fatalf("expected persisted overview %q, got %q", draftPlan.Overview, loadedPlan.Overview)
	}
	if loadedPlan.ImplementationSteps[0] != draftPlan.ImplementationSteps[0] {
		t.Fatalf("expected persisted implementation step %q, got %q", draftPlan.ImplementationSteps[0], loadedPlan.ImplementationSteps[0])
	}
}

func TestGenerateDraftPlanForWorkObjectRefreshPreservesCreatedAt(t *testing.T) {
	originalNow := draftPlanNow
	t.Cleanup(func() { draftPlanNow = originalNow })

	store := newTestStore(t, true)
	initialTime := time.Date(2026, 3, 29, 21, 5, 0, 0, time.UTC)
	refreshedTime := time.Date(2026, 3, 29, 21, 6, 0, 0, time.UTC)

	draftPlanNow = func() time.Time { return initialTime }
	firstPlan, err := store.GenerateDraftPlanForWorkObject(context.Background(), "wo_mobile_perf")
	if err != nil {
		t.Fatalf("expected first draft plan generation: %v", err)
	}

	draftPlanNow = func() time.Time { return refreshedTime }
	refreshedPlan, err := store.GenerateDraftPlanForWorkObject(context.Background(), "wo_mobile_perf")
	if err != nil {
		t.Fatalf("expected refreshed draft plan generation: %v", err)
	}

	if !refreshedPlan.CreatedAt.Equal(firstPlan.CreatedAt) {
		t.Fatalf("expected created_at to stay %s, got %s", firstPlan.CreatedAt, refreshedPlan.CreatedAt)
	}
	if !refreshedPlan.UpdatedAt.Equal(refreshedTime) {
		t.Fatalf("expected updated_at %s, got %s", refreshedTime, refreshedPlan.UpdatedAt)
	}
}

func TestGenerateDraftPlanForWorkObjectReturnsNotFound(t *testing.T) {
	t.Parallel()

	store := newTestStore(t, true)
	if _, err := store.GenerateDraftPlanForWorkObject(context.Background(), "missing"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected missing work object error, got %v", err)
	}
}

func TestUpdateDraftPlanForWorkObjectPersistsEdits(t *testing.T) {
	originalNow := draftPlanNow
	t.Cleanup(func() { draftPlanNow = originalNow })

	store := newTestStore(t, true)
	initialTime := time.Date(2026, 3, 29, 22, 0, 0, 0, time.UTC)
	editedTime := time.Date(2026, 3, 29, 22, 5, 0, 0, time.UTC)
	draftPlanNow = func() time.Time { return initialTime }

	if _, err := store.GenerateDraftPlanForWorkObject(context.Background(), "wo_onboarding_refresh"); err != nil {
		t.Fatalf("expected initial draft plan generation: %v", err)
	}

	draftPlanNow = func() time.Time { return editedTime }
	updatedPlan, err := store.UpdateDraftPlanForWorkObject(
		context.Background(),
		"wo_onboarding_refresh",
		"Edited overview for the onboarding refresh.",
		[]string{"Review the revised prototype set.", "Confirm the smallest onboarding slice."},
		[]string{"The prototype set may still change before implementation."},
		[]string{"Which onboarding moments should stay in scope for the first pass?"},
	)
	if err != nil {
		t.Fatalf("expected draft plan update: %v", err)
	}

	if updatedPlan.Overview != "Edited overview for the onboarding refresh." {
		t.Fatalf("expected updated overview, got %q", updatedPlan.Overview)
	}
	if updatedPlan.ImplementationSteps[0] != "Review the revised prototype set." {
		t.Fatalf("expected updated implementation steps, got %#v", updatedPlan.ImplementationSteps)
	}
	if !updatedPlan.UpdatedAt.Equal(editedTime) {
		t.Fatalf("expected edited updated_at %s, got %s", editedTime, updatedPlan.UpdatedAt)
	}

	loadedPlan, err := store.LookupDraftPlanForWorkObject(context.Background(), "wo_onboarding_refresh")
	if err != nil {
		t.Fatalf("expected persisted edited draft plan: %v", err)
	}
	if loadedPlan == nil {
		t.Fatal("expected persisted edited draft plan")
	}
	if loadedPlan.Overview != "Edited overview for the onboarding refresh." {
		t.Fatalf("expected persisted overview, got %q", loadedPlan.Overview)
	}
	if loadedPlan.OpenQuestions[0] != "Which onboarding moments should stay in scope for the first pass?" {
		t.Fatalf("expected persisted open question, got %#v", loadedPlan.OpenQuestions)
	}
}

func TestUpdateDraftPlanForWorkObjectRejectsInvalidFields(t *testing.T) {
	t.Parallel()

	store := newTestStore(t, true)
	if _, err := store.GenerateDraftPlanForWorkObject(context.Background(), "wo_mobile_perf"); err != nil {
		t.Fatalf("expected initial draft plan generation: %v", err)
	}

	if _, err := store.UpdateDraftPlanForWorkObject(context.Background(), "wo_mobile_perf", "  ", []string{"Valid step"}, []string{"Valid edge case"}, []string{"Valid question"}); !errors.Is(err, ErrInvalidDraftPlanFields) {
		t.Fatalf("expected invalid overview error, got %v", err)
	}
	if _, err := store.UpdateDraftPlanForWorkObject(context.Background(), "wo_mobile_perf", "Valid overview", []string{"   "}, []string{"Valid edge case"}, []string{"Valid question"}); !errors.Is(err, ErrInvalidDraftPlanFields) {
		t.Fatalf("expected invalid implementation steps error, got %v", err)
	}
}

func TestUpdateDraftPlanForWorkObjectReturnsNotFound(t *testing.T) {
	t.Parallel()

	store := newTestStore(t, true)
	if _, err := store.UpdateDraftPlanForWorkObject(context.Background(), "missing", "Overview", []string{"Step"}, []string{"Edge case"}, []string{"Question"}); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected missing work object error, got %v", err)
	}
}

func TestUpdateDraftPlanForWorkObjectReturnsDraftPlanNotFound(t *testing.T) {
	t.Parallel()

	store := newTestStore(t, true)
	if _, err := store.UpdateDraftPlanForWorkObject(context.Background(), "wo_mobile_perf", "Overview", []string{"Step"}, []string{"Edge case"}, []string{"Question"}); !errors.Is(err, ErrDraftPlanNotFound) {
		t.Fatalf("expected missing draft plan error, got %v", err)
	}
}

func TestLookupDraftPlanForWorkObjectReturnsDecodeErrors(t *testing.T) {
	t.Parallel()

	store := newTestStore(t, true)
	mustExec(
		t,
		store.DB,
		`INSERT INTO draft_plans (work_object_id, overview, implementation_steps_json, edge_cases_json, open_questions_json, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"wo_mobile_perf",
		"Bad plan",
		"{",
		`["one"]`,
		`["two"]`,
		"2026-03-29T21:10:00Z",
		"2026-03-29T21:10:00Z",
	)

	if _, err := store.LookupDraftPlanForWorkObject(context.Background(), "wo_mobile_perf"); err == nil {
		t.Fatal("expected invalid draft plan json error")
	}
}

func TestBuildDraftPlanWithoutPrototypeOrClusterUsesFallbacks(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 3, 29, 21, 15, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 3, 29, 21, 20, 0, 0, time.UTC)

	draftPlan := buildDraftPlan(
		domain.WorkObject{
			ID:      "wo_standalone",
			Title:   "Standalone work object",
			Summary: "A seeded summary without prototype context.",
			Goal:    "Clarify the smallest fallback plan path.",
		},
		nil,
		createdAt,
		updatedAt,
	)

	if !strings.Contains(draftPlan.Overview, "stays grounded in the Work Object itself") {
		t.Fatalf("expected fallback cluster context, got %q", draftPlan.Overview)
	}
	if !strings.Contains(draftPlan.Overview, "prototype note should be expanded") {
		t.Fatalf("expected fallback prototype context, got %q", draftPlan.Overview)
	}
	if draftPlan.ImplementationSteps[2] != "Capture any missing context that would normally come from a linked Cluster before expanding the scope." {
		t.Fatalf("expected fallback implementation step, got %q", draftPlan.ImplementationSteps[2])
	}
	if draftPlan.EdgeCases[2] != "This Work Object has no linked Cluster, so surrounding Signal context may still be incomplete." {
		t.Fatalf("expected fallback edge case, got %q", draftPlan.EdgeCases[2])
	}
}
