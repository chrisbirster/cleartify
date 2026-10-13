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
	`CREATE TABLE IF NOT EXISTS social_profiles (did TEXT PRIMARY KEY, handle TEXT NOT NULL UNIQUE, display_name TEXT NOT NULL, bio TEXT NOT NULL DEFAULT '', avatar_url TEXT NOT NULL DEFAULT '', kind TEXT NOT NULL CHECK(kind IN ('person','agent')), public_visible INTEGER NOT NULL DEFAULT 0 CHECK(public_visible IN (0,1)), claimed INTEGER NOT NULL DEFAULT 0 CHECK(claimed IN (0,1)), updated_at TEXT NOT NULL)`,
	`CREATE INDEX IF NOT EXISTS idx_social_profiles_visibility ON social_profiles(public_visible,handle)`,
	`CREATE TABLE IF NOT EXISTS social_follows (follower_did TEXT NOT NULL, subject_did TEXT NOT NULL, created_at TEXT NOT NULL, PRIMARY KEY(follower_did,subject_did), FOREIGN KEY(follower_did) REFERENCES social_profiles(did), FOREIGN KEY(subject_did) REFERENCES social_profiles(did), CHECK(follower_did <> subject_did))`,
	`CREATE INDEX IF NOT EXISTS idx_social_follows_subject ON social_follows(subject_did)`,
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
	`CREATE TABLE IF NOT EXISTS clear_run_requests (
        id TEXT PRIMARY KEY,
        user_id TEXT NOT NULL REFERENCES users(id),
        publication_id TEXT NOT NULL REFERENCES published_protocols(id),
        status TEXT NOT NULL DEFAULT 'pending_approval' CHECK(status IN ('pending_approval','approved','rejected','cancelled')),
        requested_at TEXT NOT NULL,
        decided_at TEXT,
        UNIQUE(id,user_id)
    )`,
    `CREATE INDEX IF NOT EXISTS idx_clear_run_requests_user ON clear_run_requests(user_id,status)`,
	`CREATE TABLE IF NOT EXISTS alpha_memberships (
       user_id TEXT PRIMARY KEY REFERENCES users(id),
       admitted_by TEXT NOT NULL,
       admitted_at TEXT NOT NULL,
       revoked_at TEXT
    )`,
    `CREATE INDEX IF NOT EXISTS idx_alpha_memberships_revoked ON alpha_memberships(revoked_at)`,
	`CREATE TABLE IF NOT EXISTS agent_registrations (
        id TEXT PRIMARY KEY,
        owner_did TEXT NOT NULL REFERENCES social_profiles(did),
        display_name TEXT NOT NULL,
        description TEXT NOT NULL DEFAULT '',
        status TEXT NOT NULL DEFAULT 'unverified' CHECK(status IN ('unverified','verified','disabled')),
        created_at TEXT NOT NULL,
        UNIQUE(owner_did,display_name)
    )`,
    `CREATE INDEX IF NOT EXISTS idx_agent_registrations_owner ON agent_registrations(owner_did,status)`,
	`CREATE TABLE IF NOT EXISTS agent_interop_keys (
       key_id TEXT PRIMARY KEY,
       agent_id TEXT NOT NULL,
       algorithm TEXT NOT NULL CHECK(algorithm='Ed25519'),
       public_key BLOB NOT NULL,
       created_at TEXT NOT NULL,
       revoked_at TEXT
    )`,
    `CREATE INDEX IF NOT EXISTS idx_agent_interop_keys_agent ON agent_interop_keys(agent_id,revoked_at)`,
    `CREATE TABLE IF NOT EXISTS agent_interop_messages (
       message_id TEXT PRIMARY KEY,
       sender_agent_id TEXT NOT NULL,
       recipient_agent_id TEXT NOT NULL,
       key_id TEXT NOT NULL REFERENCES agent_interop_keys(key_id),
       payload_digest TEXT NOT NULL,
       signature BLOB NOT NULL,
       issued_at TEXT NOT NULL,
       expires_at TEXT NOT NULL,
       status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending','rejected','accepted','expired'))
    )`,
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
