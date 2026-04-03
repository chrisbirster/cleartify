package server

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"cleartify/internal/httpx"
)

func (s *Server) handleClusterActions(w http.ResponseWriter, r *http.Request) {
	clusterID, action, ok := clusterActionPath(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}

	switch action {
	case "create-work-object":
		if r.Method != http.MethodPost {
			httpx.MethodNotAllowed(w)
			return
		}
		if _, ok := requireVisibleClusterSession(s, w, r, clusterID); !ok {
			return
		}

		workObject, err := s.store.CreateWorkObjectFromCluster(r.Context(), clusterID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				httpx.WriteError(w, http.StatusNotFound, "cluster_not_found", "Cluster not found.")
				return
			}
			httpx.WriteError(w, http.StatusInternalServerError, "create_work_object_failed", "Unable to create Work Object.")
			return
		}

		httpx.WriteJSON(w, http.StatusCreated, workObject)
	case "resolve":
		if r.Method != http.MethodPost {
			httpx.MethodNotAllowed(w)
			return
		}
		if _, ok := requireVisibleClusterSession(s, w, r, clusterID); !ok {
			return
		}

		cluster, err := s.store.ResolveCluster(r.Context(), clusterID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				httpx.WriteError(w, http.StatusNotFound, "cluster_not_found", "Cluster not found.")
				return
			}
			httpx.WriteError(w, http.StatusInternalServerError, "resolve_cluster_failed", "Unable to resolve Cluster.")
			return
		}

		httpx.WriteJSON(w, http.StatusOK, cluster)
	default:
		http.NotFound(w, r)
	}
}

func clusterActionPath(path string) (clusterID string, action string, ok bool) {
	trimmed := strings.Trim(strings.TrimPrefix(path, "/api/clusters/"), "/")
	parts := strings.Split(trimmed, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", false
	}

	return parts[0], parts[1], true
}
