package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// The existing bootstrap remains the baseline. New schema changes are applied
// transactionally and recorded so an existing database receives each upgrade.
var migrations = []struct {
	version    int
	statements []string
}{
	{1, []string{
		`CREATE TABLE beta_invite_identities (
			invite_id TEXT PRIMARY KEY REFERENCES beta_invites(id),
			name TEXT NOT NULL,
			email_normalized TEXT NOT NULL,
			phone_e164 TEXT NOT NULL
		)`,
		`CREATE TRIGGER beta_invite_identity_no_update
		 BEFORE UPDATE ON beta_invite_identities
		 BEGIN SELECT RAISE(ABORT,'immutable invite identity'); END`,
	}},
	{2, []string{`ALTER TABLE clear_run_requests ADD COLUMN spec_sha256 TEXT NOT NULL DEFAULT ''`}},
	{3, []string{
		`ALTER TABLE protocol_installations ADD COLUMN publication_id TEXT REFERENCES published_protocols(id)`,
		`ALTER TABLE protocol_installations ADD COLUMN spec_sha256 TEXT NOT NULL DEFAULT ''`,
	}},
}

func (s *Store) migrateVersions(ctx context.Context) error {
	if _, err := s.DB.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY, applied_at TEXT NOT NULL
	)`); err != nil {
		return err
	}
	var latest int
	if err := s.DB.QueryRowContext(ctx, `SELECT COALESCE(MAX(version),0) FROM schema_migrations`).Scan(&latest); err != nil {
		return err
	}
	if latest > migrations[len(migrations)-1].version {
		return fmt.Errorf("database schema version %d is newer than this application", latest)
	}
	for _, migration := range migrations {
		tx, err := s.DB.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		err = func() error {
			defer tx.Rollback()
			var version int
			err := tx.QueryRowContext(ctx, `SELECT version FROM schema_migrations WHERE version=?`, migration.version).Scan(&version)
			if err == nil {
				return nil
			}
			if !errors.Is(err, sql.ErrNoRows) {
				return err
			}
			for _, statement := range migration.statements {
				if _, err = tx.ExecContext(ctx, statement); err != nil {
					return err
				}
			}
			if _, err = tx.ExecContext(ctx, `INSERT INTO schema_migrations(version,applied_at) VALUES(?,?)`, migration.version, time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
				return err
			}
			return tx.Commit()
		}()
		if err != nil {
			return fmt.Errorf("apply migration %d: %w", migration.version, err)
		}
	}
	return nil
}
