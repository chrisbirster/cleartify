package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"cleartify/internal/httpx"
	"cleartify/internal/store"
)

type signalRequest struct {
	SourceType string `json:"sourceType"`
	SourceID   string `json:"sourceId"`
	Title      string `json:"title"`
	Body       string `json:"body"`
	URL        string `json:"url"`
}

type decisionRequest struct {
	WorkObjectID string `json:"workObjectId"`
	Title        string `json:"title"`
	Summary      string `json:"summary"`
	Rationale    string `json:"rationale"`
}

type artifactRequest struct {
	WorkObjectID string `json:"workObjectId"`
	Type         string `json:"type"`
	Title        string `json:"title"`
	Body         string `json:"body"`
	LinkType     string `json:"linkType"`
	URL          string `json:"url"`
}

func (s *Server) handleSignals(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		signals, err := s.store.ListSignals(r.Context())
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, "signals_failed", "Unable to load Signals.")
			return
		}
		httpx.WriteJSON(w, http.StatusOK, signals)
	case http.MethodPost:
		var request signalRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "invalid_signal", "Signal source, id, title, and body are required.")
			return
		}
		signal, err := s.store.CreateSignal(r.Context(), request.SourceType, request.SourceID, request.Title, request.Body, request.URL)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrInvalidSignalFields):
				httpx.WriteError(w, http.StatusBadRequest, "invalid_signal", "Signal source, id, title, and body are required.")
			default:
				httpx.WriteError(w, http.StatusInternalServerError, "create_signal_failed", "Unable to create Signal.")
			}
			return
		}
		httpx.WriteJSON(w, http.StatusCreated, signal)
	default:
		httpx.MethodNotAllowed(w)
	}
}

func (s *Server) handleSignalActions(w http.ResponseWriter, r *http.Request) {
	signalID, ok := entityPathID(r.URL.Path, "/api/signals/")
	if !ok {
		http.NotFound(w, r)
		return
	}

	switch r.Method {
	case http.MethodGet:
		signal, err := s.store.LookupSignal(r.Context(), signalID)
		if err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
				httpx.WriteError(w, http.StatusNotFound, "signal_not_found", "Signal not found.")
			default:
				httpx.WriteError(w, http.StatusInternalServerError, "signal_failed", "Unable to load Signal.")
			}
			return
		}
		httpx.WriteJSON(w, http.StatusOK, signal)
	case http.MethodPut:
		var request signalRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "invalid_signal", "Signal source, id, title, and body are required.")
			return
		}
		signal, err := s.store.UpdateSignal(r.Context(), signalID, request.SourceType, request.SourceID, request.Title, request.Body, request.URL)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrInvalidSignalFields):
				httpx.WriteError(w, http.StatusBadRequest, "invalid_signal", "Signal source, id, title, and body are required.")
			case errors.Is(err, sql.ErrNoRows):
				httpx.WriteError(w, http.StatusNotFound, "signal_not_found", "Signal not found.")
			default:
				httpx.WriteError(w, http.StatusInternalServerError, "update_signal_failed", "Unable to update Signal.")
			}
			return
		}
		httpx.WriteJSON(w, http.StatusOK, signal)
	case http.MethodDelete:
		if err := s.store.DeleteSignal(r.Context(), signalID); err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
				httpx.WriteError(w, http.StatusNotFound, "signal_not_found", "Signal not found.")
			default:
				httpx.WriteError(w, http.StatusInternalServerError, "delete_signal_failed", "Unable to delete Signal.")
			}
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		httpx.MethodNotAllowed(w)
	}
}

func (s *Server) handleDecisions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		decisions, err := s.store.ListDecisions(r.Context())
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, "decisions_failed", "Unable to load Decisions.")
			return
		}
		httpx.WriteJSON(w, http.StatusOK, decisions)
	case http.MethodPost:
		var request decisionRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "invalid_decision", "Decision work object, title, summary, and rationale are required.")
			return
		}
		decision, err := s.store.CreateDecision(r.Context(), request.WorkObjectID, request.Title, request.Summary, request.Rationale)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrInvalidDecisionFields):
				httpx.WriteError(w, http.StatusBadRequest, "invalid_decision", "Decision work object, title, summary, and rationale are required.")
			case errors.Is(err, sql.ErrNoRows):
				httpx.WriteError(w, http.StatusNotFound, "work_object_not_found", "Work Object not found.")
			default:
				httpx.WriteError(w, http.StatusInternalServerError, "create_decision_failed", "Unable to create Decision.")
			}
			return
		}
		httpx.WriteJSON(w, http.StatusCreated, decision)
	default:
		httpx.MethodNotAllowed(w)
	}
}

func (s *Server) handleDecisionActions(w http.ResponseWriter, r *http.Request) {
	decisionID, ok := entityPathID(r.URL.Path, "/api/decisions/")
	if !ok {
		http.NotFound(w, r)
		return
	}

	switch r.Method {
	case http.MethodGet:
		decision, err := s.store.LookupDecision(r.Context(), decisionID)
		if err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
				httpx.WriteError(w, http.StatusNotFound, "decision_not_found", "Decision not found.")
			default:
				httpx.WriteError(w, http.StatusInternalServerError, "decision_failed", "Unable to load Decision.")
			}
			return
		}
		httpx.WriteJSON(w, http.StatusOK, decision)
	case http.MethodPut:
		var request decisionRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "invalid_decision", "Decision work object, title, summary, and rationale are required.")
			return
		}
		decision, err := s.store.UpdateDecision(r.Context(), decisionID, request.WorkObjectID, request.Title, request.Summary, request.Rationale)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrInvalidDecisionFields):
				httpx.WriteError(w, http.StatusBadRequest, "invalid_decision", "Decision work object, title, summary, and rationale are required.")
			case errors.Is(err, sql.ErrNoRows):
				httpx.WriteError(w, http.StatusNotFound, "decision_not_found", "Decision not found.")
			default:
				httpx.WriteError(w, http.StatusInternalServerError, "update_decision_failed", "Unable to update Decision.")
			}
			return
		}
		httpx.WriteJSON(w, http.StatusOK, decision)
	case http.MethodDelete:
		if err := s.store.DeleteDecision(r.Context(), decisionID); err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
				httpx.WriteError(w, http.StatusNotFound, "decision_not_found", "Decision not found.")
			default:
				httpx.WriteError(w, http.StatusInternalServerError, "delete_decision_failed", "Unable to delete Decision.")
			}
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		httpx.MethodNotAllowed(w)
	}
}

func (s *Server) handleArtifacts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		artifacts, err := s.store.ListArtifacts(r.Context())
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, "artifacts_failed", "Unable to load Artifacts.")
			return
		}
		httpx.WriteJSON(w, http.StatusOK, artifacts)
	case http.MethodPost:
		var request artifactRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "invalid_artifact", "Artifact work object, type, title, and body are required.")
			return
		}
		artifact, err := s.store.CreateArtifact(r.Context(), request.WorkObjectID, request.Type, request.Title, request.Body, request.LinkType, request.URL)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrInvalidArtifactFields):
				httpx.WriteError(w, http.StatusBadRequest, "invalid_artifact", "Artifact work object, type, title, and body are required, and any link must include a valid type and URL.")
			case errors.Is(err, sql.ErrNoRows):
				httpx.WriteError(w, http.StatusNotFound, "work_object_not_found", "Work Object not found.")
			default:
				httpx.WriteError(w, http.StatusInternalServerError, "create_artifact_failed", "Unable to create Artifact.")
			}
			return
		}
		httpx.WriteJSON(w, http.StatusCreated, artifact)
	default:
		httpx.MethodNotAllowed(w)
	}
}

func (s *Server) handleArtifactActions(w http.ResponseWriter, r *http.Request) {
	artifactID, ok := entityPathID(r.URL.Path, "/api/artifacts/")
	if !ok {
		http.NotFound(w, r)
		return
	}

	switch r.Method {
	case http.MethodGet:
		artifact, err := s.store.LookupArtifact(r.Context(), artifactID)
		if err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
				httpx.WriteError(w, http.StatusNotFound, "artifact_not_found", "Artifact not found.")
			default:
				httpx.WriteError(w, http.StatusInternalServerError, "artifact_failed", "Unable to load Artifact.")
			}
			return
		}
		httpx.WriteJSON(w, http.StatusOK, artifact)
	case http.MethodPut:
		var request artifactRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "invalid_artifact", "Artifact work object, type, title, and body are required.")
			return
		}
		artifact, err := s.store.UpdateArtifact(r.Context(), artifactID, request.WorkObjectID, request.Type, request.Title, request.Body, request.LinkType, request.URL)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrInvalidArtifactFields):
				httpx.WriteError(w, http.StatusBadRequest, "invalid_artifact", "Artifact work object, type, title, and body are required, and any link must include a valid type and URL.")
			case errors.Is(err, sql.ErrNoRows):
				httpx.WriteError(w, http.StatusNotFound, "artifact_not_found", "Artifact not found.")
			default:
				httpx.WriteError(w, http.StatusInternalServerError, "update_artifact_failed", "Unable to update Artifact.")
			}
			return
		}
		httpx.WriteJSON(w, http.StatusOK, artifact)
	case http.MethodDelete:
		if err := s.store.DeleteArtifact(r.Context(), artifactID); err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
				httpx.WriteError(w, http.StatusNotFound, "artifact_not_found", "Artifact not found.")
			default:
				httpx.WriteError(w, http.StatusInternalServerError, "delete_artifact_failed", "Unable to delete Artifact.")
			}
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		httpx.MethodNotAllowed(w)
	}
}

func entityPathID(path string, prefix string) (string, bool) {
	trimmed := strings.Trim(strings.TrimPrefix(path, prefix), "/")
	if trimmed == "" || strings.Contains(trimmed, "/") {
		return "", false
	}
	return trimmed, true
}
