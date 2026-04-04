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

var listSignalsStatement = sqldata.MustQuery("list_signals.sql")
var lookupSignalStatement = sqldata.MustQuery("lookup_signal.sql")
var insertSignalStatement = sqldata.MustQuery("insert_signal.sql")
var updateSignalStatement = sqldata.MustQuery("update_signal.sql")
var deleteClusterSignalLinksForSignalStatement = sqldata.MustQuery("delete_cluster_signal_links_for_signal.sql")
var deleteSignalStatement = sqldata.MustQuery("delete_signal.sql")
var signalNow = time.Now
var signalIDFromTime = func(createdAt time.Time) string {
	return fmt.Sprintf("sig_%d", createdAt.UTC().UnixNano())
}

var ErrInvalidSignalFields = errors.New("invalid signal fields")

func (s *Store) ListSignals(ctx context.Context) ([]domain.Signal, error) {
	rows, err := s.DB.QueryContext(ctx, listSignalsStatement)
	if err != nil {
		return nil, fmt.Errorf("list signals: %w", err)
	}
	defer rows.Close()

	var signals []domain.Signal
	for rows.Next() {
		signal, err := scanSignal(rows)
		if err != nil {
			return nil, err
		}
		signals = append(signals, signal)
	}

	return signals, rows.Err()
}

func (s *Store) LookupSignal(ctx context.Context, signalID string) (domain.Signal, error) {
	row := s.DB.QueryRowContext(ctx, lookupSignalStatement, strings.TrimSpace(signalID))
	signal, err := scanSignal(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Signal{}, sql.ErrNoRows
		}
		return domain.Signal{}, fmt.Errorf("lookup signal: %w", err)
	}
	return signal, nil
}

func (s *Store) CreateSignal(ctx context.Context, sourceType string, sourceID string, title string, body string, externalURL string) (domain.Signal, error) {
	sourceType, sourceID, title, body, externalURL, err := normalizeSignalFields(sourceType, sourceID, title, body, externalURL)
	if err != nil {
		return domain.Signal{}, err
	}

	createdAt := signalNow().UTC().Truncate(time.Second)
	signal := domain.Signal{
		ID:         signalIDFromTime(createdAt),
		SourceType: sourceType,
		SourceID:   sourceID,
		Title:      title,
		Body:       body,
		URL:        externalURL,
		CreatedAt:  createdAt,
	}

	if _, err := s.DB.ExecContext(
		ctx,
		insertSignalStatement,
		signal.ID,
		nil,
		signal.SourceType,
		signal.SourceID,
		signal.Title,
		signal.Body,
		signal.URL,
		nil,
		signal.CreatedAt.Format(time.RFC3339),
	); err != nil {
		return domain.Signal{}, fmt.Errorf("insert signal: %w", err)
	}

	return signal, nil
}

func (s *Store) UpdateSignal(ctx context.Context, signalID string, sourceType string, sourceID string, title string, body string, externalURL string) (domain.Signal, error) {
	sourceType, sourceID, title, body, externalURL, err := normalizeSignalFields(sourceType, sourceID, title, body, externalURL)
	if err != nil {
		return domain.Signal{}, err
	}

	result, err := s.DB.ExecContext(ctx, updateSignalStatement, sourceType, sourceID, title, body, externalURL, nil, strings.TrimSpace(signalID))
	if err != nil {
		return domain.Signal{}, fmt.Errorf("update signal: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return domain.Signal{}, fmt.Errorf("update signal rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return domain.Signal{}, sql.ErrNoRows
	}

	return s.LookupSignal(ctx, signalID)
}

func (s *Store) DeleteSignal(ctx context.Context, signalID string) error {
	if _, err := s.DB.ExecContext(ctx, deleteClusterSignalLinksForSignalStatement, strings.TrimSpace(signalID)); err != nil {
		return fmt.Errorf("delete cluster signal links: %w", err)
	}
	result, err := s.DB.ExecContext(ctx, deleteSignalStatement, strings.TrimSpace(signalID))
	if err != nil {
		return fmt.Errorf("delete signal: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete signal rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

type signalScanner interface {
	Scan(dest ...any) error
}

func scanSignal(scanner signalScanner) (domain.Signal, error) {
	var signal domain.Signal
	var rawMetadata sql.NullString
	var createdAt string
	if err := scanner.Scan(
		&signal.ID,
		&signal.ProjectID,
		&signal.SourceType,
		&signal.SourceID,
		&signal.Title,
		&signal.Body,
		&signal.URL,
		&rawMetadata,
		&createdAt,
	); err != nil {
		return domain.Signal{}, fmt.Errorf("scan signal: %w", err)
	}

	var err error
	signal.Metadata, err = decodeJSONMap(rawMetadata)
	if err != nil {
		return domain.Signal{}, fmt.Errorf("decode signal metadata: %w", err)
	}
	signal.CreatedAt, err = decodeTime(createdAt)
	if err != nil {
		return domain.Signal{}, fmt.Errorf("decode signal time: %w", err)
	}
	return signal, nil
}

func normalizeSignalFields(sourceType string, sourceID string, title string, body string, externalURL string) (string, string, string, string, string, error) {
	normalizedSourceType, err := links.NormalizeSignalSourceType(sourceType)
	if err != nil {
		return "", "", "", "", "", ErrInvalidSignalFields
	}
	normalizedSourceID := strings.TrimSpace(sourceID)
	normalizedTitle := strings.TrimSpace(title)
	normalizedBody := strings.TrimSpace(body)
	normalizedURL, err := links.NormalizeExternalURL(externalURL)
	if err != nil {
		return "", "", "", "", "", ErrInvalidSignalFields
	}

	if normalizedSourceID == "" || normalizedTitle == "" || normalizedBody == "" {
		return "", "", "", "", "", ErrInvalidSignalFields
	}
	return normalizedSourceType, normalizedSourceID, normalizedTitle, normalizedBody, normalizedURL, nil
}
