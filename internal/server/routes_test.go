package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"cleartify/internal/auth"
	"cleartify/internal/config"
	"cleartify/internal/domain"
)

func TestRoutesRegisterKnownAPIEndpoints(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)

	for _, path := range []string{
		"/api/health",
		"/api/auth/session",
		"/api/me/onboarding/complete",
		"/api/signals",
		"/api/clusters",
		"/api/work-objects",
		"/login",
		"/auth/callback",
		"/logout",
		"/api/work-objects/wo_1/core",
		"/api/work-objects/wo_1/prototype",
		"/api/work-objects/wo_1/status",
		"/api/work-objects/wo_1/priority",
		"/api/work-objects/wo_1/release-ready",
		"/api/work-objects/wo_1/decisions",
		"/api/work-objects/wo_1/artifacts",
		"/api/work-objects/wo_1/implementation-artifact",
		"/api/work-objects/wo_1/task-breakdown-artifact",
	} {
		method := http.MethodGet
		if path == "/api/me/onboarding/complete" || path == "/api/work-objects/wo_1/core" || path == "/api/work-objects/wo_1/prototype" || path == "/api/work-objects/wo_1/status" || path == "/api/work-objects/wo_1/priority" || path == "/api/work-objects/wo_1/release-ready" || path == "/api/work-objects/wo_1/decisions" || path == "/api/work-objects/wo_1/artifacts" || path == "/api/work-objects/wo_1/implementation-artifact" || path == "/api/work-objects/wo_1/task-breakdown-artifact" {
			method = http.MethodPost
		}
		request := httptest.NewRequest(method, path, nil)
		recorder := httptest.NewRecorder()
		server.ServeHTTP(recorder, request)
		if recorder.Code == http.StatusNotFound {
			t.Fatalf("expected route %s to be registered", path)
		}
	}
}

func TestRoutesRegisterWorkObjectPlanEndpoints(t *testing.T) {
	t.Parallel()

	server, err := New(
		config.Config{AppMarketingURL: "http://localhost:4173"},
		&fakeStore{
			lookupDraftPlan: &domain.DraftPlan{
				WorkObjectID:        "wo_1",
				Overview:            "Persisted plan overview",
				ImplementationSteps: []string{"Step 1"},
				EdgeCases:           []string{"Edge case 1"},
				OpenQuestions:       []string{"Question 1"},
			},
			generatedDraftPlan: domain.DraftPlan{
				WorkObjectID:        "wo_1",
				Overview:            "Generated plan overview",
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

	for _, method := range []string{http.MethodGet, http.MethodPost, http.MethodPut} {
		request := httptest.NewRequest(method, "/api/work-objects/wo_1/plan", nil)
		recorder := httptest.NewRecorder()
		server.ServeHTTP(recorder, request)
		if recorder.Code == http.StatusNotFound {
			t.Fatalf("expected work object plan route to be registered for %s", method)
		}
	}
}

func TestRoutesRegisterClusterWorkObjectCreateEndpoint(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)

	request := httptest.NewRequest(http.MethodPost, "/api/clusters/clu_1/create-work-object", nil)
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, request)
	if recorder.Code == http.StatusNotFound {
		t.Fatal("expected cluster create-work-object route to be registered")
	}
}

func TestClusterWorkObjectCreateRejectsWrongMethod(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)

	request := httptest.NewRequest(http.MethodGet, "/api/clusters/clu_1/create-work-object", nil)
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405 for wrong method, got %d", recorder.Code)
	}
}

func TestWorkObjectStatusUpdateRejectsWrongMethod(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)

	request := httptest.NewRequest(http.MethodGet, "/api/work-objects/wo_1/status", nil)
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405 for wrong method, got %d", recorder.Code)
	}
}

func TestWorkObjectPrototypeUpdateRejectsWrongMethod(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)

	request := httptest.NewRequest(http.MethodGet, "/api/work-objects/wo_1/prototype", nil)
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405 for wrong method, got %d", recorder.Code)
	}
}

func TestWorkObjectCoreFieldUpdateRejectsWrongMethod(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)

	request := httptest.NewRequest(http.MethodGet, "/api/work-objects/wo_1/core", nil)
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405 for wrong method, got %d", recorder.Code)
	}
}

func TestWorkObjectPriorityUpdateRejectsWrongMethod(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)

	request := httptest.NewRequest(http.MethodGet, "/api/work-objects/wo_1/priority", nil)
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405 for wrong method, got %d", recorder.Code)
	}
}

func TestWorkObjectReleaseReadyRejectsWrongMethod(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)

	request := httptest.NewRequest(http.MethodGet, "/api/work-objects/wo_1/release-ready", nil)
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405 for wrong method, got %d", recorder.Code)
	}
}

func TestWorkObjectDecisionCreateRejectsWrongMethod(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)

	request := httptest.NewRequest(http.MethodGet, "/api/work-objects/wo_1/decisions", nil)
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405 for wrong method, got %d", recorder.Code)
	}
}

func TestWorkObjectArtifactCreateRejectsWrongMethod(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)

	request := httptest.NewRequest(http.MethodGet, "/api/work-objects/wo_1/artifacts", nil)
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405 for wrong method, got %d", recorder.Code)
	}
}

func TestWorkObjectImplementationArtifactRejectsWrongMethod(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)

	request := httptest.NewRequest(http.MethodGet, "/api/work-objects/wo_1/implementation-artifact", nil)
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405 for wrong method, got %d", recorder.Code)
	}
}

func TestWorkObjectTaskBreakdownArtifactRejectsWrongMethod(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)

	request := httptest.NewRequest(http.MethodGet, "/api/work-objects/wo_1/task-breakdown-artifact", nil)
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405 for wrong method, got %d", recorder.Code)
	}
}

func TestWorkObjectPlanRejectsUnsupportedMethod(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)

	request := httptest.NewRequest(http.MethodDelete, "/api/work-objects/wo_1/plan", nil)
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405 for wrong method, got %d", recorder.Code)
	}
}
