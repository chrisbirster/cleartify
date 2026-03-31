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
	signals            []domain.Signal
	clusters           []domain.Cluster
	workObjects        []domain.WorkObject
	lookupDraftPlan    *domain.DraftPlan
	generatedDraftPlan domain.DraftPlan
	updatedDraftPlan   domain.DraftPlan
	generatedArtifact  domain.Artifact
	createdWorkObject  domain.WorkObject
	createdDecision    domain.Decision
	createdArtifact    domain.Artifact
	updatedWorkObject  domain.WorkObject
	err                error
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

func (f *fakeStore) LookupDraftPlanForWorkObject(_ context.Context, workObjectID string) (*domain.DraftPlan, error) {
	if f.err != nil {
		return nil, f.err
	}
	if f.lookupDraftPlan == nil {
		return nil, nil
	}
	if f.lookupDraftPlan.WorkObjectID == "" {
		return &domain.DraftPlan{
			WorkObjectID:        workObjectID,
			Overview:            "Persisted plan overview",
			ImplementationSteps: []string{"Step 1"},
			EdgeCases:           []string{"Edge case 1"},
			OpenQuestions:       []string{"Open question 1"},
		}, nil
	}
	return f.lookupDraftPlan, nil
}

func (f *fakeStore) GenerateDraftPlanForWorkObject(_ context.Context, workObjectID string) (domain.DraftPlan, error) {
	if f.err != nil {
		return domain.DraftPlan{}, f.err
	}
	if f.generatedDraftPlan.WorkObjectID == "" {
		return domain.DraftPlan{
			WorkObjectID:        workObjectID,
			Overview:            "Generated plan overview",
			ImplementationSteps: []string{"Step 1"},
			EdgeCases:           []string{"Edge case 1"},
			OpenQuestions:       []string{"Open question 1"},
		}, nil
	}
	return f.generatedDraftPlan, nil
}

func (f *fakeStore) UpdateDraftPlanForWorkObject(
	_ context.Context,
	workObjectID string,
	overview string,
	implementationSteps []string,
	edgeCases []string,
	openQuestions []string,
) (domain.DraftPlan, error) {
	if f.err != nil {
		return domain.DraftPlan{}, f.err
	}
	if f.updatedDraftPlan.WorkObjectID == "" {
		return domain.DraftPlan{
			WorkObjectID:        workObjectID,
			Overview:            overview,
			ImplementationSteps: implementationSteps,
			EdgeCases:           edgeCases,
			OpenQuestions:       openQuestions,
		}, nil
	}
	return f.updatedDraftPlan, nil
}

func (f *fakeStore) GenerateImplementationArtifactForWorkObject(_ context.Context, workObjectID string) (domain.Artifact, error) {
	if f.err != nil {
		return domain.Artifact{}, f.err
	}
	if f.generatedArtifact.ID == "" {
		return domain.Artifact{
			ID:           "art_generated",
			WorkObjectID: workObjectID,
			ArtifactType: "implementation_brief",
			Title:        "Generated implementation brief",
			Body:         "Deterministic mock artifact body.",
		}, nil
	}
	return f.generatedArtifact, nil
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

func (f *fakeStore) CreateDecisionForWorkObject(_ context.Context, workObjectID string, title string, summary string, rationale string) (domain.Decision, error) {
	if f.err != nil {
		return domain.Decision{}, f.err
	}
	if f.createdDecision.ID == "" {
		return domain.Decision{
			ID:           "dec_created",
			WorkObjectID: workObjectID,
			Title:        title,
			Summary:      summary,
			Rationale:    rationale,
		}, nil
	}
	return f.createdDecision, nil
}

func (f *fakeStore) CreateArtifactForWorkObject(_ context.Context, workObjectID string, artifactType string, title string, body string) (domain.Artifact, error) {
	if f.err != nil {
		return domain.Artifact{}, f.err
	}
	if f.createdArtifact.ID == "" {
		return domain.Artifact{
			ID:           "art_created",
			WorkObjectID: workObjectID,
			ArtifactType: artifactType,
			Title:        title,
			Body:         body,
		}, nil
	}
	return f.createdArtifact, nil
}

func (f *fakeStore) UpdateWorkObjectStatus(_ context.Context, workObjectID string, status string) (domain.WorkObject, error) {
	if f.err != nil {
		return domain.WorkObject{}, f.err
	}
	if f.updatedWorkObject.ID == "" {
		return domain.WorkObject{
			ID:     workObjectID,
			Title:  "Updated Work Object",
			Status: status,
		}, nil
	}
	return f.updatedWorkObject, nil
}

func (f *fakeStore) UpdateWorkObjectCoreFields(_ context.Context, workObjectID string, title string, summary string, goal string) (domain.WorkObject, error) {
	if f.err != nil {
		return domain.WorkObject{}, f.err
	}
	if f.updatedWorkObject.ID == "" {
		return domain.WorkObject{
			ID:      workObjectID,
			Title:   title,
			Summary: summary,
			Goal:    goal,
		}, nil
	}
	return f.updatedWorkObject, nil
}

func (f *fakeStore) UpdateWorkObjectPrototype(_ context.Context, workObjectID string, prototypeURL string, prototypeNote string) (domain.WorkObject, error) {
	if f.err != nil {
		return domain.WorkObject{}, f.err
	}
	if f.updatedWorkObject.ID == "" {
		return domain.WorkObject{
			ID:            workObjectID,
			Title:         "Updated Work Object",
			PrototypeURL:  prototypeURL,
			PrototypeNote: prototypeNote,
		}, nil
	}
	return f.updatedWorkObject, nil
}

func (f *fakeStore) UpdateWorkObjectPriority(_ context.Context, workObjectID string, priority string) (domain.WorkObject, error) {
	if f.err != nil {
		return domain.WorkObject{}, f.err
	}
	if f.updatedWorkObject.ID == "" {
		return domain.WorkObject{
			ID:       workObjectID,
			Title:    "Updated Work Object",
			Priority: priority,
		}, nil
	}
	return f.updatedWorkObject, nil
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
