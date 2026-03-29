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
)

func TestHandleClusterActionsCreatesWorkObject(t *testing.T) {
	t.Parallel()

	server, err := New(
		config.Config{AppMarketingURL: "http://localhost:4173"},
		&fakeStore{createdWorkObject: domain.WorkObject{ID: "wo_created", ClusterID: "clu_1", Title: "Created Work Object"}},
		auth.NewDemoAuthService(),
		newTestServer(t).frontend,
	)
	if err != nil {
		t.Fatalf("expected server build: %v", err)
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/clusters/clu_1/create-work-object", nil)
	server.handleClusterActions(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), `"id":"wo_created"`) {
		t.Fatalf("expected created work object in response, got %q", recorder.Body.String())
	}
}

func TestHandleClusterActionsReturnsNotFoundForMissingCluster(t *testing.T) {
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
	request := httptest.NewRequest(http.MethodPost, "/api/clusters/missing/create-work-object", nil)
	server.handleClusterActions(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", recorder.Code)
	}
}

func TestHandleClusterActionsReturnsInternalErrorForCreateFailure(t *testing.T) {
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
	request := httptest.NewRequest(http.MethodPost, "/api/clusters/clu_1/create-work-object", nil)
	server.handleClusterActions(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", recorder.Code)
	}
}

func TestHandleClusterActionsRejectsUnknownPaths(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/clusters/clu_1/not-real", nil)
	server.handleClusterActions(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown cluster action, got %d", recorder.Code)
	}
}

func TestClusterActionPath(t *testing.T) {
	t.Parallel()

	clusterID, action, ok := clusterActionPath("/api/clusters/clu_1/create-work-object")
	if !ok {
		t.Fatal("expected valid cluster action path")
	}
	if clusterID != "clu_1" || action != "create-work-object" {
		t.Fatalf("expected parsed path values, got %q %q", clusterID, action)
	}

	if _, _, ok := clusterActionPath("/api/clusters/clu_1"); ok {
		t.Fatal("expected invalid cluster action path")
	}
}
