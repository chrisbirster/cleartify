package server

import (
	"net/http"

	"cleartify/internal/httpx"
)

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	httpx.WriteJSON(w, 200, map[string]string{"status": "ok"})
}
