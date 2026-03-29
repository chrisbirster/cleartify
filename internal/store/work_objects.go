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
var workObjectIDFromCluster = func(clusterID string, createdAt time.Time) string {
	return fmt.Sprintf("wo_%s_%d", strings.TrimPrefix(clusterID, "clu_"), createdAt.UTC().UnixNano())
}

var listWorkObjectsStatement = sqldata.MustQuery("list_work_objects.sql")
var insertWorkObjectFromClusterStatement = sqldata.MustQuery("insert_work_object_from_cluster.sql")
var lookupClusterForWorkObjectStatement = sqldata.MustQuery("lookup_cluster_for_work_object.sql")
var listWorkObjectDecisionsStatement = sqldata.MustQuery("list_work_object_decisions.sql")
var listWorkObjectArtifactsStatement = sqldata.MustQuery("list_work_object_artifacts.sql")

func (s *Store) ListWorkObjects(ctx context.Context) ([]domain.WorkObject, error) {
	rows, err := s.DB.QueryContext(ctx, listWorkObjectsStatement)
	if err != nil {
		return nil, fmt.Errorf("list work objects: %w", err)
	}
	defer rows.Close()

	var workObjects []domain.WorkObject
	for rows.Next() {
		var workObject domain.WorkObject
		var createdAt string
		var updatedAt string
		if err := rows.Scan(
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
			&createdAt,
			&updatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan work object: %w", err)
		}

		workObject.CreatedAt, err = decodeTime(createdAt)
		if err != nil {
			return nil, fmt.Errorf("decode work object created_at: %w", err)
		}
		workObject.UpdatedAt, err = decodeTime(updatedAt)
		if err != nil {
			return nil, fmt.Errorf("decode work object updated_at: %w", err)
		}
		workObject.Decisions, err = s.listDecisionsForWorkObject(ctx, workObject.ID)
		if err != nil {
			return nil, err
		}
		workObject.Artifacts, err = s.listArtifactsForWorkObject(ctx, workObject.ID)
		if err != nil {
			return nil, err
		}

		workObjects = append(workObjects, workObject)
	}

	return workObjects, rows.Err()
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
