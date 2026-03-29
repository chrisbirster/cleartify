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
	CreateWorkObjectFromCluster(ctx context.Context, clusterID string) (domain.WorkObject, error)
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
