package api

import (
	"database/sql"
	"errors"
	"github.com/chrisbirster/cleartify/internal/social"
	"net/http"
)

// Social account creation never consumes a user-controlled DID, phone or
// email. It requires a valid, decryptable Indigo browser OAuth session.
func (s *Server) registerSocialAccount(w http.ResponseWriter, r *http.Request) {
	if s.socialAuth == nil || s.socialAccounts == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "social_login_not_configured"})
		return
	}
	if r.Header.Get("Origin") != s.appOrigin {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "invalid_origin"})
		return
	}
	did, ok := s.socialAuth.SocialSessionDID(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "atproto_login_required"})
		return
	}
	acc, err := s.socialAccounts.EnsureAuthenticated(r.Context(), did)
	if err != nil {
		if errors.Is(err, social.ErrSocialIdentityRequired) {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "atproto_login_required"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal_error"})
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusCreated, acc)
}

func (s *Server) setupSocialProfile(w http.ResponseWriter, r *http.Request) {
	if s.socialAuth == nil || s.socialAccounts == nil {
		writeJSON(w, 503, map[string]string{"error": "social_login_not_configured"})
		return
	}
	did, ok := s.socialAuth.SocialSessionDID(r)
	if !ok {
		writeJSON(w, 401, map[string]string{"error": "atproto_login_required"})
		return
	}
	if s.moderationBroader && !s.enforceSubjectHold(w, r, did) {
		return
	}
	var input social.ProfileInput
	if !decodeJSON(w, r, &input) {
		return
	}
	result, err := s.socialAccounts.Setup(r.Context(), did, input)
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": "profile_setup_denied"})
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, 200, result)
}

func (s *Server) getSocialProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	did, ok := s.socialAuth.SocialSessionDID(r)
	if !ok {
		writeJSON(w, 401, map[string]string{"error": "atproto_login_required"})
		return
	}
	profile, err := s.socialAccounts.Own(r.Context(), did)
	if errors.Is(err, sql.ErrNoRows) {
		writeJSON(w, 404, map[string]string{"error": "profile_not_set_up"})
		return
	}
	if err != nil {
		writeJSON(w, 503, map[string]string{"error": "profile_unavailable"})
		return
	}
	writeJSON(w, 200, profile)
}
