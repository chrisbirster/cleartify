package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"cleartify/internal/httpx"
	"cleartify/internal/store"
)

type projectRequest struct {
	Name    string `json:"name"`
	Summary string `json:"summary"`
	Status  string `json:"status"`
}

func (s *Server) handleProjects(w http.ResponseWriter, r *http.Request) {
	session, ok := s.requireAuthenticatedSession(w, r)
	if !ok {
		return
	}

	switch r.Method {
	case http.MethodGet:
		projects, err := s.store.ListProjects(r.Context())
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, "projects_failed", "Unable to load Projects.")
			return
		}
		httpx.WriteJSON(w, http.StatusOK, filterProjectsForWorkspace(session, projects))
	case http.MethodPost:
		if session.Workspace == nil {
			httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated", "Sign in to create a Project.")
			return
		}

		var request projectRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "invalid_project", "Project name and summary are required.")
			return
		}

		project, err := s.store.CreateProject(r.Context(), session.Workspace.ID, request.Name, request.Summary, request.Status)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrInvalidProjectFields):
				httpx.WriteError(w, http.StatusBadRequest, "invalid_project", "Project name and summary are required.")
			default:
				httpx.WriteError(w, http.StatusInternalServerError, "create_project_failed", "Unable to create Project.")
			}
			return
		}

		httpx.WriteJSON(w, http.StatusCreated, project)
	default:
		httpx.MethodNotAllowed(w)
	}
}

func (s *Server) handleProjectActions(w http.ResponseWriter, r *http.Request) {
	projectID, ok := entityPathID(r.URL.Path, "/api/projects/")
	if !ok {
		http.NotFound(w, r)
		return
	}

	session, ok := s.requireAuthenticatedSession(w, r)
	if !ok {
		return
	}

	project, err := s.store.LookupProject(r.Context(), projectID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			httpx.WriteError(w, http.StatusNotFound, "project_not_found", "Project not found.")
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "project_failed", "Unable to load Project.")
		}
		return
	}
	if !projectVisibleInWorkspace(session, project) {
		httpx.WriteError(w, http.StatusNotFound, "project_not_found", "Project not found.")
		return
	}

	switch r.Method {
	case http.MethodGet:
		httpx.WriteJSON(w, http.StatusOK, project)
	case http.MethodPut:
		var request projectRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "invalid_project", "Project name and summary are required.")
			return
		}

		updatedProject, err := s.store.UpdateProject(r.Context(), projectID, request.Name, request.Summary, request.Status)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrInvalidProjectFields):
				httpx.WriteError(w, http.StatusBadRequest, "invalid_project", "Project name and summary are required.")
			case errors.Is(err, sql.ErrNoRows):
				httpx.WriteError(w, http.StatusNotFound, "project_not_found", "Project not found.")
			default:
				httpx.WriteError(w, http.StatusInternalServerError, "update_project_failed", "Unable to update Project.")
			}
			return
		}

		httpx.WriteJSON(w, http.StatusOK, updatedProject)
	default:
		httpx.MethodNotAllowed(w)
	}
}
