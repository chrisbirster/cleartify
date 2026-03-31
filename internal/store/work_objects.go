package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	sqldata "cleartify/data"

	"cleartify/internal/domain"
)

var workObjectNow = time.Now
var decisionNow = time.Now
var artifactNow = time.Now
var workObjectIDFromCluster = func(clusterID string, createdAt time.Time) string {
	return fmt.Sprintf("wo_%s_%d", strings.TrimPrefix(clusterID, "clu_"), createdAt.UTC().UnixNano())
}
var decisionIDFromWorkObject = func(workObjectID string, createdAt time.Time) string {
	return fmt.Sprintf("dec_%s_%d", strings.TrimPrefix(workObjectID, "wo_"), createdAt.UTC().UnixNano())
}
var artifactIDFromWorkObject = func(workObjectID string, createdAt time.Time) string {
	return fmt.Sprintf("art_%s_%d", strings.TrimPrefix(workObjectID, "wo_"), createdAt.UTC().UnixNano())
}

var listWorkObjectsStatement = sqldata.MustQuery("list_work_objects.sql")
var lookupWorkObjectStatement = sqldata.MustQuery("lookup_work_object.sql")
var insertWorkObjectFromClusterStatement = sqldata.MustQuery("insert_work_object_from_cluster.sql")
var insertWorkObjectDecisionStatement = sqldata.MustQuery("insert_work_object_decision.sql")
var insertWorkObjectArtifactStatement = sqldata.MustQuery("insert_work_object_artifact.sql")
var updateWorkObjectCoreFieldsStatement = sqldata.MustQuery("update_work_object_core_fields.sql")
var updateWorkObjectPrototypeFieldsStatement = sqldata.MustQuery("update_work_object_prototype_fields.sql")
var updateWorkObjectStatusStatement = sqldata.MustQuery("update_work_object_status.sql")
var updateWorkObjectPriorityStatement = sqldata.MustQuery("update_work_object_priority.sql")
var lookupClusterForWorkObjectStatement = sqldata.MustQuery("lookup_cluster_for_work_object.sql")
var listWorkObjectDecisionsStatement = sqldata.MustQuery("list_work_object_decisions.sql")
var listWorkObjectArtifactsStatement = sqldata.MustQuery("list_work_object_artifacts.sql")

var ErrInvalidWorkObjectStatus = errors.New("invalid work object status")
var ErrInvalidWorkObjectPriority = errors.New("invalid work object priority")
var ErrInvalidWorkObjectCoreFields = errors.New("invalid work object core fields")
var ErrInvalidWorkObjectPrototypeFields = errors.New("invalid work object prototype fields")
var ErrInvalidDecisionFields = errors.New("invalid decision fields")
var ErrInvalidArtifactFields = errors.New("invalid artifact fields")

var allowedWorkObjectStatuses = map[string]struct{}{
	"planning":    {},
	"ready":       {},
	"in_progress": {},
	"in_review":   {},
	"shipped":     {},
}

var allowedWorkObjectPriorities = map[string]struct{}{
	"low":    {},
	"medium": {},
	"high":   {},
	"urgent": {},
}

func (s *Store) ListWorkObjects(ctx context.Context) ([]domain.WorkObject, error) {
	rows, err := s.DB.QueryContext(ctx, listWorkObjectsStatement)
	if err != nil {
		return nil, fmt.Errorf("list work objects: %w", err)
	}
	defer rows.Close()

	var workObjects []domain.WorkObject
	for rows.Next() {
		workObject, err := s.scanWorkObject(ctx, rows)
		if err != nil {
			return nil, fmt.Errorf("scan work object: %w", err)
		}
		workObjects = append(workObjects, workObject)
	}

	return workObjects, rows.Err()
}

func (s *Store) UpdateWorkObjectStatus(ctx context.Context, workObjectID string, status string) (domain.WorkObject, error) {
	if !isValidWorkObjectStatus(status) {
		return domain.WorkObject{}, ErrInvalidWorkObjectStatus
	}

	updatedAt := workObjectNow().UTC().Truncate(time.Second)
	result, err := s.DB.ExecContext(ctx, updateWorkObjectStatusStatement, status, updatedAt.Format(time.RFC3339), workObjectID)
	if err != nil {
		return domain.WorkObject{}, fmt.Errorf("update work object status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return domain.WorkObject{}, fmt.Errorf("update work object status rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return domain.WorkObject{}, sql.ErrNoRows
	}

	workObject, err := s.lookupWorkObject(ctx, workObjectID)
	if err != nil {
		return domain.WorkObject{}, err
	}
	return workObject, nil
}

func (s *Store) UpdateWorkObjectCoreFields(ctx context.Context, workObjectID string, title string, summary string, goal string) (domain.WorkObject, error) {
	title = strings.TrimSpace(title)
	summary = strings.TrimSpace(summary)
	goal = strings.TrimSpace(goal)
	if title == "" || summary == "" || goal == "" {
		return domain.WorkObject{}, ErrInvalidWorkObjectCoreFields
	}

	updatedAt := workObjectNow().UTC().Truncate(time.Second)
	result, err := s.DB.ExecContext(
		ctx,
		updateWorkObjectCoreFieldsStatement,
		title,
		summary,
		goal,
		updatedAt.Format(time.RFC3339),
		workObjectID,
	)
	if err != nil {
		return domain.WorkObject{}, fmt.Errorf("update work object core fields: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return domain.WorkObject{}, fmt.Errorf("update work object core field rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return domain.WorkObject{}, sql.ErrNoRows
	}

	workObject, err := s.lookupWorkObject(ctx, workObjectID)
	if err != nil {
		return domain.WorkObject{}, err
	}
	return workObject, nil
}

func (s *Store) UpdateWorkObjectPrototype(ctx context.Context, workObjectID string, prototypeURL string, prototypeNote string) (domain.WorkObject, error) {
	prototypeURL = strings.TrimSpace(prototypeURL)
	prototypeNote = strings.TrimSpace(prototypeNote)
	if prototypeURL == "" && prototypeNote == "" {
		return domain.WorkObject{}, ErrInvalidWorkObjectPrototypeFields
	}

	updatedAt := workObjectNow().UTC().Truncate(time.Second)
	result, err := s.DB.ExecContext(
		ctx,
		updateWorkObjectPrototypeFieldsStatement,
		prototypeURL,
		prototypeNote,
		updatedAt.Format(time.RFC3339),
		workObjectID,
	)
	if err != nil {
		return domain.WorkObject{}, fmt.Errorf("update work object prototype fields: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return domain.WorkObject{}, fmt.Errorf("update work object prototype field rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return domain.WorkObject{}, sql.ErrNoRows
	}

	workObject, err := s.lookupWorkObject(ctx, workObjectID)
	if err != nil {
		return domain.WorkObject{}, err
	}
	return workObject, nil
}

func (s *Store) UpdateWorkObjectPriority(ctx context.Context, workObjectID string, priority string) (domain.WorkObject, error) {
	if !isValidWorkObjectPriority(priority) {
		return domain.WorkObject{}, ErrInvalidWorkObjectPriority
	}

	updatedAt := workObjectNow().UTC().Truncate(time.Second)
	result, err := s.DB.ExecContext(ctx, updateWorkObjectPriorityStatement, priority, updatedAt.Format(time.RFC3339), workObjectID)
	if err != nil {
		return domain.WorkObject{}, fmt.Errorf("update work object priority: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return domain.WorkObject{}, fmt.Errorf("update work object priority rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return domain.WorkObject{}, sql.ErrNoRows
	}

	workObject, err := s.lookupWorkObject(ctx, workObjectID)
	if err != nil {
		return domain.WorkObject{}, err
	}
	return workObject, nil
}

func (s *Store) CreateDecisionForWorkObject(ctx context.Context, workObjectID string, title string, summary string, rationale string) (domain.Decision, error) {
	title = strings.TrimSpace(title)
	summary = strings.TrimSpace(summary)
	rationale = strings.TrimSpace(rationale)
	if title == "" || summary == "" || rationale == "" {
		return domain.Decision{}, ErrInvalidDecisionFields
	}

	if _, err := s.lookupWorkObject(ctx, workObjectID); err != nil {
		return domain.Decision{}, err
	}

	createdAt := decisionNow().UTC().Truncate(time.Second)
	decision := domain.Decision{
		ID:           decisionIDFromWorkObject(workObjectID, createdAt),
		WorkObjectID: workObjectID,
		Title:        title,
		Summary:      summary,
		Rationale:    rationale,
		CreatedAt:    createdAt,
	}

	if _, err := s.DB.ExecContext(
		ctx,
		insertWorkObjectDecisionStatement,
		decision.ID,
		decision.WorkObjectID,
		decision.Title,
		decision.Summary,
		decision.Rationale,
		decision.CreatedAt.Format(time.RFC3339),
	); err != nil {
		return domain.Decision{}, fmt.Errorf("insert decision for work object: %w", err)
	}

	return decision, nil
}

func (s *Store) CreateArtifactForWorkObject(ctx context.Context, workObjectID string, artifactType string, title string, body string) (domain.Artifact, error) {
	artifactType = strings.TrimSpace(artifactType)
	title = strings.TrimSpace(title)
	body = strings.TrimSpace(body)
	if artifactType == "" || title == "" || body == "" {
		return domain.Artifact{}, ErrInvalidArtifactFields
	}

	if _, err := s.lookupWorkObject(ctx, workObjectID); err != nil {
		return domain.Artifact{}, err
	}

	createdAt := artifactNow().UTC().Truncate(time.Second)
	artifact := domain.Artifact{
		ID:           artifactIDFromWorkObject(workObjectID, createdAt),
		WorkObjectID: workObjectID,
		ArtifactType: artifactType,
		Title:        title,
		Body:         body,
		Metadata:     nil,
		CreatedAt:    createdAt,
	}

	if err := s.persistArtifact(ctx, artifact); err != nil {
		return domain.Artifact{}, err
	}

	return artifact, nil
}

func (s *Store) GenerateImplementationArtifactForWorkObject(ctx context.Context, workObjectID string) (domain.Artifact, error) {
	workObject, draftPlan, linkedCluster, err := s.loadArtifactGenerationContext(ctx, workObjectID)
	if err != nil {
		return domain.Artifact{}, err
	}

	createdAt := artifactNow().UTC().Truncate(time.Second)
	artifact := domain.Artifact{
		ID:           artifactIDFromWorkObject(workObjectID, createdAt),
		WorkObjectID: workObjectID,
		ArtifactType: "implementation_brief",
		Title:        fmt.Sprintf("%s implementation brief", workObject.Title),
		Body:         buildImplementationArtifactBody(workObject, linkedCluster, draftPlan),
		Metadata:     nil,
		CreatedAt:    createdAt,
	}

	if err := s.persistArtifact(ctx, artifact); err != nil {
		return domain.Artifact{}, err
	}

	return artifact, nil
}

func (s *Store) GenerateTaskBreakdownArtifactForWorkObject(ctx context.Context, workObjectID string) (domain.Artifact, error) {
	workObject, draftPlan, linkedCluster, err := s.loadArtifactGenerationContext(ctx, workObjectID)
	if err != nil {
		return domain.Artifact{}, err
	}

	createdAt := artifactNow().UTC().Truncate(time.Second)
	artifact := domain.Artifact{
		ID:           artifactIDFromWorkObject(workObjectID, createdAt),
		WorkObjectID: workObjectID,
		ArtifactType: "task_breakdown",
		Title:        fmt.Sprintf("%s task breakdown", workObject.Title),
		Body:         buildTaskBreakdownArtifactBody(workObject, linkedCluster, draftPlan),
		Metadata:     nil,
		CreatedAt:    createdAt,
	}

	if err := s.persistArtifact(ctx, artifact); err != nil {
		return domain.Artifact{}, err
	}

	return artifact, nil
}

func (s *Store) CreateWorkObjectFromCluster(ctx context.Context, clusterID string) (domain.WorkObject, error) {
	cluster, err := s.lookupClusterForWorkObject(ctx, clusterID)
	if err != nil {
		return domain.WorkObject{}, err
	}

	createdAt := workObjectNow().UTC().Truncate(time.Second)
	workObject := domain.WorkObject{
		ID:        workObjectIDFromCluster(clusterID, createdAt),
		ClusterID: clusterID,
		Title:     cluster.Title,
		Summary:   cluster.Summary,
		Goal:      fmt.Sprintf("Turn the %s Cluster into executable product work.", cluster.Title),
		Status:    "planning",
		Priority:  workObjectPriorityForSeverity(cluster.Severity),
		OwnerType: "user",
		OwnerID:   "usr_demo",
		Decisions: []domain.Decision{},
		Artifacts: []domain.Artifact{},
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	}

	if _, err := s.DB.ExecContext(ctx, insertWorkObjectFromClusterStatement,
		workObject.ID,
		workObject.ClusterID,
		workObject.Title,
		workObject.Summary,
		workObject.Goal,
		workObject.Status,
		workObject.Priority,
		workObject.OwnerType,
		workObject.OwnerID,
		workObject.PrototypeURL,
		workObject.PrototypeNote,
		workObject.CreatedAt.Format(time.RFC3339),
		workObject.UpdatedAt.Format(time.RFC3339),
	); err != nil {
		return domain.WorkObject{}, fmt.Errorf("insert work object from cluster: %w", err)
	}

	return workObject, nil
}

func (s *Store) lookupClusterForWorkObject(ctx context.Context, clusterID string) (domain.Cluster, error) {
	row := s.DB.QueryRowContext(ctx, lookupClusterForWorkObjectStatement, clusterID)

	var cluster domain.Cluster
	var createdAt string
	var updatedAt string
	var err error
	if err := row.Scan(
		&cluster.ID,
		&cluster.Title,
		&cluster.Summary,
		&cluster.Severity,
		&cluster.Status,
		&createdAt,
		&updatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Cluster{}, sql.ErrNoRows
		}
		return domain.Cluster{}, fmt.Errorf("lookup cluster: %w", err)
	}

	cluster.CreatedAt, err = decodeTime(createdAt)
	if err != nil {
		return domain.Cluster{}, fmt.Errorf("decode cluster created_at: %w", err)
	}
	cluster.UpdatedAt, err = decodeTime(updatedAt)
	if err != nil {
		return domain.Cluster{}, fmt.Errorf("decode cluster updated_at: %w", err)
	}

	return cluster, nil
}

func workObjectPriorityForSeverity(severity string) string {
	switch severity {
	case "high":
		return "high"
	case "low":
		return "low"
	default:
		return "medium"
	}
}

func (s *Store) loadArtifactGenerationContext(ctx context.Context, workObjectID string) (domain.WorkObject, domain.DraftPlan, *domain.Cluster, error) {
	workObject, err := s.lookupWorkObject(ctx, workObjectID)
	if err != nil {
		return domain.WorkObject{}, domain.DraftPlan{}, nil, err
	}

	draftPlan, err := s.LookupDraftPlanForWorkObject(ctx, workObjectID)
	if err != nil {
		return domain.WorkObject{}, domain.DraftPlan{}, nil, err
	}
	if draftPlan == nil {
		return domain.WorkObject{}, domain.DraftPlan{}, nil, ErrDraftPlanNotFound
	}

	var linkedCluster *domain.Cluster
	if workObject.ClusterID != "" {
		cluster, err := s.lookupClusterForWorkObject(ctx, workObject.ClusterID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return domain.WorkObject{}, domain.DraftPlan{}, nil, err
		}
		if err == nil {
			linkedCluster = &cluster
		}
	}

	return workObject, *draftPlan, linkedCluster, nil
}

func isValidWorkObjectStatus(status string) bool {
	_, ok := allowedWorkObjectStatuses[status]
	return ok
}

func isValidWorkObjectPriority(priority string) bool {
	_, ok := allowedWorkObjectPriorities[priority]
	return ok
}

type workObjectScanner interface {
	Scan(dest ...any) error
}

func (s *Store) lookupWorkObject(ctx context.Context, workObjectID string) (domain.WorkObject, error) {
	row := s.DB.QueryRowContext(ctx, lookupWorkObjectStatement, workObjectID)
	workObject, err := s.scanWorkObject(ctx, row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.WorkObject{}, sql.ErrNoRows
		}
		return domain.WorkObject{}, fmt.Errorf("lookup work object: %w", err)
	}
	return workObject, nil
}

func (s *Store) scanWorkObject(ctx context.Context, scanner workObjectScanner) (domain.WorkObject, error) {
	var workObject domain.WorkObject
	var createdAt string
	var updatedAt string
	if err := scanner.Scan(
		&workObject.ID,
		&workObject.ClusterID,
		&workObject.Title,
		&workObject.Summary,
		&workObject.Goal,
		&workObject.Status,
		&workObject.Priority,
		&workObject.OwnerType,
		&workObject.OwnerID,
		&workObject.PrototypeURL,
		&workObject.PrototypeNote,
		&createdAt,
		&updatedAt,
	); err != nil {
		return domain.WorkObject{}, err
	}

	var err error
	workObject.CreatedAt, err = decodeTime(createdAt)
	if err != nil {
		return domain.WorkObject{}, fmt.Errorf("decode work object created_at: %w", err)
	}
	workObject.UpdatedAt, err = decodeTime(updatedAt)
	if err != nil {
		return domain.WorkObject{}, fmt.Errorf("decode work object updated_at: %w", err)
	}
	workObject.Decisions, err = s.listDecisionsForWorkObject(ctx, workObject.ID)
	if err != nil {
		return domain.WorkObject{}, err
	}
	workObject.Artifacts, err = s.listArtifactsForWorkObject(ctx, workObject.ID)
	if err != nil {
		return domain.WorkObject{}, err
	}

	return workObject, nil
}

func (s *Store) listDecisionsForWorkObject(ctx context.Context, workObjectID string) ([]domain.Decision, error) {
	rows, err := s.DB.QueryContext(ctx, listWorkObjectDecisionsStatement, workObjectID)
	if err != nil {
		return nil, fmt.Errorf("list decisions: %w", err)
	}
	defer rows.Close()

	var decisions []domain.Decision
	for rows.Next() {
		var decision domain.Decision
		var createdAt string
		if err := rows.Scan(
			&decision.ID,
			&decision.WorkObjectID,
			&decision.Title,
			&decision.Summary,
			&decision.Rationale,
			&createdAt,
		); err != nil {
			return nil, fmt.Errorf("scan decision: %w", err)
		}
		decision.CreatedAt, err = decodeTime(createdAt)
		if err != nil {
			return nil, fmt.Errorf("decode decision time: %w", err)
		}
		decisions = append(decisions, decision)
	}

	return decisions, rows.Err()
}

func (s *Store) listArtifactsForWorkObject(ctx context.Context, workObjectID string) ([]domain.Artifact, error) {
	rows, err := s.DB.QueryContext(ctx, listWorkObjectArtifactsStatement, workObjectID)
	if err != nil {
		return nil, fmt.Errorf("list artifacts: %w", err)
	}
	defer rows.Close()

	var artifacts []domain.Artifact
	for rows.Next() {
		var artifact domain.Artifact
		var rawMetadata sql.NullString
		var createdAt string
		if err := rows.Scan(
			&artifact.ID,
			&artifact.WorkObjectID,
			&artifact.ArtifactType,
			&artifact.Title,
			&artifact.Body,
			&rawMetadata,
			&createdAt,
		); err != nil {
			return nil, fmt.Errorf("scan artifact: %w", err)
		}
		artifact.Metadata, err = decodeJSONMap(rawMetadata)
		if err != nil {
			return nil, fmt.Errorf("decode artifact metadata: %w", err)
		}
		artifact.CreatedAt, err = decodeTime(createdAt)
		if err != nil {
			return nil, fmt.Errorf("decode artifact time: %w", err)
		}
		artifacts = append(artifacts, artifact)
	}

	return artifacts, rows.Err()
}

func (s *Store) persistArtifact(ctx context.Context, artifact domain.Artifact) error {
	if _, err := s.DB.ExecContext(
		ctx,
		insertWorkObjectArtifactStatement,
		artifact.ID,
		artifact.WorkObjectID,
		artifact.ArtifactType,
		artifact.Title,
		artifact.Body,
		nil,
		artifact.CreatedAt.Format(time.RFC3339),
	); err != nil {
		return fmt.Errorf("insert artifact for work object: %w", err)
	}
	return nil
}

func buildImplementationArtifactBody(workObject domain.WorkObject, linkedCluster *domain.Cluster, draftPlan domain.DraftPlan) string {
	clusterLine := "Linked Cluster: none attached yet."
	if linkedCluster != nil {
		clusterLine = fmt.Sprintf("Linked Cluster: %s (%s severity)", linkedCluster.Title, linkedCluster.Severity)
	}

	prototypeLink := "Prototype link: none attached yet."
	if workObject.PrototypeURL != "" {
		prototypeLink = fmt.Sprintf("Prototype link: %s", workObject.PrototypeURL)
	}

	prototypeNote := "Prototype note: none attached yet."
	if workObject.PrototypeNote != "" {
		prototypeNote = fmt.Sprintf("Prototype note: %s", workObject.PrototypeNote)
	}

	return strings.Join([]string{
		fmt.Sprintf("Implementation brief for %s", workObject.Title),
		"",
		"Overview",
		draftPlan.Overview,
		"",
		"Work Object context",
		fmt.Sprintf("Summary: %s", workObject.Summary),
		fmt.Sprintf("Goal: %s", workObject.Goal),
		clusterLine,
		prototypeLink,
		prototypeNote,
		"",
		"Implementation steps",
		formatDraftPlanSection(draftPlan.ImplementationSteps, true),
		"",
		"Edge cases",
		formatDraftPlanSection(draftPlan.EdgeCases, false),
		"",
		"Open questions",
		formatDraftPlanSection(draftPlan.OpenQuestions, false),
	}, "\n")
}

func buildTaskBreakdownArtifactBody(workObject domain.WorkObject, linkedCluster *domain.Cluster, draftPlan domain.DraftPlan) string {
	clusterLine := "Linked Cluster: none attached yet."
	if linkedCluster != nil {
		clusterLine = fmt.Sprintf("Linked Cluster: %s (%s severity)", linkedCluster.Title, linkedCluster.Severity)
	}

	prototypeLink := "Prototype link: none attached yet."
	if workObject.PrototypeURL != "" {
		prototypeLink = fmt.Sprintf("Prototype link: %s", workObject.PrototypeURL)
	}

	prototypeNote := "Prototype note: none attached yet."
	if workObject.PrototypeNote != "" {
		prototypeNote = fmt.Sprintf("Prototype note: %s", workObject.PrototypeNote)
	}

	var tasks []string
	for index, step := range draftPlan.ImplementationSteps {
		tasks = append(tasks, fmt.Sprintf("%d. %s", index+1, step))
	}

	return strings.Join([]string{
		fmt.Sprintf("Task breakdown for %s", workObject.Title),
		"",
		"Scope anchor",
		fmt.Sprintf("Summary: %s", workObject.Summary),
		fmt.Sprintf("Goal: %s", workObject.Goal),
		clusterLine,
		prototypeLink,
		prototypeNote,
		"",
		"Execution tasks",
		strings.Join(tasks, "\n"),
		"",
		"Edge cases to track",
		formatDraftPlanSection(draftPlan.EdgeCases, false),
		"",
		"Open questions to resolve",
		formatDraftPlanSection(draftPlan.OpenQuestions, false),
	}, "\n")
}

func formatDraftPlanSection(items []string, ordered bool) string {
	var lines []string
	for index, item := range items {
		prefix := "-"
		if ordered {
			prefix = fmt.Sprintf("%d.", index+1)
		}
		lines = append(lines, fmt.Sprintf("%s %s", prefix, item))
	}
	return strings.Join(lines, "\n")
}
