package server

import (
	"net/http"
	"strings"

	"cleartify/internal/domain"
	"cleartify/internal/httpx"
)

var sampleSignalIDs = map[string]struct{}{
	"sig_mobile_feedback": {},
	"sig_design_thread":   {},
	"sig_pr_context":      {},
	"sig_comp_notes":      {},
	"sig_proto_attach":    {},
	"sig_release_post":    {},
}

var sampleClusterIDs = map[string]struct{}{
	"clu_mobile_perf":        {},
	"clu_onboarding_refresh": {},
	"clu_product_memory":     {},
}

var sampleProjectIDs = map[string]struct{}{
	"prj_sample_cleartify": {},
}

var sampleDecisionIDs = map[string]struct{}{
	"dec_mobile_virtualize": {},
	"dec_onboarding_scope":  {},
}

var sampleArtifactIDs = map[string]struct{}{
	"art_mobile_plan":      {},
	"art_onboarding_notes": {},
}

var sampleWorkObjectPrefixes = []string{
	"wo_mobile_perf",
	"wo_onboarding_refresh",
	"wo_product_memory",
}

var sampleDecisionPrefixes = []string{
	"dec_mobile_perf",
	"dec_onboarding_refresh",
	"dec_product_memory",
}

var sampleArtifactPrefixes = []string{
	"art_mobile_perf",
	"art_onboarding_refresh",
	"art_product_memory",
}

func workspaceMode(session domain.AuthSession) string {
	if session.Workspace == nil {
		return ""
	}
	return strings.TrimSpace(session.Workspace.Mode)
}

func sessionShowsSampleData(session domain.AuthSession) bool {
	return workspaceMode(session) == "sample"
}

func filterSignalsForWorkspace(session domain.AuthSession, signals []domain.Signal) []domain.Signal {
	filtered := make([]domain.Signal, 0, len(signals))
	for _, signal := range signals {
		if signalVisibleInWorkspace(session, signal.ID) {
			filtered = append(filtered, signal)
		}
	}
	return filtered
}

func filterProjectsForWorkspace(session domain.AuthSession, projects []domain.Project) []domain.Project {
	filtered := make([]domain.Project, 0, len(projects))
	for _, project := range projects {
		if projectVisibleInWorkspace(session, project) {
			filtered = append(filtered, project)
		}
	}
	return filtered
}

func filterClustersForWorkspace(session domain.AuthSession, clusters []domain.Cluster) []domain.Cluster {
	filtered := make([]domain.Cluster, 0, len(clusters))
	for _, cluster := range clusters {
		if clusterVisibleInWorkspace(session, cluster.ID) {
			filtered = append(filtered, cluster)
		}
	}
	return filtered
}

func filterWorkObjectsForWorkspace(session domain.AuthSession, workObjects []domain.WorkObject) []domain.WorkObject {
	filtered := make([]domain.WorkObject, 0, len(workObjects))
	for _, workObject := range workObjects {
		if workObjectVisibleInWorkspace(session, workObject) {
			filtered = append(filtered, workObject)
		}
	}
	return filtered
}

func filterDecisionsForWorkspace(session domain.AuthSession, decisions []domain.Decision) []domain.Decision {
	filtered := make([]domain.Decision, 0, len(decisions))
	for _, decision := range decisions {
		if decisionVisibleInWorkspace(session, decision) {
			filtered = append(filtered, decision)
		}
	}
	return filtered
}

func filterArtifactsForWorkspace(session domain.AuthSession, artifacts []domain.Artifact) []domain.Artifact {
	filtered := make([]domain.Artifact, 0, len(artifacts))
	for _, artifact := range artifacts {
		if artifactVisibleInWorkspace(session, artifact) {
			filtered = append(filtered, artifact)
		}
	}
	return filtered
}

func filterSignalsForProject(projectID string, signals []domain.Signal) []domain.Signal {
	if strings.TrimSpace(projectID) == "" {
		return signals
	}

	filtered := make([]domain.Signal, 0, len(signals))
	for _, signal := range signals {
		if strings.TrimSpace(signal.ProjectID) == strings.TrimSpace(projectID) {
			filtered = append(filtered, signal)
		}
	}
	return filtered
}

func filterClustersForProject(projectID string, clusters []domain.Cluster) []domain.Cluster {
	if strings.TrimSpace(projectID) == "" {
		return clusters
	}

	filtered := make([]domain.Cluster, 0, len(clusters))
	for _, cluster := range clusters {
		if strings.TrimSpace(cluster.ProjectID) == strings.TrimSpace(projectID) {
			filtered = append(filtered, cluster)
		}
	}
	return filtered
}

func filterWorkObjectsForProject(projectID string, workObjects []domain.WorkObject) []domain.WorkObject {
	if strings.TrimSpace(projectID) == "" {
		return workObjects
	}

	filtered := make([]domain.WorkObject, 0, len(workObjects))
	for _, workObject := range workObjects {
		if strings.TrimSpace(workObject.ProjectID) == strings.TrimSpace(projectID) {
			filtered = append(filtered, workObject)
		}
	}
	return filtered
}

func filterDecisionsForProject(projectID string, decisions []domain.Decision, workObjects []domain.WorkObject) []domain.Decision {
	if strings.TrimSpace(projectID) == "" {
		return decisions
	}

	workObjectProjectIDs := make(map[string]string, len(workObjects))
	for _, workObject := range workObjects {
		workObjectProjectIDs[workObject.ID] = workObject.ProjectID
	}

	filtered := make([]domain.Decision, 0, len(decisions))
	for _, decision := range decisions {
		if strings.TrimSpace(workObjectProjectIDs[decision.WorkObjectID]) == strings.TrimSpace(projectID) {
			filtered = append(filtered, decision)
		}
	}
	return filtered
}

func filterArtifactsForProject(projectID string, artifacts []domain.Artifact, workObjects []domain.WorkObject) []domain.Artifact {
	if strings.TrimSpace(projectID) == "" {
		return artifacts
	}

	workObjectProjectIDs := make(map[string]string, len(workObjects))
	for _, workObject := range workObjects {
		workObjectProjectIDs[workObject.ID] = workObject.ProjectID
	}

	filtered := make([]domain.Artifact, 0, len(artifacts))
	for _, artifact := range artifacts {
		if strings.TrimSpace(workObjectProjectIDs[artifact.WorkObjectID]) == strings.TrimSpace(projectID) {
			filtered = append(filtered, artifact)
		}
	}
	return filtered
}

func signalVisibleInWorkspace(session domain.AuthSession, signalID string) bool {
	return entityVisibleForWorkspace(session, isSampleSignalID(signalID))
}

func projectVisibleInWorkspace(session domain.AuthSession, project domain.Project) bool {
	if isSampleProjectID(project.ID) {
		return sessionShowsSampleData(session)
	}
	if session.Workspace == nil {
		return false
	}
	return strings.TrimSpace(project.WorkspaceID) == strings.TrimSpace(session.Workspace.ID)
}

func clusterVisibleInWorkspace(session domain.AuthSession, clusterID string) bool {
	return entityVisibleForWorkspace(session, isSampleClusterID(clusterID))
}

func workObjectVisibleInWorkspace(session domain.AuthSession, workObject domain.WorkObject) bool {
	return entityVisibleForWorkspace(session, isSampleWorkObject(workObject))
}

func workObjectIDVisibleInWorkspace(session domain.AuthSession, workObjectID string) bool {
	return entityVisibleForWorkspace(session, isSampleWorkObjectID(workObjectID))
}

func decisionVisibleInWorkspace(session domain.AuthSession, decision domain.Decision) bool {
	return entityVisibleForWorkspace(session, isSampleDecision(decision))
}

func decisionIDVisibleInWorkspace(session domain.AuthSession, decisionID string) bool {
	return entityVisibleForWorkspace(session, isSampleDecisionID(decisionID))
}

func artifactVisibleInWorkspace(session domain.AuthSession, artifact domain.Artifact) bool {
	return entityVisibleForWorkspace(session, isSampleArtifact(artifact))
}

func artifactIDVisibleInWorkspace(session domain.AuthSession, artifactID string) bool {
	return entityVisibleForWorkspace(session, isSampleArtifactID(artifactID))
}

func requireVisibleClusterSession(s *Server, w http.ResponseWriter, r *http.Request, clusterID string) (domain.AuthSession, bool) {
	session, ok := s.requireAuthenticatedSession(w, r)
	if !ok {
		return domain.AuthSession{}, false
	}
	if !clusterVisibleInWorkspace(session, clusterID) {
		httpx.WriteError(w, http.StatusNotFound, "cluster_not_found", "Cluster not found.")
		return domain.AuthSession{}, false
	}
	return session, true
}

func requireVisibleWorkObjectSession(s *Server, w http.ResponseWriter, r *http.Request, workObjectID string) (domain.AuthSession, bool) {
	session, ok := s.requireAuthenticatedSession(w, r)
	if !ok {
		return domain.AuthSession{}, false
	}
	if !workObjectIDVisibleInWorkspace(session, workObjectID) {
		httpx.WriteError(w, http.StatusNotFound, "work_object_not_found", "Work Object not found.")
		return domain.AuthSession{}, false
	}
	return session, true
}

func requireVisibleSignalSession(s *Server, w http.ResponseWriter, r *http.Request, signalID string) (domain.AuthSession, bool) {
	session, ok := s.requireAuthenticatedSession(w, r)
	if !ok {
		return domain.AuthSession{}, false
	}
	if !signalVisibleInWorkspace(session, signalID) {
		httpx.WriteError(w, http.StatusNotFound, "signal_not_found", "Signal not found.")
		return domain.AuthSession{}, false
	}
	return session, true
}

func requireVisibleDecisionSession(s *Server, w http.ResponseWriter, r *http.Request, decisionID string) (domain.AuthSession, bool) {
	session, ok := s.requireAuthenticatedSession(w, r)
	if !ok {
		return domain.AuthSession{}, false
	}
	if !decisionIDVisibleInWorkspace(session, decisionID) {
		httpx.WriteError(w, http.StatusNotFound, "decision_not_found", "Decision not found.")
		return domain.AuthSession{}, false
	}
	return session, true
}

func requireVisibleArtifactSession(s *Server, w http.ResponseWriter, r *http.Request, artifactID string) (domain.AuthSession, bool) {
	session, ok := s.requireAuthenticatedSession(w, r)
	if !ok {
		return domain.AuthSession{}, false
	}
	if !artifactIDVisibleInWorkspace(session, artifactID) {
		httpx.WriteError(w, http.StatusNotFound, "artifact_not_found", "Artifact not found.")
		return domain.AuthSession{}, false
	}
	return session, true
}

func entityVisibleForWorkspace(session domain.AuthSession, isSample bool) bool {
	if sessionShowsSampleData(session) {
		return true
	}
	return !isSample
}

func isSampleSignalID(signalID string) bool {
	_, ok := sampleSignalIDs[strings.TrimSpace(signalID)]
	return ok
}

func isSampleClusterID(clusterID string) bool {
	_, ok := sampleClusterIDs[strings.TrimSpace(clusterID)]
	return ok
}

func isSampleProjectID(projectID string) bool {
	_, ok := sampleProjectIDs[strings.TrimSpace(projectID)]
	return ok
}

func isSampleWorkObject(workObject domain.WorkObject) bool {
	if strings.TrimSpace(workObject.OwnerID) == "usr_demo" {
		return true
	}
	if isSampleClusterID(workObject.ClusterID) {
		return true
	}
	return isSampleWorkObjectID(workObject.ID)
}

func isSampleWorkObjectID(workObjectID string) bool {
	return hasSamplePrefix(workObjectID, sampleWorkObjectPrefixes)
}

func isSampleDecision(decision domain.Decision) bool {
	if isSampleWorkObjectID(decision.WorkObjectID) {
		return true
	}
	return isSampleDecisionID(decision.ID)
}

func isSampleDecisionID(decisionID string) bool {
	if _, ok := sampleDecisionIDs[strings.TrimSpace(decisionID)]; ok {
		return true
	}
	return hasSamplePrefix(decisionID, sampleDecisionPrefixes)
}

func isSampleArtifact(artifact domain.Artifact) bool {
	if isSampleWorkObjectID(artifact.WorkObjectID) {
		return true
	}
	return isSampleArtifactID(artifact.ID)
}

func isSampleArtifactID(artifactID string) bool {
	if _, ok := sampleArtifactIDs[strings.TrimSpace(artifactID)]; ok {
		return true
	}
	return hasSamplePrefix(artifactID, sampleArtifactPrefixes)
}

func hasSamplePrefix(value string, prefixes []string) bool {
	trimmed := strings.TrimSpace(value)
	for _, prefix := range prefixes {
		if trimmed == prefix || strings.HasPrefix(trimmed, prefix+"_") {
			return true
		}
	}
	return false
}
