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

var listProjectsStatement = sqldata.MustQuery("list_projects.sql")
var lookupProjectStatement = sqldata.MustQuery("lookup_project.sql")
var insertProjectStatement = sqldata.MustQuery("insert_project.sql")
var updateProjectStatement = sqldata.MustQuery("update_project.sql")
var projectNow = time.Now
var projectIDFromTime = func(createdAt time.Time) string {
	return fmt.Sprintf("prj_%d", createdAt.UTC().UnixNano())
}

var ErrInvalidProjectFields = errors.New("invalid project fields")

var allowedProjectStatuses = map[string]struct{}{
	"active":   {},
	"paused":   {},
	"archived": {},
}

func (s *Store) ListProjects(ctx context.Context) ([]domain.Project, error) {
	rows, err := s.DB.QueryContext(ctx, listProjectsStatement)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	defer rows.Close()

	projects := make([]domain.Project, 0)
	for rows.Next() {
		project, err := scanProject(rows)
		if err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}

	return projects, rows.Err()
}

func (s *Store) LookupProject(ctx context.Context, projectID string) (domain.Project, error) {
	row := s.DB.QueryRowContext(ctx, lookupProjectStatement, strings.TrimSpace(projectID))
	project, err := scanProject(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Project{}, sql.ErrNoRows
		}
		return domain.Project{}, fmt.Errorf("lookup project: %w", err)
	}
	return project, nil
}

func (s *Store) CreateProject(ctx context.Context, workspaceID string, name string, summary string, status string) (domain.Project, error) {
	workspaceID, name, slug, summary, status, err := normalizeProjectFields(workspaceID, name, summary, status)
	if err != nil {
		return domain.Project{}, err
	}

	createdAt := projectNow().UTC().Truncate(time.Second)
	project := domain.Project{
		ID:          projectIDFromTime(createdAt),
		WorkspaceID: workspaceID,
		Name:        name,
		Slug:        slug,
		Summary:     summary,
		Status:      status,
		CreatedAt:   createdAt,
		UpdatedAt:   createdAt,
	}

	if _, err := s.DB.ExecContext(
		ctx,
		insertProjectStatement,
		project.ID,
		project.WorkspaceID,
		project.Name,
		project.Slug,
		project.Summary,
		project.Status,
		project.CreatedAt.Format(time.RFC3339),
		project.UpdatedAt.Format(time.RFC3339),
	); err != nil {
		return domain.Project{}, fmt.Errorf("insert project: %w", err)
	}

	return project, nil
}

func (s *Store) UpdateProject(ctx context.Context, projectID string, name string, summary string, status string) (domain.Project, error) {
	existingProject, err := s.LookupProject(ctx, projectID)
	if err != nil {
		return domain.Project{}, err
	}

	_, name, slug, summary, status, err := normalizeProjectFields(existingProject.WorkspaceID, name, summary, status)
	if err != nil {
		return domain.Project{}, err
	}

	updatedAt := projectNow().UTC().Truncate(time.Second)
	result, err := s.DB.ExecContext(
		ctx,
		updateProjectStatement,
		name,
		slug,
		summary,
		status,
		updatedAt.Format(time.RFC3339),
		strings.TrimSpace(projectID),
	)
	if err != nil {
		return domain.Project{}, fmt.Errorf("update project: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return domain.Project{}, fmt.Errorf("update project rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return domain.Project{}, sql.ErrNoRows
	}

	return s.LookupProject(ctx, projectID)
}

type projectScanner interface {
	Scan(dest ...any) error
}

func scanProject(scanner projectScanner) (domain.Project, error) {
	var project domain.Project
	var createdAt string
	var updatedAt string
	if err := scanner.Scan(
		&project.ID,
		&project.WorkspaceID,
		&project.Name,
		&project.Slug,
		&project.Summary,
		&project.Status,
		&createdAt,
		&updatedAt,
	); err != nil {
		return domain.Project{}, fmt.Errorf("scan project: %w", err)
	}

	var err error
	project.CreatedAt, err = decodeTime(createdAt)
	if err != nil {
		return domain.Project{}, fmt.Errorf("decode project created_at: %w", err)
	}
	project.UpdatedAt, err = decodeTime(updatedAt)
	if err != nil {
		return domain.Project{}, fmt.Errorf("decode project updated_at: %w", err)
	}
	return project, nil
}

func normalizeProjectFields(workspaceID string, name string, summary string, status string) (string, string, string, string, string, error) {
	normalizedWorkspaceID := strings.TrimSpace(workspaceID)
	normalizedName := strings.TrimSpace(name)
	normalizedSummary := strings.TrimSpace(summary)
	normalizedStatus := strings.TrimSpace(status)
	if normalizedStatus == "" {
		normalizedStatus = "active"
	}
	if normalizedWorkspaceID == "" || normalizedName == "" || normalizedSummary == "" {
		return "", "", "", "", "", ErrInvalidProjectFields
	}
	if _, ok := allowedProjectStatuses[normalizedStatus]; !ok {
		return "", "", "", "", "", ErrInvalidProjectFields
	}

	slug := slugify(normalizedName)
	if slug == "" {
		return "", "", "", "", "", ErrInvalidProjectFields
	}

	return normalizedWorkspaceID, normalizedName, slug, normalizedSummary, normalizedStatus, nil
}
