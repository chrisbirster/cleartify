package server

import (
	"net/http"

	"cleartify/internal/httpx"
)

func (s *Server) handleSession(w http.ResponseWriter, r *http.Request) {
	session, err := s.authService.CurrentSession(r.Context(), r)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "session_failed", "Unable to resolve the current session.")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, session)
}
