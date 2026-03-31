package server

import (
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"cleartify/internal/auth"
	"cleartify/internal/config"
	"cleartify/internal/domain"
	"cleartify/internal/store"
)

func TestHandleWorkObjectActionsUpdatesStatus(t *testing.T) {
	t.Parallel()

	server, err := New(
		config.Config{AppMarketingURL: "http://localhost:4173"},
		&fakeStore{updatedWorkObject: domain.WorkObject{ID: "wo_1", Title: "Updated Work Object", Status: "in_progress"}},
		auth.NewDemoAuthService(),
		newTestServer(t).frontend,
	)
	if err != nil {
		t.Fatalf("expected server build: %v", err)
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/work-objects/wo_1/status", strings.NewReader(`{"status":"in_progress"}`))
	server.handleWorkObjectActions(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), `"status":"in_progress"`) {
		t.Fatalf("expected updated work object status in response, got %q", recorder.Body.String())
	}
}

func TestHandleWorkObjectActionsLoadsDraftPlan(t *testing.T) {
	t.Parallel()

	server, err := New(
		config.Config{AppMarketingURL: "http://localhost:4173"},
		&fakeStore{
			lookupDraftPlan: &domain.DraftPlan{
				WorkObjectID:        "wo_1",
				Overview:            "Persisted overview",
				ImplementationSteps: []string{"Step 1"},
				EdgeCases:           []string{"Edge case 1"},
				OpenQuestions:       []string{"Question 1"},
			},
		},
		auth.NewDemoAuthService(),
		newTestServer(t).frontend,
	)
	if err != nil {
		t.Fatalf("expected server build: %v", err)
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/work-objects/wo_1/plan", nil)
	server.handleWorkObjectActions(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), `"overview":"Persisted overview"`) {
		t.Fatalf("expected persisted draft plan in response, got %q", recorder.Body.String())
	}
}

func TestHandleWorkObjectActionsGeneratesDraftPlan(t *testing.T) {
	t.Parallel()

	server, err := New(
		config.Config{AppMarketingURL: "http://localhost:4173"},
		&fakeStore{
			generatedDraftPlan: domain.DraftPlan{
				WorkObjectID:        "wo_1",
				Overview:            "Generated overview",
				ImplementationSteps: []string{"Step 1"},
				EdgeCases:           []string{"Edge case 1"},
				OpenQuestions:       []string{"Question 1"},
			},
		},
		auth.NewDemoAuthService(),
		newTestServer(t).frontend,
	)
	if err != nil {
		t.Fatalf("expected server build: %v", err)
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/work-objects/wo_1/plan", nil)
	server.handleWorkObjectActions(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), `"overview":"Generated overview"`) {
		t.Fatalf("expected generated draft plan in response, got %q", recorder.Body.String())
	}
}

func TestHandleWorkObjectActionsUpdatesDraftPlan(t *testing.T) {
	t.Parallel()

	server, err := New(
		config.Config{AppMarketingURL: "http://localhost:4173"},
		&fakeStore{
			updatedDraftPlan: domain.DraftPlan{
				WorkObjectID:        "wo_1",
				Overview:            "Edited overview",
				ImplementationSteps: []string{"Edited step"},
				EdgeCases:           []string{"Edited edge case"},
				OpenQuestions:       []string{"Edited question"},
			},
		},
		auth.NewDemoAuthService(),
		newTestServer(t).frontend,
	)
	if err != nil {
		t.Fatalf("expected server build: %v", err)
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPut,
		"/api/work-objects/wo_1/plan",
		strings.NewReader(`{"overview":"Edited overview","implementationSteps":["Edited step"],"edgeCases":["Edited edge case"],"openQuestions":["Edited question"]}`),
	)
	server.handleWorkObjectActions(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), `"overview":"Edited overview"`) {
		t.Fatalf("expected updated draft plan in response, got %q", recorder.Body.String())
	}
}

func TestHandleWorkObjectActionsRejectsInvalidStatus(t *testing.T) {
	t.Parallel()

	server, err := New(
		config.Config{AppMarketingURL: "http://localhost:4173"},
		&fakeStore{err: store.ErrInvalidWorkObjectStatus},
		auth.NewDemoAuthService(),
		newTestServer(t).frontend,
	)
	if err != nil {
		t.Fatalf("expected server build: %v", err)
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/work-objects/wo_1/status", strings.NewReader(`{"status":"review"}`))
	server.handleWorkObjectActions(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}
}

func TestHandleWorkObjectActionsUpdatesPriority(t *testing.T) {
	t.Parallel()

	server, err := New(
		config.Config{AppMarketingURL: "http://localhost:4173"},
		&fakeStore{updatedWorkObject: domain.WorkObject{ID: "wo_1", Title: "Updated Work Object", Priority: "urgent"}},
		auth.NewDemoAuthService(),
		newTestServer(t).frontend,
	)
	if err != nil {
		t.Fatalf("expected server build: %v", err)
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/work-objects/wo_1/priority", strings.NewReader(`{"priority":"urgent"}`))
	server.handleWorkObjectActions(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), `"priority":"urgent"`) {
		t.Fatalf("expected updated work object priority in response, got %q", recorder.Body.String())
	}
}

func TestHandleWorkObjectActionsCreatesDecision(t *testing.T) {
	t.Parallel()

	server, err := New(
		config.Config{AppMarketingURL: "http://localhost:4173"},
		&fakeStore{
			createdDecision: domain.Decision{
				ID:           "dec_1",
				WorkObjectID: "wo_1",
				Title:        "Keep the rollout narrow",
				Summary:      "Focus the first pass on the smallest useful scope.",
				Rationale:    "This preserves momentum while product context is still fresh.",
			},
		},
		auth.NewDemoAuthService(),
		newTestServer(t).frontend,
	)
	if err != nil {
		t.Fatalf("expected server build: %v", err)
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/work-objects/wo_1/decisions",
		strings.NewReader(`{"title":"Keep the rollout narrow","summary":"Focus the first pass on the smallest useful scope.","rationale":"This preserves momentum while product context is still fresh."}`),
	)
	server.handleWorkObjectActions(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), `"title":"Keep the rollout narrow"`) {
		t.Fatalf("expected created decision in response, got %q", recorder.Body.String())
	}
}

func TestHandleWorkObjectActionsCreatesArtifact(t *testing.T) {
	t.Parallel()

	server, err := New(
		config.Config{AppMarketingURL: "http://localhost:4173"},
		&fakeStore{
			createdArtifact: domain.Artifact{
				ID:           "art_1",
				WorkObjectID: "wo_1",
				ArtifactType: "note",
				Title:        "Mobile rollout note",
				Body:         "Capture the first pass output in one concise place.",
			},
		},
		auth.NewDemoAuthService(),
		newTestServer(t).frontend,
	)
	if err != nil {
		t.Fatalf("expected server build: %v", err)
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/work-objects/wo_1/artifacts",
		strings.NewReader(`{"type":"note","title":"Mobile rollout note","body":"Capture the first pass output in one concise place."}`),
	)
	server.handleWorkObjectActions(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), `"artifactType":"note"`) {
		t.Fatalf("expected created artifact in response, got %q", recorder.Body.String())
	}
}

func TestHandleWorkObjectActionsGeneratesImplementationArtifact(t *testing.T) {
	t.Parallel()

	server, err := New(
		config.Config{AppMarketingURL: "http://localhost:4173"},
		&fakeStore{
			generatedArtifact: domain.Artifact{
				ID:           "art_impl_1",
				WorkObjectID: "wo_1",
				ArtifactType: "implementation_brief",
				Title:        "Updated Work Object implementation brief",
				Body:         "Overview\nDeterministic implementation brief.",
			},
		},
		auth.NewDemoAuthService(),
		newTestServer(t).frontend,
	)
	if err != nil {
		t.Fatalf("expected server build: %v", err)
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/work-objects/wo_1/implementation-artifact", nil)
	server.handleWorkObjectActions(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), `"artifactType":"implementation_brief"`) {
		t.Fatalf("expected generated implementation artifact in response, got %q", recorder.Body.String())
	}
}

func TestHandleWorkObjectActionsUpdatesCoreFields(t *testing.T) {
	t.Parallel()

	server, err := New(
		config.Config{AppMarketingURL: "http://localhost:4173"},
		&fakeStore{
			updatedWorkObject: domain.WorkObject{
				ID:      "wo_1",
				Title:   "Updated Work Object",
				Summary: "Sharper summary.",
				Goal:    "Keep the goal grounded in product context.",
			},
		},
		auth.NewDemoAuthService(),
		newTestServer(t).frontend,
	)
	if err != nil {
		t.Fatalf("expected server build: %v", err)
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/work-objects/wo_1/core",
		strings.NewReader(`{"title":"Updated Work Object","summary":"Sharper summary.","goal":"Keep the goal grounded in product context."}`),
	)
	server.handleWorkObjectActions(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), `"title":"Updated Work Object"`) {
		t.Fatalf("expected updated work object title in response, got %q", recorder.Body.String())
	}
}

func TestHandleWorkObjectActionsUpdatesPrototype(t *testing.T) {
	t.Parallel()

	server, err := New(
		config.Config{AppMarketingURL: "http://localhost:4173"},
		&fakeStore{
			updatedWorkObject: domain.WorkObject{
				ID:            "wo_1",
				Title:         "Updated Work Object",
				PrototypeURL:  "https://example.com/prototype/1",
				PrototypeNote: "Prototype note",
			},
		},
		auth.NewDemoAuthService(),
		newTestServer(t).frontend,
	)
	if err != nil {
		t.Fatalf("expected server build: %v", err)
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/work-objects/wo_1/prototype",
		strings.NewReader(`{"prototypeUrl":"https://example.com/prototype/1","prototypeNote":"Prototype note"}`),
	)
	server.handleWorkObjectActions(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), `"prototypeUrl":"https://example.com/prototype/1"`) {
		t.Fatalf("expected updated prototype in response, got %q", recorder.Body.String())
	}
}

func TestHandleWorkObjectActionsRejectsInvalidPriority(t *testing.T) {
	t.Parallel()

	server, err := New(
		config.Config{AppMarketingURL: "http://localhost:4173"},
		&fakeStore{err: store.ErrInvalidWorkObjectPriority},
		auth.NewDemoAuthService(),
		newTestServer(t).frontend,
	)
	if err != nil {
		t.Fatalf("expected server build: %v", err)
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/work-objects/wo_1/priority", strings.NewReader(`{"priority":"critical"}`))
	server.handleWorkObjectActions(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}
}

func TestHandleWorkObjectActionsRejectsInvalidPrototype(t *testing.T) {
	t.Parallel()

	server, err := New(
		config.Config{AppMarketingURL: "http://localhost:4173"},
		&fakeStore{err: store.ErrInvalidWorkObjectPrototypeFields},
		auth.NewDemoAuthService(),
		newTestServer(t).frontend,
	)
	if err != nil {
		t.Fatalf("expected server build: %v", err)
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/work-objects/wo_1/prototype", strings.NewReader(`{"prototypeUrl":"","prototypeNote":"  "}`))
	server.handleWorkObjectActions(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}
}

func TestHandleWorkObjectActionsRejectsInvalidCoreFields(t *testing.T) {
	t.Parallel()

	server, err := New(
		config.Config{AppMarketingURL: "http://localhost:4173"},
		&fakeStore{err: store.ErrInvalidWorkObjectCoreFields},
		auth.NewDemoAuthService(),
		newTestServer(t).frontend,
	)
	if err != nil {
		t.Fatalf("expected server build: %v", err)
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/work-objects/wo_1/core", strings.NewReader(`{"title":"","summary":"Valid","goal":"Valid"}`))
	server.handleWorkObjectActions(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}
}

func TestHandleWorkObjectActionsReturnsNotFoundForMissingPrototypeWorkObject(t *testing.T) {
	t.Parallel()

	server, err := New(
		config.Config{AppMarketingURL: "http://localhost:4173"},
		&fakeStore{err: sql.ErrNoRows},
		auth.NewDemoAuthService(),
		newTestServer(t).frontend,
	)
	if err != nil {
		t.Fatalf("expected server build: %v", err)
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/work-objects/missing/prototype",
		strings.NewReader(`{"prototypeUrl":"https://example.com/prototype/missing","prototypeNote":"Prototype note"}`),
	)
	server.handleWorkObjectActions(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", recorder.Code)
	}
}

func TestHandleWorkObjectActionsRejectsInvalidDecision(t *testing.T) {
	t.Parallel()

	server, err := New(
		config.Config{AppMarketingURL: "http://localhost:4173"},
		&fakeStore{err: store.ErrInvalidDecisionFields},
		auth.NewDemoAuthService(),
		newTestServer(t).frontend,
	)
	if err != nil {
		t.Fatalf("expected server build: %v", err)
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/work-objects/wo_1/decisions", strings.NewReader(`{"title":"","summary":"Valid","rationale":"Valid"}`))
	server.handleWorkObjectActions(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}
}

func TestHandleWorkObjectActionsRejectsInvalidArtifact(t *testing.T) {
	t.Parallel()

	server, err := New(
		config.Config{AppMarketingURL: "http://localhost:4173"},
		&fakeStore{err: store.ErrInvalidArtifactFields},
		auth.NewDemoAuthService(),
		newTestServer(t).frontend,
	)
	if err != nil {
		t.Fatalf("expected server build: %v", err)
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/work-objects/wo_1/artifacts", strings.NewReader(`{"type":"","title":"Valid","body":"Valid"}`))
	server.handleWorkObjectActions(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}
}

func TestHandleWorkObjectActionsReturnsNotFoundForMissingDraftPlan(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/work-objects/wo_1/plan", nil)
	server.handleWorkObjectActions(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", recorder.Code)
	}
}

func TestHandleWorkObjectActionsReturnsNotFoundForMissingPriorityWorkObject(t *testing.T) {
	t.Parallel()

	server, err := New(
		config.Config{AppMarketingURL: "http://localhost:4173"},
		&fakeStore{err: sql.ErrNoRows},
		auth.NewDemoAuthService(),
		newTestServer(t).frontend,
	)
	if err != nil {
		t.Fatalf("expected server build: %v", err)
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/work-objects/missing/priority", strings.NewReader(`{"priority":"high"}`))
	server.handleWorkObjectActions(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", recorder.Code)
	}
}

func TestHandleWorkObjectActionsReturnsNotFoundForMissingCoreWorkObject(t *testing.T) {
	t.Parallel()

	server, err := New(
		config.Config{AppMarketingURL: "http://localhost:4173"},
		&fakeStore{err: sql.ErrNoRows},
		auth.NewDemoAuthService(),
		newTestServer(t).frontend,
	)
	if err != nil {
		t.Fatalf("expected server build: %v", err)
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/work-objects/missing/core", strings.NewReader(`{"title":"Valid","summary":"Valid","goal":"Valid"}`))
	server.handleWorkObjectActions(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", recorder.Code)
	}
}

func TestHandleWorkObjectActionsReturnsNotFoundForMissingWorkObject(t *testing.T) {
	t.Parallel()

	server, err := New(
		config.Config{AppMarketingURL: "http://localhost:4173"},
		&fakeStore{err: sql.ErrNoRows},
		auth.NewDemoAuthService(),
		newTestServer(t).frontend,
	)
	if err != nil {
		t.Fatalf("expected server build: %v", err)
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/work-objects/missing/status", strings.NewReader(`{"status":"ready"}`))
	server.handleWorkObjectActions(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", recorder.Code)
	}
}

func TestHandleWorkObjectActionsReturnsNotFoundForMissingDecisionWorkObject(t *testing.T) {
	t.Parallel()

	server, err := New(
		config.Config{AppMarketingURL: "http://localhost:4173"},
		&fakeStore{err: sql.ErrNoRows},
		auth.NewDemoAuthService(),
		newTestServer(t).frontend,
	)
	if err != nil {
		t.Fatalf("expected server build: %v", err)
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/work-objects/missing/decisions", strings.NewReader(`{"title":"Valid","summary":"Valid","rationale":"Valid"}`))
	server.handleWorkObjectActions(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", recorder.Code)
	}
}

func TestHandleWorkObjectActionsReturnsNotFoundForMissingArtifactWorkObject(t *testing.T) {
	t.Parallel()

	server, err := New(
		config.Config{AppMarketingURL: "http://localhost:4173"},
		&fakeStore{err: sql.ErrNoRows},
		auth.NewDemoAuthService(),
		newTestServer(t).frontend,
	)
	if err != nil {
		t.Fatalf("expected server build: %v", err)
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/work-objects/missing/artifacts", strings.NewReader(`{"type":"note","title":"Valid","body":"Valid"}`))
	server.handleWorkObjectActions(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", recorder.Code)
	}
}

func TestHandleWorkObjectActionsReturnsNotFoundForMissingPlanWorkObject(t *testing.T) {
	t.Parallel()

	server, err := New(
		config.Config{AppMarketingURL: "http://localhost:4173"},
		&fakeStore{err: sql.ErrNoRows},
		auth.NewDemoAuthService(),
		newTestServer(t).frontend,
	)
	if err != nil {
		t.Fatalf("expected server build: %v", err)
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/work-objects/missing/plan", nil)
	server.handleWorkObjectActions(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", recorder.Code)
	}
}

func TestHandleWorkObjectActionsReturnsNotFoundForMissingImplementationArtifactWorkObject(t *testing.T) {
	t.Parallel()

	server, err := New(
		config.Config{AppMarketingURL: "http://localhost:4173"},
		&fakeStore{err: sql.ErrNoRows},
		auth.NewDemoAuthService(),
		newTestServer(t).frontend,
	)
	if err != nil {
		t.Fatalf("expected server build: %v", err)
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/work-objects/missing/implementation-artifact", nil)
	server.handleWorkObjectActions(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", recorder.Code)
	}
}

func TestHandleWorkObjectActionsReturnsNotFoundForMissingPersistedDraftPlan(t *testing.T) {
	t.Parallel()

	server, err := New(
		config.Config{AppMarketingURL: "http://localhost:4173"},
		&fakeStore{err: store.ErrDraftPlanNotFound},
		auth.NewDemoAuthService(),
		newTestServer(t).frontend,
	)
	if err != nil {
		t.Fatalf("expected server build: %v", err)
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPut,
		"/api/work-objects/wo_1/plan",
		strings.NewReader(`{"overview":"Edited overview","implementationSteps":["Edited step"],"edgeCases":["Edited edge case"],"openQuestions":["Edited question"]}`),
	)
	server.handleWorkObjectActions(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", recorder.Code)
	}
}

func TestHandleWorkObjectActionsReturnsNotFoundForMissingDraftPlanDuringImplementationArtifactGeneration(t *testing.T) {
	t.Parallel()

	server, err := New(
		config.Config{AppMarketingURL: "http://localhost:4173"},
		&fakeStore{err: store.ErrDraftPlanNotFound},
		auth.NewDemoAuthService(),
		newTestServer(t).frontend,
	)
	if err != nil {
		t.Fatalf("expected server build: %v", err)
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/work-objects/wo_1/implementation-artifact", nil)
	server.handleWorkObjectActions(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", recorder.Code)
	}
}

func TestHandleWorkObjectActionsReturnsInternalErrorForUpdateFailure(t *testing.T) {
	t.Parallel()

	server, err := New(
		config.Config{AppMarketingURL: "http://localhost:4173"},
		&fakeStore{err: errors.New("boom")},
		auth.NewDemoAuthService(),
		newTestServer(t).frontend,
	)
	if err != nil {
		t.Fatalf("expected server build: %v", err)
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/work-objects/wo_1/status", strings.NewReader(`{"status":"ready"}`))
	server.handleWorkObjectActions(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", recorder.Code)
	}

	recorder = httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPost, "/api/work-objects/wo_1/core", strings.NewReader(`{"title":"Valid","summary":"Valid","goal":"Valid"}`))
	server.handleWorkObjectActions(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 for core field update, got %d", recorder.Code)
	}

	recorder = httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPost, "/api/work-objects/wo_1/decisions", strings.NewReader(`{"title":"Valid","summary":"Valid","rationale":"Valid"}`))
	server.handleWorkObjectActions(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 for decision create, got %d", recorder.Code)
	}

	recorder = httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPost, "/api/work-objects/wo_1/artifacts", strings.NewReader(`{"type":"note","title":"Valid","body":"Valid"}`))
	server.handleWorkObjectActions(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 for artifact create, got %d", recorder.Code)
	}

	recorder = httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodGet, "/api/work-objects/wo_1/plan", nil)
	server.handleWorkObjectActions(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 for draft plan lookup, got %d", recorder.Code)
	}

	recorder = httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPost, "/api/work-objects/wo_1/plan", nil)
	server.handleWorkObjectActions(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 for draft plan generate, got %d", recorder.Code)
	}

	recorder = httptest.NewRecorder()
	request = httptest.NewRequest(
		http.MethodPut,
		"/api/work-objects/wo_1/plan",
		strings.NewReader(`{"overview":"Edited overview","implementationSteps":["Edited step"],"edgeCases":["Edited edge case"],"openQuestions":["Edited question"]}`),
	)
	server.handleWorkObjectActions(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 for draft plan update, got %d", recorder.Code)
	}

	recorder = httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPost, "/api/work-objects/wo_1/implementation-artifact", nil)
	server.handleWorkObjectActions(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 for implementation artifact generate, got %d", recorder.Code)
	}
}

func TestHandleWorkObjectActionsRejectsBadBodies(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)

	for _, body := range []string{"", "{}", "{not-json"} {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/api/work-objects/wo_1/status", strings.NewReader(body))
		server.handleWorkObjectActions(recorder, request)

		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for body %q, got %d", body, recorder.Code)
		}
	}

	for _, body := range []string{"", "{}", "{not-json"} {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/api/work-objects/wo_1/core", strings.NewReader(body))
		server.handleWorkObjectActions(recorder, request)

		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for core body %q, got %d", body, recorder.Code)
		}
	}

	for _, body := range []string{"", "{}", "{not-json"} {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/api/work-objects/wo_1/priority", strings.NewReader(body))
		server.handleWorkObjectActions(recorder, request)

		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for priority body %q, got %d", body, recorder.Code)
		}
	}

	for _, body := range []string{"", "{}", "{not-json"} {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/api/work-objects/wo_1/decisions", strings.NewReader(body))
		server.handleWorkObjectActions(recorder, request)

		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for decision body %q, got %d", body, recorder.Code)
		}
	}

	for _, body := range []string{"", "{}", "{not-json"} {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/api/work-objects/wo_1/artifacts", strings.NewReader(body))
		server.handleWorkObjectActions(recorder, request)

		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for artifact body %q, got %d", body, recorder.Code)
		}
	}

	for _, body := range []string{"", "{not-json"} {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPut, "/api/work-objects/wo_1/plan", strings.NewReader(body))
		server.handleWorkObjectActions(recorder, request)

		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for draft plan body %q, got %d", body, recorder.Code)
		}
	}
}

func TestHandleWorkObjectActionsRejectsUnknownPaths(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/work-objects/wo_1/not-real", nil)
	server.handleWorkObjectActions(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown work object action, got %d", recorder.Code)
	}
}

func TestWorkObjectActionPath(t *testing.T) {
	t.Parallel()

	workObjectID, action, ok := workObjectActionPath("/api/work-objects/wo_1/status")
	if !ok {
		t.Fatal("expected valid work object action path")
	}
	if workObjectID != "wo_1" || action != "status" {
		t.Fatalf("expected parsed path values, got %q %q", workObjectID, action)
	}

	workObjectID, action, ok = workObjectActionPath("/api/work-objects/wo_1/core")
	if !ok {
		t.Fatal("expected valid work object core action path")
	}
	if workObjectID != "wo_1" || action != "core" {
		t.Fatalf("expected parsed core path values, got %q %q", workObjectID, action)
	}

	workObjectID, action, ok = workObjectActionPath("/api/work-objects/wo_1/priority")
	if !ok {
		t.Fatal("expected valid work object priority action path")
	}
	if workObjectID != "wo_1" || action != "priority" {
		t.Fatalf("expected parsed priority path values, got %q %q", workObjectID, action)
	}

	workObjectID, action, ok = workObjectActionPath("/api/work-objects/wo_1/decisions")
	if !ok {
		t.Fatal("expected valid work object decision action path")
	}
	if workObjectID != "wo_1" || action != "decisions" {
		t.Fatalf("expected parsed decision path values, got %q %q", workObjectID, action)
	}

	workObjectID, action, ok = workObjectActionPath("/api/work-objects/wo_1/artifacts")
	if !ok {
		t.Fatal("expected valid work object artifact action path")
	}
	if workObjectID != "wo_1" || action != "artifacts" {
		t.Fatalf("expected parsed artifact path values, got %q %q", workObjectID, action)
	}

	workObjectID, action, ok = workObjectActionPath("/api/work-objects/wo_1/plan")
	if !ok {
		t.Fatal("expected valid work object draft plan path")
	}
	if workObjectID != "wo_1" || action != "plan" {
		t.Fatalf("expected parsed draft plan path values, got %q %q", workObjectID, action)
	}

	workObjectID, action, ok = workObjectActionPath("/api/work-objects/wo_1/implementation-artifact")
	if !ok {
		t.Fatal("expected valid work object implementation artifact path")
	}
	if workObjectID != "wo_1" || action != "implementation-artifact" {
		t.Fatalf("expected parsed implementation artifact path values, got %q %q", workObjectID, action)
	}

	if _, _, ok := workObjectActionPath("/api/work-objects/wo_1"); ok {
		t.Fatal("expected invalid work object action path")
	}
}
