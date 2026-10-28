package api

import (
	"encoding/hex"
	"net/http"
	"strings"
)

func (s *Server) agentOwnerSession(w http.ResponseWriter, r *http.Request) (string, bool) {
	if s.socialAuth == nil || s.agentOwnership == nil {
		writeJSON(w, 503, map[string]string{"error": "ownership_unavailable"})
		return "", false
	}
	if s.appOrigin == "" || r.Header.Get("Origin") != s.appOrigin {
		writeJSON(w, 403, map[string]string{"error": "invalid_origin"})
		return "", false
	}
	did, ok := s.socialAuth.SocialSessionDID(r)
	if !ok {
		writeJSON(w, 401, map[string]string{"error": "social_auth_required"})
		return "", false
	}
	return did, true
}
func (s *Server) agentOwnershipStart(w http.ResponseWriter, r *http.Request) {
	did, ok := s.agentOwnerSession(w, r)
	if !ok {
		return
	}
	var input struct {
		PublicKey string `json:"publicKey"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	if len(input.PublicKey) != 64 {
		writeJSON(w, 400, map[string]string{"error": "invalid_key"})
		return
	}
	key, err := hex.DecodeString(strings.ToLower(input.PublicKey))
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid_key"})
		return
	}
	nonce, err := s.agentOwnership.Start(r.Context(), did, r.PathValue("id"), key)
	if err != nil {
		writeJSON(w, 403, map[string]string{"error": "ownership_denied"})
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, 201, map[string]any{"challenge": nonce, "expiresInSeconds": 300, "proofAlgorithm": "Ed25519"})
}
func (s *Server) agentOwnershipConfirm(w http.ResponseWriter, r *http.Request) {
	did, ok := s.agentOwnerSession(w, r)
	if !ok {
		return
	}
	var input struct {
		Challenge string `json:"challenge"`
		Proof     string `json:"proof"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	if err := s.agentOwnership.Confirm(r.Context(), did, r.PathValue("id"), input.Challenge, input.Proof); err != nil {
		writeJSON(w, 403, map[string]string{"error": "ownership_denied"})
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	var status string
	if err := s.agentOwnership.DB.QueryRowContext(r.Context(), "SELECT status FROM agent_registrations WHERE id=?", r.PathValue("id")).Scan(&status); err != nil {
		writeJSON(w, 503, map[string]string{"error": "ownership_status_unavailable"})
		return
	}
	writeJSON(w, 200, map[string]any{"ownershipVerified": true, "status": status, "activationRequired": status != "verified", "fakeVerifierOnly": false})
}

func (s *Server) activateAgent(w http.ResponseWriter, r *http.Request) {
	if s.adminToken == "" || !secureBearerMatch(r.Header.Get("Authorization"), s.adminToken) {
		writeJSON(w, 401, map[string]string{"error": "unauthorized"})
		return
	}
	if s.appOrigin == "" || r.Header.Get("Origin") != s.appOrigin {
		writeJSON(w, 403, map[string]string{"error": "invalid_origin"})
		return
	}
	if err := s.agentOwnership.Promote(r.Context(), r.PathValue("id")); err != nil {
		writeJSON(w, 409, map[string]string{"error": "agent_activation_denied"})
		return
	}
	writeJSON(w, 200, map[string]string{"status": "verified"})
}
