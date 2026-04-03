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

var listDecisionsStatement = sqldata.MustQuery("list_decisions.sql")
var lookupDecisionStatement = sqldata.MustQuery("lookup_decision.sql")
var insertDecisionStatement = sqldata.MustQuery("insert_decision.sql")
var updateDecisionStatement = sqldata.MustQuery("update_decision.sql")
var deleteDecisionStatement = sqldata.MustQuery("delete_decision.sql")

func (s *Store) ListDecisions(ctx context.Context) ([]domain.Decision, error) {
	rows, err := s.DB.QueryContext(ctx, listDecisionsStatement)
	if err != nil {
		return nil, fmt.Errorf("list decisions: %w", err)
	}
	defer rows.Close()

	decisions := make([]domain.Decision, 0)
	for rows.Next() {
		decision, err := scanDecision(rows)
		if err != nil {
			return nil, err
		}
		decisions = append(decisions, decision)
	}
	return decisions, rows.Err()
}

func (s *Store) LookupDecision(ctx context.Context, decisionID string) (domain.Decision, error) {
	row := s.DB.QueryRowContext(ctx, lookupDecisionStatement, strings.TrimSpace(decisionID))
	decision, err := scanDecision(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Decision{}, sql.ErrNoRows
		}
		return domain.Decision{}, fmt.Errorf("lookup decision: %w", err)
	}
	return decision, nil
}

func (s *Store) CreateDecision(ctx context.Context, workObjectID string, title string, summary string, rationale string) (domain.Decision, error) {
	return s.createDecision(ctx, workObjectID, title, summary, rationale)
}

func (s *Store) UpdateDecision(ctx context.Context, decisionID string, workObjectID string, title string, summary string, rationale string) (domain.Decision, error) {
	workObjectID = strings.TrimSpace(workObjectID)
	title = strings.TrimSpace(title)
	summary = strings.TrimSpace(summary)
	rationale = strings.TrimSpace(rationale)
	if workObjectID == "" || title == "" || summary == "" || rationale == "" {
		return domain.Decision{}, ErrInvalidDecisionFields
	}

	if _, err := s.lookupWorkObject(ctx, workObjectID); err != nil {
		return domain.Decision{}, err
	}

	result, err := s.DB.ExecContext(ctx, updateDecisionStatement, workObjectID, title, summary, rationale, strings.TrimSpace(decisionID))
	if err != nil {
		return domain.Decision{}, fmt.Errorf("update decision: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return domain.Decision{}, fmt.Errorf("update decision rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return domain.Decision{}, sql.ErrNoRows
	}

	return s.LookupDecision(ctx, decisionID)
}

func (s *Store) DeleteDecision(ctx context.Context, decisionID string) error {
	result, err := s.DB.ExecContext(ctx, deleteDecisionStatement, strings.TrimSpace(decisionID))
	if err != nil {
		return fmt.Errorf("delete decision: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete decision rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *Store) createDecision(ctx context.Context, workObjectID string, title string, summary string, rationale string) (domain.Decision, error) {
	workObjectID = strings.TrimSpace(workObjectID)
	title = strings.TrimSpace(title)
	summary = strings.TrimSpace(summary)
	rationale = strings.TrimSpace(rationale)
	if workObjectID == "" || title == "" || summary == "" || rationale == "" {
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
		insertDecisionStatement,
		decision.ID,
		decision.WorkObjectID,
		decision.Title,
		decision.Summary,
		decision.Rationale,
		decision.CreatedAt.Format(time.RFC3339),
	); err != nil {
		return domain.Decision{}, fmt.Errorf("insert decision: %w", err)
	}

	return decision, nil
}

type decisionScanner interface {
	Scan(dest ...any) error
}

func scanDecision(scanner decisionScanner) (domain.Decision, error) {
	var decision domain.Decision
	var createdAt string
	if err := scanner.Scan(
		&decision.ID,
		&decision.WorkObjectID,
		&decision.Title,
		&decision.Summary,
		&decision.Rationale,
		&createdAt,
	); err != nil {
		return domain.Decision{}, fmt.Errorf("scan decision: %w", err)
	}

	var err error
	decision.CreatedAt, err = decodeTime(createdAt)
	if err != nil {
		return domain.Decision{}, fmt.Errorf("decode decision time: %w", err)
	}
	return decision, nil
}
