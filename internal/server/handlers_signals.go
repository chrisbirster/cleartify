package server

import (
	"net/http"

	"cleartify/internal/httpx"
)

func (s *Server) handleSignals(w http.ResponseWriter, r *http.Request) {
	signals, err := s.store.ListSignals(r.Context())
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "signals_failed", "Unable to load Signals.")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, signals)
}
