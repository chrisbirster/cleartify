package server

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"testing/fstest"

	"cleartify/internal/auth"
	"cleartify/internal/config"
	"cleartify/internal/domain"
)

type fakeStore struct {
	signals           []domain.Signal
	clusters          []domain.Cluster
	workObjects       []domain.WorkObject
	createdWorkObject domain.WorkObject
	err               error
}

func (f *fakeStore) ListSignals(context.Context) ([]domain.Signal, error) {
	return f.signals, f.err
}

func (f *fakeStore) ListClusters(context.Context) ([]domain.Cluster, error) {
	return f.clusters, f.err
}

func (f *fakeStore) ListWorkObjects(context.Context) ([]domain.WorkObject, error) {
	return f.workObjects, f.err
}

func (f *fakeStore) CreateWorkObjectFromCluster(_ context.Context, clusterID string) (domain.WorkObject, error) {
	if f.err != nil {
		return domain.WorkObject{}, f.err
	}
	if f.createdWorkObject.ID == "" {
		return domain.WorkObject{
			ID:        "wo_created",
			ClusterID: clusterID,
			Title:     "Created Work Object",
		}, nil
	}
	return f.createdWorkObject, nil
}

type fakeAuthService struct {
	session domain.AuthSession
	err     error
}

func (f *fakeAuthService) CurrentSession(context.Context, *http.Request) (domain.AuthSession, error) {
	return f.session, f.err
}

func newTestServer(t *testing.T) *Server {
	t.Helper()

	server, err := New(
		config.Config{AppMarketingURL: "http://localhost:4173"},
		&fakeStore{},
		auth.NewDemoAuthService(),
		fstest.MapFS{
			"index.html":           {Data: []byte("<html>app</html>")},
			"assets/app.js":        {Data: []byte("console.log('app')")},
			"assets/app.css":       {Data: []byte("body{}")},
			"favicon.ico":          {Data: []byte("ico")},
			"manifest.webmanifest": {Data: []byte("{}")},
		},
	)
	if err != nil {
		t.Fatalf("expected server to build: %v", err)
	}
	return server
}

func TestNewBuildsHandler(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	if server.Handler() == nil {
		t.Fatal("expected handler to be initialized")
	}
}

func TestFakeStoreCanReturnError(t *testing.T) {
	t.Parallel()

	store := &fakeStore{err: errors.New("boom")}
	if _, err := store.ListSignals(context.Background()); err == nil {
		t.Fatal("expected fake store error")
	}
}
