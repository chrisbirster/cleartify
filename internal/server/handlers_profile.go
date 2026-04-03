package server

import (
	"net/http"

	"cleartify/internal/domain"
	"cleartify/internal/httpx"
)

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	session, ok := s.requireAuthenticatedSession(w, r)
	if !ok {
		return
	}

	httpx.WriteJSON(w, http.StatusOK, domain.Me{
		User:            *session.User,
		Organization:    session.Organization,
		Workspace:       session.Workspace,
		NeedsOnboarding: session.NeedsOnboarding,
	})
}

func (s *Server) handleTeam(w http.ResponseWriter, r *http.Request) {
	session, ok := s.requireAuthenticatedSession(w, r)
	if !ok {
		return
	}
	if session.Workspace == nil {
		httpx.WriteError(w, http.StatusNotFound, "team_not_found", "Current team not found.")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, domain.Team{
		Workspace:    *session.Workspace,
		Organization: session.Organization,
		CurrentUser:  session.User,
	})
}

func (s *Server) requireAuthenticatedSession(w http.ResponseWriter, r *http.Request) (domain.AuthSession, bool) {
	session, err := s.authService.CurrentSession(r.Context(), r)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "session_failed", "Unable to load the current session.")
		return domain.AuthSession{}, false
	}
	if !session.Authenticated || session.User == nil {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated", "Authentication required.")
		return domain.AuthSession{}, false
	}
	return session, true
}
