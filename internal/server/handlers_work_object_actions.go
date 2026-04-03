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

type updateWorkObjectStatusRequest struct {
	Status string `json:"status"`
}

type updateWorkObjectCoreFieldsRequest struct {
	Title   string `json:"title"`
	Summary string `json:"summary"`
	Goal    string `json:"goal"`
}

type updateWorkObjectPrototypeRequest struct {
	PrototypeURL  string `json:"prototypeUrl"`
	PrototypeNote string `json:"prototypeNote"`
}

type updateWorkObjectPriorityRequest struct {
	Priority string `json:"priority"`
}

type createDecisionRequest struct {
	Title     string `json:"title"`
	Summary   string `json:"summary"`
	Rationale string `json:"rationale"`
}

type createArtifactRequest struct {
	Type  string `json:"type"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

type updateDraftPlanRequest struct {
	Overview            string   `json:"overview"`
	ImplementationSteps []string `json:"implementationSteps"`
	EdgeCases           []string `json:"edgeCases"`
	OpenQuestions       []string `json:"openQuestions"`
}

func (s *Server) handleWorkObjectActions(w http.ResponseWriter, r *http.Request) {
	workObjectID, action, ok := workObjectActionPath(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}
	if _, ok := requireVisibleWorkObjectSession(s, w, r, workObjectID); !ok {
		return
	}

	switch action {
	case "plan":
		switch r.Method {
		case http.MethodGet:
			s.handleWorkObjectDraftPlanLookup(w, r, workObjectID)
		case http.MethodPost:
			s.handleWorkObjectDraftPlanGenerate(w, r, workObjectID)
		case http.MethodPut:
			s.handleWorkObjectDraftPlanUpdate(w, r, workObjectID)
		default:
			httpx.MethodNotAllowed(w)
		}
	case "core":
		if r.Method != http.MethodPost {
			httpx.MethodNotAllowed(w)
			return
		}
		s.handleWorkObjectCoreFieldUpdate(w, r, workObjectID)
	case "prototype":
		if r.Method != http.MethodPost {
			httpx.MethodNotAllowed(w)
			return
		}
		s.handleWorkObjectPrototypeUpdate(w, r, workObjectID)
	case "status":
		if r.Method != http.MethodPost {
			httpx.MethodNotAllowed(w)
			return
		}
		s.handleWorkObjectStatusUpdate(w, r, workObjectID)
	case "priority":
		if r.Method != http.MethodPost {
			httpx.MethodNotAllowed(w)
			return
		}
		s.handleWorkObjectPriorityUpdate(w, r, workObjectID)
	case "release-ready":
		if r.Method != http.MethodPost {
			httpx.MethodNotAllowed(w)
			return
		}
		s.handleWorkObjectReleaseReady(w, r, workObjectID)
	case "decisions":
		if r.Method != http.MethodPost {
			httpx.MethodNotAllowed(w)
			return
		}
		s.handleWorkObjectDecisionCreate(w, r, workObjectID)
	case "artifacts":
		if r.Method != http.MethodPost {
			httpx.MethodNotAllowed(w)
			return
		}
		s.handleWorkObjectArtifactCreate(w, r, workObjectID)
	case "implementation-artifact":
		if r.Method != http.MethodPost {
			httpx.MethodNotAllowed(w)
			return
		}
		s.handleWorkObjectImplementationArtifactGenerate(w, r, workObjectID)
	case "task-breakdown-artifact":
		if r.Method != http.MethodPost {
			httpx.MethodNotAllowed(w)
			return
		}
		s.handleWorkObjectTaskBreakdownArtifactGenerate(w, r, workObjectID)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) handleWorkObjectDraftPlanLookup(w http.ResponseWriter, r *http.Request, workObjectID string) {
	draftPlan, err := s.store.LookupDraftPlanForWorkObject(r.Context(), workObjectID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			httpx.WriteError(w, http.StatusNotFound, "draft_plan_not_found", "Draft plan not found.")
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "load_draft_plan_failed", "Unable to load draft plan.")
		}
		return
	}
	if draftPlan == nil {
		httpx.WriteError(w, http.StatusNotFound, "draft_plan_not_found", "Draft plan not found.")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, draftPlan)
}

func (s *Server) handleWorkObjectDraftPlanGenerate(w http.ResponseWriter, r *http.Request, workObjectID string) {
	draftPlan, err := s.store.GenerateDraftPlanForWorkObject(r.Context(), workObjectID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			httpx.WriteError(w, http.StatusNotFound, "work_object_not_found", "Work Object not found.")
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "generate_draft_plan_failed", "Unable to generate draft plan.")
		}
		return
	}

	httpx.WriteJSON(w, http.StatusOK, draftPlan)
}

func (s *Server) handleWorkObjectDraftPlanUpdate(w http.ResponseWriter, r *http.Request, workObjectID string) {
	var request updateDraftPlanRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_draft_plan", "Draft plan sections are invalid.")
		return
	}

	draftPlan, err := s.store.UpdateDraftPlanForWorkObject(
		r.Context(),
		workObjectID,
		request.Overview,
		request.ImplementationSteps,
		request.EdgeCases,
		request.OpenQuestions,
	)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrInvalidDraftPlanFields):
			httpx.WriteError(w, http.StatusBadRequest, "invalid_draft_plan", "Draft plan sections are invalid.")
		case errors.Is(err, store.ErrDraftPlanNotFound):
			httpx.WriteError(w, http.StatusNotFound, "draft_plan_not_found", "Draft plan not found.")
		case errors.Is(err, sql.ErrNoRows):
			httpx.WriteError(w, http.StatusNotFound, "work_object_not_found", "Work Object not found.")
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "update_draft_plan_failed", "Unable to update draft plan.")
		}
		return
	}

	httpx.WriteJSON(w, http.StatusOK, draftPlan)
}

func (s *Server) handleWorkObjectCoreFieldUpdate(w http.ResponseWriter, r *http.Request, workObjectID string) {
	var request updateWorkObjectCoreFieldsRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil ||
		strings.TrimSpace(request.Title) == "" ||
		strings.TrimSpace(request.Summary) == "" ||
		strings.TrimSpace(request.Goal) == "" {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_work_object_core_fields", "Work Object details are invalid.")
		return
	}

	workObject, err := s.store.UpdateWorkObjectCoreFields(r.Context(), workObjectID, request.Title, request.Summary, request.Goal)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrInvalidWorkObjectCoreFields):
			httpx.WriteError(w, http.StatusBadRequest, "invalid_work_object_core_fields", "Work Object details are invalid.")
		case errors.Is(err, sql.ErrNoRows):
			httpx.WriteError(w, http.StatusNotFound, "work_object_not_found", "Work Object not found.")
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "update_work_object_core_fields_failed", "Unable to update Work Object details.")
		}
		return
	}

	httpx.WriteJSON(w, http.StatusOK, workObject)
}

func (s *Server) handleWorkObjectPrototypeUpdate(w http.ResponseWriter, r *http.Request, workObjectID string) {
	var request updateWorkObjectPrototypeRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil ||
		(strings.TrimSpace(request.PrototypeURL) == "" && strings.TrimSpace(request.PrototypeNote) == "") {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_work_object_prototype", "Prototype link or note is required.")
		return
	}

	workObject, err := s.store.UpdateWorkObjectPrototype(r.Context(), workObjectID, request.PrototypeURL, request.PrototypeNote)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrInvalidWorkObjectPrototypeFields):
			httpx.WriteError(w, http.StatusBadRequest, "invalid_work_object_prototype", "Prototype link or note is required.")
		case errors.Is(err, sql.ErrNoRows):
			httpx.WriteError(w, http.StatusNotFound, "work_object_not_found", "Work Object not found.")
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "update_work_object_prototype_failed", "Unable to update Work Object prototype context.")
		}
		return
	}

	httpx.WriteJSON(w, http.StatusOK, workObject)
}

func (s *Server) handleWorkObjectStatusUpdate(w http.ResponseWriter, r *http.Request, workObjectID string) {
	var request updateWorkObjectStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil || request.Status == "" {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_work_object_status", "Work Object status is invalid.")
		return
	}

	workObject, err := s.store.UpdateWorkObjectStatus(r.Context(), workObjectID, request.Status)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrInvalidWorkObjectStatus):
			httpx.WriteError(w, http.StatusBadRequest, "invalid_work_object_status", "Work Object status is invalid.")
		case errors.Is(err, sql.ErrNoRows):
			httpx.WriteError(w, http.StatusNotFound, "work_object_not_found", "Work Object not found.")
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "update_work_object_status_failed", "Unable to update Work Object status.")
		}
		return
	}

	httpx.WriteJSON(w, http.StatusOK, workObject)
}

func (s *Server) handleWorkObjectPriorityUpdate(w http.ResponseWriter, r *http.Request, workObjectID string) {
	var request updateWorkObjectPriorityRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil || request.Priority == "" {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_work_object_priority", "Work Object priority is invalid.")
		return
	}

	workObject, err := s.store.UpdateWorkObjectPriority(r.Context(), workObjectID, request.Priority)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrInvalidWorkObjectPriority):
			httpx.WriteError(w, http.StatusBadRequest, "invalid_work_object_priority", "Work Object priority is invalid.")
		case errors.Is(err, sql.ErrNoRows):
			httpx.WriteError(w, http.StatusNotFound, "work_object_not_found", "Work Object not found.")
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "update_work_object_priority_failed", "Unable to update Work Object priority.")
		}
		return
	}

	httpx.WriteJSON(w, http.StatusOK, workObject)
}

func (s *Server) handleWorkObjectReleaseReady(w http.ResponseWriter, r *http.Request, workObjectID string) {
	workObject, err := s.store.MarkWorkObjectReadyForRelease(r.Context(), workObjectID)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrDraftPlanNotFound):
			httpx.WriteError(w, http.StatusConflict, "draft_plan_not_found", "Persist a draft plan before marking the Work Object ready for release.")
		case errors.Is(err, sql.ErrNoRows):
			httpx.WriteError(w, http.StatusNotFound, "work_object_not_found", "Work Object not found.")
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "mark_work_object_release_ready_failed", "Unable to mark the Work Object ready for release.")
		}
		return
	}

	httpx.WriteJSON(w, http.StatusOK, workObject)
}

func (s *Server) handleWorkObjectDecisionCreate(w http.ResponseWriter, r *http.Request, workObjectID string) {
	var request createDecisionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil ||
		strings.TrimSpace(request.Title) == "" ||
		strings.TrimSpace(request.Summary) == "" ||
		strings.TrimSpace(request.Rationale) == "" {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_decision", "Decision title, summary, and rationale are required.")
		return
	}

	decision, err := s.store.CreateDecisionForWorkObject(r.Context(), workObjectID, request.Title, request.Summary, request.Rationale)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrInvalidDecisionFields):
			httpx.WriteError(w, http.StatusBadRequest, "invalid_decision", "Decision title, summary, and rationale are required.")
		case errors.Is(err, sql.ErrNoRows):
			httpx.WriteError(w, http.StatusNotFound, "work_object_not_found", "Work Object not found.")
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "create_decision_failed", "Unable to create Decision.")
		}
		return
	}

	httpx.WriteJSON(w, http.StatusOK, decision)
}

func (s *Server) handleWorkObjectArtifactCreate(w http.ResponseWriter, r *http.Request, workObjectID string) {
	var request createArtifactRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil ||
		strings.TrimSpace(request.Type) == "" ||
		strings.TrimSpace(request.Title) == "" ||
		strings.TrimSpace(request.Body) == "" {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_artifact", "Artifact type, title, and body are required.")
		return
	}

	artifact, err := s.store.CreateArtifactForWorkObject(r.Context(), workObjectID, request.Type, request.Title, request.Body)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrInvalidArtifactFields):
			httpx.WriteError(w, http.StatusBadRequest, "invalid_artifact", "Artifact type, title, and body are required.")
		case errors.Is(err, sql.ErrNoRows):
			httpx.WriteError(w, http.StatusNotFound, "work_object_not_found", "Work Object not found.")
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "create_artifact_failed", "Unable to create Artifact.")
		}
		return
	}

	httpx.WriteJSON(w, http.StatusOK, artifact)
}

func (s *Server) handleWorkObjectImplementationArtifactGenerate(w http.ResponseWriter, r *http.Request, workObjectID string) {
	artifact, err := s.store.GenerateImplementationArtifactForWorkObject(r.Context(), workObjectID)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrDraftPlanNotFound):
			httpx.WriteError(w, http.StatusNotFound, "draft_plan_not_found", "Draft plan not found.")
		case errors.Is(err, sql.ErrNoRows):
			httpx.WriteError(w, http.StatusNotFound, "work_object_not_found", "Work Object not found.")
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "generate_implementation_artifact_failed", "Unable to generate implementation Artifact.")
		}
		return
	}

	httpx.WriteJSON(w, http.StatusOK, artifact)
}

func (s *Server) handleWorkObjectTaskBreakdownArtifactGenerate(w http.ResponseWriter, r *http.Request, workObjectID string) {
	artifact, err := s.store.GenerateTaskBreakdownArtifactForWorkObject(r.Context(), workObjectID)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrDraftPlanNotFound):
			httpx.WriteError(w, http.StatusNotFound, "draft_plan_not_found", "Draft plan not found.")
		case errors.Is(err, sql.ErrNoRows):
			httpx.WriteError(w, http.StatusNotFound, "work_object_not_found", "Work Object not found.")
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "generate_task_breakdown_artifact_failed", "Unable to generate task-breakdown Artifact.")
		}
		return
	}

	httpx.WriteJSON(w, http.StatusOK, artifact)
}

func workObjectActionPath(path string) (workObjectID string, action string, ok bool) {
	trimmed := strings.Trim(strings.TrimPrefix(path, "/api/work-objects/"), "/")
	parts := strings.Split(trimmed, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", false
	}

	return parts[0], parts[1], true
}
