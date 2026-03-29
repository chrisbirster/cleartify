package store

import (
	"context"
	"database/sql"
	"fmt"

	sqldata "cleartify/data"

	"cleartify/internal/domain"
)

var listSignalsStatement = sqldata.MustQuery("list_signals.sql")

func (s *Store) ListSignals(ctx context.Context) ([]domain.Signal, error) {
	rows, err := s.DB.QueryContext(ctx, listSignalsStatement)
	if err != nil {
		return nil, fmt.Errorf("list signals: %w", err)
	}
	defer rows.Close()

	var signals []domain.Signal
	for rows.Next() {
		var signal domain.Signal
		var rawMetadata sql.NullString
		var createdAt string
		if err := rows.Scan(
			&signal.ID,
			&signal.SourceType,
			&signal.SourceID,
			&signal.Title,
			&signal.Body,
			&signal.URL,
			&rawMetadata,
			&createdAt,
		); err != nil {
			return nil, fmt.Errorf("scan signal: %w", err)
		}

		signal.Metadata, err = decodeJSONMap(rawMetadata)
		if err != nil {
			return nil, fmt.Errorf("decode signal metadata: %w", err)
		}
		signal.CreatedAt, err = decodeTime(createdAt)
		if err != nil {
			return nil, fmt.Errorf("decode signal time: %w", err)
		}
		signals = append(signals, signal)
	}

	return signals, rows.Err()
}
