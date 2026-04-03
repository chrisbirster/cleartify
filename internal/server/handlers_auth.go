package server

import (
	"encoding/json"
	"errors"
	"net/http"

	"cleartify/internal/auth"
	"cleartify/internal/httpx"
)

type workspaceModeRequest struct {
	Mode string `json:"mode"`
}

func (s *Server) handleSession(w http.ResponseWriter, r *http.Request) {
	session, err := s.authService.CurrentSession(r.Context(), r)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "session_failed", "Unable to resolve the current session.")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, session)
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	redirectTo, err := s.authService.LoginURL(r.Context(), r)
	if err != nil && redirectTo == "" {
		redirectTo = authErrorURL(s.config.AppPublicURL, "auth_unavailable")
	}
	http.Redirect(w, r, redirectTo, http.StatusFound)
}

func (s *Server) handleAuthCallback(w http.ResponseWriter, r *http.Request) {
	result, err := s.authService.Authenticate(r.Context(), r)
	if err != nil {
		http.Redirect(w, r, authErrorURL(s.config.AppPublicURL, "auth_failed"), http.StatusFound)
		return
	}

	http.SetCookie(w, s.sessionCookie(result.SessionID))
	http.Redirect(w, r, result.ReturnTo, http.StatusFound)
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	redirectTo, err := s.authService.LogoutURL(r.Context(), r)
	if err != nil || redirectTo == "" {
		redirectTo = s.config.AppMarketingURL
	}

	http.SetCookie(w, s.expiredSessionCookie())
	http.Redirect(w, r, redirectTo, http.StatusFound)
}

func (s *Server) handleCompleteOnboarding(w http.ResponseWriter, r *http.Request) {
	session, err := s.authService.CompleteOnboarding(r.Context(), r)
	if err != nil {
		if errors.Is(err, auth.ErrUnauthenticated) {
			httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated", "Sign in to complete onboarding.")
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "onboarding_complete_failed", "Unable to complete onboarding.")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, session)
}

func (s *Server) handleWorkspaceMode(w http.ResponseWriter, r *http.Request) {
	var request workspaceModeRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_workspace_mode", "Select either sample or live workspace mode.")
		return
	}

	session, err := s.authService.SelectWorkspaceMode(r.Context(), r, request.Mode)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrUnauthenticated):
			httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated", "Sign in to choose a workspace mode.")
		case errors.Is(err, auth.ErrInvalidWorkspaceMode):
			httpx.WriteError(w, http.StatusBadRequest, "invalid_workspace_mode", "Select either sample or live workspace mode.")
		case errors.Is(err, auth.ErrAuthUnavailable):
			httpx.WriteError(w, http.StatusServiceUnavailable, "workspace_mode_unavailable", "Workspace mode selection is unavailable right now.")
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "workspace_mode_failed", "Unable to save workspace mode.")
		}
		return
	}

	httpx.WriteJSON(w, http.StatusOK, session)
}
