package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"cleartify/internal/config"
	"cleartify/internal/domain"
)

func TestHandleProjectsReturnsWorkspaceVisibleProjects(t *testing.T) {
	t.Parallel()

	server, err := New(
		config.Config{AppMarketingURL: "http://localhost:4173"},
		&fakeStore{
			projects: []domain.Project{
				{ID: "prj_live_1", WorkspaceID: "wrk_live", Name: "Live project"},
				{ID: "prj_other_1", WorkspaceID: "wrk_other", Name: "Other project"},
				{ID: "prj_sample_cleartify", WorkspaceID: "workspace_sample", Name: "Sample project"},
			},
		},
		&fakeAuthService{
			session: domain.AuthSession{
				Authenticated: true,
				User:          &domain.User{ID: "usr_1"},
				Workspace:     &domain.Workspace{ID: "wrk_live", Mode: "live"},
			},
		},
		newTestServer(t).frontend,
	)
	if err != nil {
		t.Fatalf("expected server build: %v", err)
	}

	recorder := httptest.NewRecorder()
	server.handleProjects(recorder, httptest.NewRequest(http.MethodGet, "/api/projects", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), `"id":"prj_live_1"`) {
		t.Fatalf("expected live project in response, got %q", recorder.Body.String())
	}
	if strings.Contains(recorder.Body.String(), `"id":"prj_other_1"`) {
		t.Fatalf("did not expect other workspace project in response, got %q", recorder.Body.String())
	}
	if strings.Contains(recorder.Body.String(), `"id":"prj_sample_cleartify"`) {
		t.Fatalf("did not expect sample project in live workspace response, got %q", recorder.Body.String())
	}
}

func TestHandleProjectsCreateUsesCurrentWorkspace(t *testing.T) {
	t.Parallel()

	server, err := New(
		config.Config{AppMarketingURL: "http://localhost:4173"},
		&fakeStore{},
		&fakeAuthService{
			session: domain.AuthSession{
				Authenticated: true,
				User:          &domain.User{ID: "usr_1"},
				Workspace:     &domain.Workspace{ID: "wrk_live", Mode: "live"},
			},
		},
		newTestServer(t).frontend,
	)
	if err != nil {
		t.Fatalf("expected server build: %v", err)
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/projects", strings.NewReader(`{"name":"Guided onboarding refresh","summary":"Create a calmer first-run project lane."}`))
	server.handleProjects(recorder, request)
	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), `"workspaceId":"wrk_live"`) {
		t.Fatalf("expected created project to use current workspace, got %q", recorder.Body.String())
	}
}

func TestHandleClustersFiltersByProjectID(t *testing.T) {
	t.Parallel()

	server, err := New(
		config.Config{AppMarketingURL: "http://localhost:4173"},
		&fakeStore{
			clusters: []domain.Cluster{
				{ID: "clu_1", ProjectID: "prj_live_1", Title: "Scoped cluster"},
				{ID: "clu_2", ProjectID: "prj_live_2", Title: "Other cluster"},
			},
		},
		&fakeAuthService{
			session: domain.AuthSession{
				Authenticated: true,
				User:          &domain.User{ID: "usr_1"},
				Workspace:     &domain.Workspace{ID: "wrk_live", Mode: "live"},
			},
		},
		newTestServer(t).frontend,
	)
	if err != nil {
		t.Fatalf("expected server build: %v", err)
	}

	recorder := httptest.NewRecorder()
	server.handleClusters(recorder, httptest.NewRequest(http.MethodGet, "/api/clusters?projectId=prj_live_1", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), `"id":"clu_1"`) {
		t.Fatalf("expected scoped cluster in response, got %q", recorder.Body.String())
	}
	if strings.Contains(recorder.Body.String(), `"id":"clu_2"`) {
		t.Fatalf("did not expect unscoped cluster in response, got %q", recorder.Body.String())
	}
}
