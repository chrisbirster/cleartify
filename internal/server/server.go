package server

import (
	"context"
	"io/fs"
	"net/http"

	"cleartify/internal/auth"
	"cleartify/internal/config"
	"cleartify/internal/domain"
)

type DataStore interface {
	ListSignals(ctx context.Context) ([]domain.Signal, error)
	ListClusters(ctx context.Context) ([]domain.Cluster, error)
	ListWorkObjects(ctx context.Context) ([]domain.WorkObject, error)
	LookupDraftPlanForWorkObject(ctx context.Context, workObjectID string) (*domain.DraftPlan, error)
	GenerateDraftPlanForWorkObject(ctx context.Context, workObjectID string) (domain.DraftPlan, error)
	UpdateDraftPlanForWorkObject(ctx context.Context, workObjectID string, overview string, implementationSteps []string, edgeCases []string, openQuestions []string) (domain.DraftPlan, error)
	GenerateImplementationArtifactForWorkObject(ctx context.Context, workObjectID string) (domain.Artifact, error)
	GenerateTaskBreakdownArtifactForWorkObject(ctx context.Context, workObjectID string) (domain.Artifact, error)
	CreateWorkObjectFromCluster(ctx context.Context, clusterID string) (domain.WorkObject, error)
	CreateDecisionForWorkObject(ctx context.Context, workObjectID string, title string, summary string, rationale string) (domain.Decision, error)
	CreateArtifactForWorkObject(ctx context.Context, workObjectID string, artifactType string, title string, body string) (domain.Artifact, error)
	UpdateWorkObjectCoreFields(ctx context.Context, workObjectID string, title string, summary string, goal string) (domain.WorkObject, error)
	UpdateWorkObjectPrototype(ctx context.Context, workObjectID string, prototypeURL string, prototypeNote string) (domain.WorkObject, error)
	UpdateWorkObjectStatus(ctx context.Context, workObjectID string, status string) (domain.WorkObject, error)
	UpdateWorkObjectPriority(ctx context.Context, workObjectID string, priority string) (domain.WorkObject, error)
	MarkWorkObjectReadyForRelease(ctx context.Context, workObjectID string) (domain.WorkObject, error)
}

type Server struct {
	config       config.Config
	store        DataStore
	authService  auth.Service
	frontend     fs.FS
	mux          *http.ServeMux
	staticAssets http.Handler
}

func New(cfg config.Config, store DataStore, authService auth.Service, frontend fs.FS) (*Server, error) {
	server := &Server{
		config:      cfg,
		store:       store,
		authService: authService,
		frontend:    frontend,
		mux:         http.NewServeMux(),
	}
	server.staticAssets = http.StripPrefix("/app/", http.FileServer(http.FS(frontend)))
	server.registerRoutes()
	return server, nil
}

func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}
