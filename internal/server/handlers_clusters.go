package server

import (
	"net/http"

	"cleartify/internal/httpx"
)

func (s *Server) handleClusters(w http.ResponseWriter, r *http.Request) {
	session, ok := s.requireAuthenticatedSession(w, r)
	if !ok {
		return
	}

	clusters, err := s.store.ListClusters(r.Context())
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "clusters_failed", "Unable to load Clusters.")
		return
	}

	projectID := r.URL.Query().Get("projectId")
	visibleClusters := filterClustersForWorkspace(session, clusters)
	httpx.WriteJSON(w, http.StatusOK, filterClustersForProject(projectID, visibleClusters))
}
