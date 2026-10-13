package database

import (
	"context"
	"fmt"
)

var schemaStatements = []string{
	`CREATE TABLE IF NOT EXISTS waitlist_entries (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		email TEXT NOT NULL,
		email_normalized TEXT NOT NULL UNIQUE,
		phone_e164 TEXT NOT NULL UNIQUE,
		email_updates INTEGER NOT NULL DEFAULT 0 CHECK (email_updates IN (0, 1)),
		status TEXT NOT NULL DEFAULT 'waiting',
		source TEXT NOT NULL DEFAULT 'landing',
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL
	)`,
	`CREATE INDEX IF NOT EXISTS idx_waitlist_status_created
		ON waitlist_entries(status, created_at)`,
	`CREATE TABLE IF NOT EXISTS consents (
		id TEXT PRIMARY KEY,
		waitlist_entry_id TEXT NOT NULL,
		kind TEXT NOT NULL,
		granted INTEGER NOT NULL CHECK (granted IN (0, 1)),
		source TEXT NOT NULL,
		recorded_at TEXT NOT NULL,
		FOREIGN KEY(waitlist_entry_id) REFERENCES waitlist_entries(id)
	)`,
	`CREATE TABLE IF NOT EXISTS policy_acceptances (
		id TEXT PRIMARY KEY,
		waitlist_entry_id TEXT NOT NULL,
		policy_version TEXT NOT NULL,
		accepted_at TEXT NOT NULL,
		FOREIGN KEY(waitlist_entry_id) REFERENCES waitlist_entries(id)
	)`,
	`CREATE INDEX IF NOT EXISTS idx_policy_acceptances_waitlist ON policy_acceptances(waitlist_entry_id, accepted_at)`,
	`CREATE INDEX IF NOT EXISTS idx_consents_waitlist
		ON consents(waitlist_entry_id, recorded_at)`,
}

func (s *Store) Migrate(ctx context.Context) error {
	for i, stmt := range schemaStatements {
		if _, err := s.DB.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("apply schema statement %d: %w", i+1, err)
		}
	}
	return s.Push(ctx)
}
