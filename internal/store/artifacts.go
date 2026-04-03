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
	"cleartify/internal/links"
)

var listArtifactsStatement = sqldata.MustQuery("list_artifacts.sql")
var lookupArtifactStatement = sqldata.MustQuery("lookup_artifact.sql")
var insertArtifactStatement = sqldata.MustQuery("insert_artifact.sql")
var updateArtifactStatement = sqldata.MustQuery("update_artifact.sql")
var deleteArtifactStatement = sqldata.MustQuery("delete_artifact.sql")

func (s *Store) ListArtifacts(ctx context.Context) ([]domain.Artifact, error) {
	rows, err := s.DB.QueryContext(ctx, listArtifactsStatement)
	if err != nil {
		return nil, fmt.Errorf("list artifacts: %w", err)
	}
	defer rows.Close()

	artifacts := make([]domain.Artifact, 0)
	for rows.Next() {
		artifact, err := scanArtifact(rows)
		if err != nil {
			return nil, err
		}
		artifacts = append(artifacts, artifact)
	}
	return artifacts, rows.Err()
}

func (s *Store) LookupArtifact(ctx context.Context, artifactID string) (domain.Artifact, error) {
	row := s.DB.QueryRowContext(ctx, lookupArtifactStatement, strings.TrimSpace(artifactID))
	artifact, err := scanArtifact(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Artifact{}, sql.ErrNoRows
		}
		return domain.Artifact{}, fmt.Errorf("lookup artifact: %w", err)
	}
	return artifact, nil
}

func (s *Store) CreateArtifact(ctx context.Context, workObjectID string, artifactType string, title string, body string, linkType string, externalURL string) (domain.Artifact, error) {
	return s.createArtifact(ctx, workObjectID, artifactType, title, body, linkType, externalURL)
}

func (s *Store) UpdateArtifact(ctx context.Context, artifactID string, workObjectID string, artifactType string, title string, body string, linkType string, externalURL string) (domain.Artifact, error) {
	workObjectID, artifactType, title, body, metadata, err := normalizeArtifactFields(workObjectID, artifactType, title, body, linkType, externalURL)
	if err != nil {
		return domain.Artifact{}, err
	}

	if _, err := s.lookupWorkObject(ctx, workObjectID); err != nil {
		return domain.Artifact{}, err
	}

	metadataJSON, err := encodeJSONMap(metadata)
	if err != nil {
		return domain.Artifact{}, fmt.Errorf("encode artifact metadata: %w", err)
	}

	result, err := s.DB.ExecContext(ctx, updateArtifactStatement, workObjectID, artifactType, title, body, metadataJSON, strings.TrimSpace(artifactID))
	if err != nil {
		return domain.Artifact{}, fmt.Errorf("update artifact: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return domain.Artifact{}, fmt.Errorf("update artifact rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return domain.Artifact{}, sql.ErrNoRows
	}

	return s.LookupArtifact(ctx, artifactID)
}

func (s *Store) DeleteArtifact(ctx context.Context, artifactID string) error {
	result, err := s.DB.ExecContext(ctx, deleteArtifactStatement, strings.TrimSpace(artifactID))
	if err != nil {
		return fmt.Errorf("delete artifact: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete artifact rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *Store) createArtifact(ctx context.Context, workObjectID string, artifactType string, title string, body string, linkType string, externalURL string) (domain.Artifact, error) {
	workObjectID, artifactType, title, body, metadata, err := normalizeArtifactFields(workObjectID, artifactType, title, body, linkType, externalURL)
	if err != nil {
		return domain.Artifact{}, err
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
		Metadata:     metadata,
		CreatedAt:    createdAt,
	}
	applyArtifactMetadata(&artifact)

	if err := s.persistArtifact(ctx, artifact); err != nil {
		return domain.Artifact{}, err
	}

	return artifact, nil
}

func normalizeArtifactFields(workObjectID string, artifactType string, title string, body string, linkType string, externalURL string) (string, string, string, string, map[string]any, error) {
	normalizedWorkObjectID := strings.TrimSpace(workObjectID)
	normalizedArtifactType := strings.TrimSpace(artifactType)
	normalizedTitle := strings.TrimSpace(title)
	normalizedBody := strings.TrimSpace(body)
	if normalizedWorkObjectID == "" || normalizedArtifactType == "" || normalizedTitle == "" || normalizedBody == "" {
		return "", "", "", "", nil, ErrInvalidArtifactFields
	}

	normalizedLinkType, normalizedURL, err := links.ValidateTypedExternalLink(linkType, externalURL)
	if err != nil {
		return "", "", "", "", nil, ErrInvalidArtifactFields
	}

	var metadata map[string]any
	if normalizedLinkType != "" && normalizedURL != "" {
		metadata = map[string]any{
			"linkType": normalizedLinkType,
			"url":      normalizedURL,
		}
	}

	return normalizedWorkObjectID, normalizedArtifactType, normalizedTitle, normalizedBody, metadata, nil
}

type artifactScanner interface {
	Scan(dest ...any) error
}

func scanArtifact(scanner artifactScanner) (domain.Artifact, error) {
	var artifact domain.Artifact
	var rawMetadata sql.NullString
	var createdAt string
	if err := scanner.Scan(
		&artifact.ID,
		&artifact.WorkObjectID,
		&artifact.ArtifactType,
		&artifact.Title,
		&artifact.Body,
		&rawMetadata,
		&createdAt,
	); err != nil {
		return domain.Artifact{}, fmt.Errorf("scan artifact: %w", err)
	}

	var err error
	artifact.Metadata, err = decodeJSONMap(rawMetadata)
	if err != nil {
		return domain.Artifact{}, fmt.Errorf("decode artifact metadata: %w", err)
	}
	applyArtifactMetadata(&artifact)
	artifact.CreatedAt, err = decodeTime(createdAt)
	if err != nil {
		return domain.Artifact{}, fmt.Errorf("decode artifact time: %w", err)
	}
	return artifact, nil
}

func applyArtifactMetadata(artifact *domain.Artifact) {
	if artifact == nil || artifact.Metadata == nil {
		return
	}
	if raw, ok := artifact.Metadata["linkType"].(string); ok {
		artifact.LinkType = raw
	}
	if raw, ok := artifact.Metadata["url"].(string); ok {
		artifact.URL = raw
	}
}

func (s *Store) persistArtifact(ctx context.Context, artifact domain.Artifact) error {
	metadataJSON, err := encodeJSONMap(artifact.Metadata)
	if err != nil {
		return fmt.Errorf("encode artifact metadata: %w", err)
	}
	if _, err := s.DB.ExecContext(
		ctx,
		insertArtifactStatement,
		artifact.ID,
		artifact.WorkObjectID,
		artifact.ArtifactType,
		artifact.Title,
		artifact.Body,
		metadataJSON,
		artifact.CreatedAt.Format(time.RFC3339),
	); err != nil {
		return fmt.Errorf("insert artifact for work object: %w", err)
	}
	return nil
}
