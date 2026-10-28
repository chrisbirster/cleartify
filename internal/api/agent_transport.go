package api

import (
	"github.com/chrisbirster/cleartify/internal/interop"
	"net/http"
)

func (s *Server) submitLocalAgentMessage(w http.ResponseWriter, r *http.Request) {
	owner, ok := s.agentOwnerSession(w, r)
	if !ok {
		return
	}
	var envelope interop.Envelope
	if !decodeJSON(w, r, &envelope) {
		return
	}
	caller := interop.VerifiedLocalCaller{OwnerDID: owner, AgentID: r.PathValue("id"), Verified: true}
	if err := s.agentWorker.Submit(r.Context(), caller, envelope); err != nil {
		writeJSON(w, 403, map[string]string{"error": "local_message_denied"})
		return
	}
	writeJSON(w, 202, map[string]any{"status": "pending", "localOnly": true})
}
func (s *Server) tickLocalAgentWorker(w http.ResponseWriter, r *http.Request) {
	if s.adminToken == "" || !secureBearerMatch(r.Header.Get("Authorization"), s.adminToken) {
		writeJSON(w, 401, map[string]string{"error": "unauthorized"})
		return
	}
	if s.appOrigin == "" || r.Header.Get("Origin") != s.appOrigin {
		writeJSON(w, 403, map[string]string{"error": "invalid_origin"})
		return
	}
	worked, err := s.agentWorker.Tick(r.Context())
	if err != nil {
		writeJSON(w, 503, map[string]string{"error": "local_worker_unavailable"})
		return
	}
	writeJSON(w, 200, map[string]any{"worked": worked, "localOnly": true})
}
