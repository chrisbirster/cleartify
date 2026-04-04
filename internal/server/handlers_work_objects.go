package server

import (
	"net/http"

	"cleartify/internal/httpx"
)

func (s *Server) handleWorkObjects(w http.ResponseWriter, r *http.Request) {
	session, ok := s.requireAuthenticatedSession(w, r)
	if !ok {
		return
	}

	workObjects, err := s.store.ListWorkObjects(r.Context())
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "work_objects_failed", "Unable to load Work Objects.")
		return
	}

	projectID := r.URL.Query().Get("projectId")
	visibleWorkObjects := filterWorkObjectsForWorkspace(session, workObjects)
	httpx.WriteJSON(w, http.StatusOK, filterWorkObjectsForProject(projectID, visibleWorkObjects))
}
