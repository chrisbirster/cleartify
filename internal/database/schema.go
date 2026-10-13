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
	`CREATE INDEX IF NOT EXISTS idx_consents_waitlist
		ON consents(waitlist_entry_id, recorded_at)`,
	`CREATE TABLE IF NOT EXISTS beta_invites (
		id TEXT PRIMARY KEY,
		waitlist_entry_id TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'sent',
		email_token_hash TEXT NOT NULL UNIQUE,
		email_token_expires_at TEXT NOT NULL,
		email_token_used_at TEXT,
		activation_token_hash TEXT UNIQUE,
		activation_expires_at TEXT,
		challenge_code_hash TEXT,
		challenge_expires_at TEXT,
		challenge_sent_at TEXT,
		challenge_attempts INTEGER NOT NULL DEFAULT 0,
		max_challenge_attempts INTEGER NOT NULL DEFAULT 5,
		activated_user_id TEXT,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL,
		activated_at TEXT,
		FOREIGN KEY(waitlist_entry_id) REFERENCES waitlist_entries(id)
	)`,
	`CREATE INDEX IF NOT EXISTS idx_beta_invites_waitlist_status
		ON beta_invites(waitlist_entry_id, status)`,
	`CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		display_name TEXT NOT NULL,
		primary_email TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'active',
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL
	)`,
	`CREATE UNIQUE INDEX IF NOT EXISTS idx_users_primary_email
		ON users(primary_email)`,
	`CREATE TABLE IF NOT EXISTS phone_identities (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		phone_e164 TEXT NOT NULL UNIQUE,
		verified_at TEXT NOT NULL,
		created_at TEXT NOT NULL,
		FOREIGN KEY(user_id) REFERENCES users(id)
	)`,
	`CREATE INDEX IF NOT EXISTS idx_phone_identities_user
		ON phone_identities(user_id)`,
	`CREATE TABLE IF NOT EXISTS protocol_installations (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		protocol_key TEXT NOT NULL,
		version INTEGER NOT NULL,
		source TEXT NOT NULL,
		enabled INTEGER NOT NULL DEFAULT 1 CHECK (enabled IN (0, 1)),
		created_at TEXT NOT NULL,
		FOREIGN KEY(user_id) REFERENCES users(id),
		UNIQUE(user_id, protocol_key)
	)`,
	`CREATE TABLE IF NOT EXISTS atproto_oauth_requests (
		state TEXT PRIMARY KEY,
		payload_ciphertext BLOB NOT NULL,
		browser_nonce_hash TEXT NOT NULL,
		created_at TEXT NOT NULL,
		expires_at TEXT NOT NULL
	)`,
	`CREATE INDEX IF NOT EXISTS idx_atproto_oauth_requests_expires ON atproto_oauth_requests(expires_at)`,
	`CREATE TABLE IF NOT EXISTS atproto_oauth_sessions (
		account_did TEXT NOT NULL,
		session_id TEXT NOT NULL,
		payload_ciphertext BLOB NOT NULL,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL,
		PRIMARY KEY(account_did,session_id)
	)`,
	`CREATE TABLE IF NOT EXISTS atproto_browser_sessions (
		token_digest TEXT PRIMARY KEY,
		account_did TEXT NOT NULL,
		oauth_session_id TEXT NOT NULL,
		expires_at TEXT NOT NULL,
		created_at TEXT NOT NULL,
		FOREIGN KEY(account_did,oauth_session_id) REFERENCES atproto_oauth_sessions(account_did,session_id)
	)`,
	`CREATE INDEX IF NOT EXISTS idx_atproto_browser_sessions_expiry ON atproto_browser_sessions(expires_at)`,
	`CREATE TABLE IF NOT EXISTS published_protocols (
		id TEXT PRIMARY KEY,
		creator_did TEXT NOT NULL,
		slug TEXT NOT NULL,
		version INTEGER NOT NULL CHECK(version > 0),
		title TEXT NOT NULL,
		summary TEXT NOT NULL,
		license TEXT NOT NULL,
		spec_json TEXT NOT NULL,
		status TEXT NOT NULL CHECK(status IN ('draft','published','withdrawn')),
		created_at TEXT NOT NULL,
		UNIQUE(creator_did,slug,version)
	)`,
	`CREATE INDEX IF NOT EXISTS idx_protocols_discovery ON published_protocols(status,created_at DESC)`,
	`CREATE INDEX IF NOT EXISTS idx_protocol_installations_user
		ON protocol_installations(user_id, enabled)`,
}

func (s *Store) Migrate(ctx context.Context) error {
	for i, stmt := range schemaStatements {
		if _, err := s.DB.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("apply schema statement %d: %w", i+1, err)
		}
	}
	return s.Push(ctx)
}
