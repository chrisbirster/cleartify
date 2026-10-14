package database

import (
	"context"
	"fmt"
)

var schemaStatements = []string{
 `CREATE TABLE IF NOT EXISTS fake_payment_idempotency (
    actor_did TEXT NOT NULL,
    request_key TEXT NOT NULL,
    operation TEXT NOT NULL CHECK(operation IN ('purchase','refund')),
    target_id TEXT NOT NULL,
    result_id TEXT NOT NULL,
    created_at TEXT NOT NULL,
    PRIMARY KEY(actor_did,request_key)
 )`,
 `CREATE TABLE IF NOT EXISTS fake_seller_compliance_placeholders (
    seller_did TEXT PRIMARY KEY,
    payout_state TEXT NOT NULL DEFAULT 'unconfigured' CHECK(payout_state IN ('unconfigured','pending_review','blocked')),
    tax_data_state TEXT NOT NULL DEFAULT 'not_collected' CHECK(tax_data_state IN ('not_collected','pending_review','blocked')),
    updated_at TEXT NOT NULL
 )`,

 `CREATE TABLE IF NOT EXISTS agent_local_outbox (
  message_id TEXT PRIMARY KEY REFERENCES agent_interop_messages(message_id),
  state TEXT NOT NULL DEFAULT 'pending' CHECK(state IN ('pending','leased','delivered','dead')),
  attempts INTEGER NOT NULL DEFAULT 0 CHECK(attempts>=0),
  available_at TEXT NOT NULL,
  lease_token TEXT,
  lease_until TEXT,
  last_error_key TEXT,
  updated_at TEXT NOT NULL
 )`,
 `CREATE INDEX IF NOT EXISTS idx_agent_local_outbox_due ON agent_local_outbox(state,available_at,lease_until)`,

 `CREATE TABLE IF NOT EXISTS fake_seller_ledger (entitlement_id TEXT NOT NULL REFERENCES marketplace_entitlements(id), entry_kind TEXT NOT NULL CHECK(entry_kind IN ('sale','reversal')), seller_did TEXT NOT NULL, gross_minor INTEGER NOT NULL, fee_minor INTEGER NOT NULL, net_minor INTEGER NOT NULL, currency TEXT NOT NULL CHECK(currency='USD'), recorded_at TEXT NOT NULL, PRIMARY KEY(entitlement_id,entry_kind))`,
 `CREATE TABLE IF NOT EXISTS fake_payment_disputes (entitlement_id TEXT PRIMARY KEY REFERENCES marketplace_entitlements(id), requester_did TEXT NOT NULL, reason_key TEXT NOT NULL CHECK(reason_key IN ('not_received','not_as_described','other')), status TEXT NOT NULL CHECK(status IN ('open','resolved_refund','resolved_no_refund')), opened_at TEXT NOT NULL, decided_at TEXT)`,
 `CREATE TABLE IF NOT EXISTS fake_payment_dispute_audit (entitlement_id TEXT NOT NULL REFERENCES fake_payment_disputes(entitlement_id), event TEXT NOT NULL CHECK(event IN ('opened','resolved_refund','resolved_no_refund')), actor_did TEXT NOT NULL, recorded_at TEXT NOT NULL, PRIMARY KEY(entitlement_id,event))`,
 `CREATE TABLE IF NOT EXISTS creator_takedown_requests (id TEXT PRIMARY KEY, protocol_id TEXT NOT NULL REFERENCES published_protocols(id), requester_did TEXT NOT NULL, reason_key TEXT NOT NULL, status TEXT NOT NULL CHECK(status IN ('pending','hidden','rejected','restored')), created_at TEXT NOT NULL, decided_at TEXT)`,
 `CREATE UNIQUE INDEX IF NOT EXISTS idx_creator_takedown_open ON creator_takedown_requests(protocol_id) WHERE status IN ('pending','hidden')`,
 `CREATE TABLE IF NOT EXISTS creator_takedown_audit (id TEXT PRIMARY KEY, request_id TEXT NOT NULL REFERENCES creator_takedown_requests(id), action TEXT NOT NULL, actor_did TEXT NOT NULL, recorded_at TEXT NOT NULL)`,
 `CREATE TABLE IF NOT EXISTS creator_remix_lineage (protocol_id TEXT PRIMARY KEY REFERENCES published_protocols(id), parent_protocol_id TEXT NOT NULL REFERENCES published_protocols(id), attribution TEXT NOT NULL, recorded_at TEXT NOT NULL, CHECK(protocol_id<>parent_protocol_id))`,
 `CREATE TABLE IF NOT EXISTS signup_abuse_windows (identity_hash TEXT NOT NULL, bucket INTEGER NOT NULL, attempts INTEGER NOT NULL CHECK(attempts>=0), PRIMARY KEY(identity_hash,bucket))`,
 `CREATE TABLE IF NOT EXISTS signup_consent_keys (waitlist_entry_id TEXT NOT NULL REFERENCES waitlist_entries(id), consent_key TEXT NOT NULL, consent_version_key TEXT NOT NULL, accepted INTEGER NOT NULL CHECK(accepted IN (0,1)), recorded_at TEXT NOT NULL, PRIMARY KEY(waitlist_entry_id,consent_key,recorded_at))`,
 `CREATE TABLE IF NOT EXISTS agent_owner_challenges (nonce_hash TEXT PRIMARY KEY, agent_id TEXT NOT NULL REFERENCES agent_registrations(id), owner_did TEXT NOT NULL, public_key BLOB NOT NULL, expires_at TEXT NOT NULL, consumed_at TEXT)`,
 `CREATE TABLE IF NOT EXISTS agent_owner_audit (id TEXT PRIMARY KEY, agent_id TEXT NOT NULL REFERENCES agent_registrations(id), owner_did TEXT NOT NULL, action TEXT NOT NULL, recorded_at TEXT NOT NULL)`,
 `CREATE TABLE IF NOT EXISTS clear_fake_install_runs (id TEXT PRIMARY KEY, install_request_id TEXT NOT NULL UNIQUE REFERENCES protocol_install_requests(id), user_id TEXT NOT NULL REFERENCES users(id), actor_did TEXT NOT NULL, publication_id TEXT NOT NULL REFERENCES published_protocols(id), manifest_sha256 TEXT NOT NULL, status TEXT NOT NULL CHECK(status IN ('reserved','completed','failed')), result_json TEXT NOT NULL DEFAULT '{}', created_at TEXT NOT NULL, finished_at TEXT)`,
 `CREATE INDEX IF NOT EXISTS idx_clear_fake_install_user ON clear_fake_install_runs(user_id,created_at)`,
 `CREATE TABLE IF NOT EXISTS clear_fake_install_audit (id TEXT PRIMARY KEY, run_id TEXT NOT NULL REFERENCES clear_fake_install_runs(id), action TEXT NOT NULL, recorded_at TEXT NOT NULL)`,
 `CREATE TABLE IF NOT EXISTS creator_publications (protocol_id TEXT PRIMARY KEY REFERENCES published_protocols(id), visibility TEXT NOT NULL CHECK(visibility IN ('private','unlisted','public')), original_source_id TEXT, provenance_note TEXT NOT NULL, approved_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
 `CREATE TABLE IF NOT EXISTS creator_publication_audit (id TEXT PRIMARY KEY, protocol_id TEXT NOT NULL REFERENCES published_protocols(id), actor_did TEXT NOT NULL, from_visibility TEXT NOT NULL, to_visibility TEXT NOT NULL, recorded_at TEXT NOT NULL)`,
 `CREATE TABLE IF NOT EXISTS social_moderation_cases (id TEXT PRIMARY KEY, subject_did TEXT NOT NULL REFERENCES social_profiles(did), kind TEXT NOT NULL CHECK(kind IN ('review','appeal')), status TEXT NOT NULL CHECK(status IN ('pending','upheld','released','rejected')), rationale TEXT NOT NULL, created_at TEXT NOT NULL, decided_at TEXT, decision_reason TEXT)`,
 `CREATE UNIQUE INDEX IF NOT EXISTS idx_social_moderation_pending ON social_moderation_cases(subject_did,kind) WHERE status='pending'`,
 `CREATE TABLE IF NOT EXISTS social_moderation_audit (id TEXT PRIMARY KEY, case_id TEXT NOT NULL REFERENCES social_moderation_cases(id), actor TEXT NOT NULL, action TEXT NOT NULL, recorded_at TEXT NOT NULL)`,
 `CREATE TABLE IF NOT EXISTS social_impersonation_holds (
 subject_did TEXT PRIMARY KEY REFERENCES social_profiles(did),
 reason TEXT NOT NULL CHECK(reason IN ('reserved','lookalike','reports')),
 created_at TEXT NOT NULL,
 released_at TEXT
 )`,

 `CREATE TABLE IF NOT EXISTS marketplace_entitlements (
 id TEXT PRIMARY KEY,
 listing_id TEXT NOT NULL REFERENCES marketplace_listings(id),
 buyer_did TEXT NOT NULL,
 provider TEXT NOT NULL CHECK(provider='fake'),
 provider_receipt TEXT NOT NULL UNIQUE,
 status TEXT NOT NULL CHECK(status IN ('active','refunded')),
 created_at TEXT NOT NULL,
 refunded_at TEXT,
 UNIQUE(listing_id,buyer_did)
 )`,
 `CREATE TABLE IF NOT EXISTS marketplace_payment_audit (
 id TEXT PRIMARY KEY,
 entitlement_id TEXT NOT NULL REFERENCES marketplace_entitlements(id),
 event TEXT NOT NULL CHECK(event IN ('fake_purchase','fake_refund')),
 actor_did TEXT NOT NULL,
 recorded_at TEXT NOT NULL
 )`,

 `CREATE TABLE IF NOT EXISTS agent_message_consents (
 sender_agent_id TEXT NOT NULL REFERENCES agent_registrations(id),
 recipient_agent_id TEXT NOT NULL REFERENCES agent_registrations(id),
 recipient_owner_did TEXT NOT NULL,
 granted_at TEXT NOT NULL,
 expires_at TEXT NOT NULL,
 revoked_at TEXT,
 PRIMARY KEY(sender_agent_id,recipient_agent_id)
 )`,

 `CREATE TABLE IF NOT EXISTS agent_owner_bindings (
 agent_id TEXT PRIMARY KEY REFERENCES agent_registrations(id),
 owner_did TEXT NOT NULL,
 agent_public_key BLOB NOT NULL,
 proof_signature BLOB NOT NULL,
 verified_at TEXT NOT NULL,
 revoked_at TEXT
 )`,
 `CREATE TABLE IF NOT EXISTS agent_scoped_grants (
 id TEXT PRIMARY KEY,
 owner_did TEXT NOT NULL,
 agent_id TEXT NOT NULL REFERENCES agent_registrations(id),
 grantee_agent_id TEXT NOT NULL REFERENCES agent_registrations(id),
 scope TEXT NOT NULL,
 created_at TEXT NOT NULL,
 expires_at TEXT NOT NULL,
 revoked_at TEXT,
 CHECK(agent_id<>grantee_agent_id)
 )`,
 `CREATE INDEX IF NOT EXISTS idx_agent_grants_active ON agent_scoped_grants(agent_id,grantee_agent_id,scope,expires_at,revoked_at)`,

 `CREATE TABLE IF NOT EXISTS social_blocks (
  actor_did TEXT NOT NULL,
  subject_did TEXT NOT NULL,
  created_at TEXT NOT NULL,
  PRIMARY KEY(actor_did,subject_did),
  CHECK(actor_did<>subject_did)
 )`,
 `CREATE INDEX IF NOT EXISTS idx_social_blocks_subject ON social_blocks(subject_did)`,
 `CREATE TABLE IF NOT EXISTS social_reports (
  id TEXT PRIMARY KEY,
  reporter_did TEXT NOT NULL,
  subject_did TEXT NOT NULL,
  reason TEXT NOT NULL CHECK(reason IN ('spam','harassment','impersonation','other')),
  status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending','reviewed','dismissed')),
  created_at TEXT NOT NULL,
  CHECK(reporter_did<>subject_did)
 )`,
 `CREATE INDEX IF NOT EXISTS idx_social_reports_subject ON social_reports(subject_did,status)`,
 `CREATE TABLE IF NOT EXISTS clear_user_sessions (
   token_digest TEXT PRIMARY KEY,
   user_id TEXT NOT NULL REFERENCES users(id),
   created_at TEXT NOT NULL,
   last_seen_at TEXT NOT NULL,
   expires_at TEXT NOT NULL,
   revoked_at TEXT
 )`,
 `CREATE INDEX IF NOT EXISTS idx_clear_user_sessions_user ON clear_user_sessions(user_id, revoked_at)`,

 `CREATE TABLE IF NOT EXISTS identity_links (
  id TEXT PRIMARY KEY,
  clear_user_id TEXT NOT NULL REFERENCES users(id),
  social_did TEXT NOT NULL,
  status TEXT NOT NULL CHECK(status IN ('active','revoked')),
  consent_version TEXT NOT NULL,
  consented_at TEXT NOT NULL,
  authn_clear_at TEXT NOT NULL,
  authn_social_at TEXT NOT NULL,
  linked_at TEXT NOT NULL,
  revoked_at TEXT
 )`,
 `CREATE UNIQUE INDEX IF NOT EXISTS idx_identity_links_active_user ON identity_links(clear_user_id) WHERE status='active'`,
 `CREATE UNIQUE INDEX IF NOT EXISTS idx_identity_links_active_did ON identity_links(social_did) WHERE status='active'`,
 `CREATE TABLE IF NOT EXISTS identity_link_challenges (
  nonce_hash TEXT PRIMARY KEY,
  clear_user_id TEXT NOT NULL REFERENCES users(id),
  social_did TEXT NOT NULL,
  state TEXT NOT NULL CHECK(state IN ('pending','consumed','cancelled')),
  consent_version TEXT NOT NULL,
  expires_at TEXT NOT NULL,
  attempts INTEGER NOT NULL DEFAULT 0 CHECK(attempts BETWEEN 0 AND 5),
  consumed_at TEXT
 )`,
 `CREATE INDEX IF NOT EXISTS idx_identity_link_challenges_expires ON identity_link_challenges(expires_at)`,
 `CREATE TABLE IF NOT EXISTS identity_link_events (
  id TEXT PRIMARY KEY,
  clear_user_id TEXT NOT NULL,
  social_did TEXT NOT NULL,
  action TEXT NOT NULL CHECK(action IN ('linked','unlinked','rejected')),
  recorded_at TEXT NOT NULL
 )`,

 `CREATE TABLE IF NOT EXISTS clear_user_sessions (
   token_digest TEXT PRIMARY KEY,
   user_id TEXT NOT NULL REFERENCES users(id),
   created_at TEXT NOT NULL,
   last_seen_at TEXT NOT NULL,
   expires_at TEXT NOT NULL,
   revoked_at TEXT
 )`,
 `CREATE INDEX IF NOT EXISTS idx_clear_user_sessions_user ON clear_user_sessions(user_id, revoked_at)`,
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
	`CREATE TABLE IF NOT EXISTS policy_acceptances (id TEXT PRIMARY KEY, waitlist_entry_id TEXT NOT NULL, policy_version TEXT NOT NULL, accepted_at TEXT NOT NULL, FOREIGN KEY(waitlist_entry_id) REFERENCES waitlist_entries(id))`,
	`CREATE INDEX IF NOT EXISTS idx_policy_acceptances_waitlist ON policy_acceptances(waitlist_entry_id, accepted_at)`,
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
	`CREATE TABLE IF NOT EXISTS marketplace_listings (
        id TEXT PRIMARY KEY,
        protocol_id TEXT NOT NULL REFERENCES published_protocols(id),
        creator_did TEXT NOT NULL,
        status TEXT NOT NULL CHECK(status IN ('draft','listed','unlisted')) DEFAULT 'draft',
        title TEXT NOT NULL,
        currency TEXT NOT NULL CHECK(currency IN ('USD')),
        price_minor INTEGER NOT NULL CHECK(price_minor >= 0 AND price_minor <= 1000000),
        pricing_kind TEXT NOT NULL CHECK(pricing_kind IN ('free','one_time')),
        created_at TEXT NOT NULL,
        UNIQUE(protocol_id)
    )`,
    `CREATE INDEX IF NOT EXISTS idx_marketplace_listings_status ON marketplace_listings(status,created_at DESC)`,
	`CREATE TABLE IF NOT EXISTS social_default_follow_optouts (follower_did TEXT NOT NULL, subject_did TEXT NOT NULL, opted_out_at TEXT NOT NULL, PRIMARY KEY(follower_did, subject_did))`,
	`CREATE TABLE IF NOT EXISTS protocol_install_requests (
        id TEXT PRIMARY KEY,
        actor_did TEXT NOT NULL,
        publication_id TEXT NOT NULL REFERENCES published_protocols(id),
        protocol_version INTEGER NOT NULL CHECK(protocol_version>0),
        spec_sha256 TEXT NOT NULL,
        manifest_json TEXT NOT NULL,
        manifest_sha256 TEXT NOT NULL,
        acknowledged_at TEXT NOT NULL,
        status TEXT NOT NULL DEFAULT 'pending_approval' CHECK(status IN ('pending_approval','approved','rejected')),
        requested_at TEXT NOT NULL,
        decided_at TEXT
    )`,
    `CREATE INDEX IF NOT EXISTS idx_protocol_install_requests_actor ON protocol_install_requests(actor_did,status)`,
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
