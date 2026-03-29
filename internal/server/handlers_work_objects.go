package server

import (
	"net/http"

	"cleartify/internal/httpx"
)

func (s *Server) handleWorkObjects(w http.ResponseWriter, r *http.Request) {
	workObjects, err := s.store.ListWorkObjects(r.Context())
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "work_objects_failed", "Unable to load Work Objects.")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, workObjects)
}
