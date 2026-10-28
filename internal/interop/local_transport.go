package interop

import (
	"context"
	"crypto/ed25519"
	"database/sql"
	"errors"
	"time"
)

var ErrTransportDisabled = errors.New("local transport disabled")
var ErrTransportDenied = errors.New("transport authorization or consent denied")

type LocalTransport struct {
	enabled bool
	db      *sql.DB
	push    func(context.Context) error
	now     func() time.Time
}

func NewLocalTransport(enabled bool, db *sql.DB, push func(context.Context) error) *LocalTransport {
	return &LocalTransport{enabled: enabled, db: db, push: push, now: time.Now}
}

// Send verifies signature, signed identity, live owner bindings, scoped grant,
// recipient consent and anti-replay within ONE DB transaction. It records only
// a pending envelope: no network, listener, external adapter or executor.
func (t *LocalTransport) Send(ctx context.Context, e Envelope) error {
	if !t.enabled {
		return ErrTransportDisabled
	}
	if t.db == nil || t.push == nil {
		return ErrTransportDenied
	}
	now := t.now().UTC()
	tx, err := t.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var pub []byte
	err = tx.QueryRowContext(ctx, `SELECT k.public_key FROM agent_interop_keys k
 JOIN agent_owner_bindings sb ON sb.agent_id=k.agent_id AND sb.revoked_at IS NULL
 JOIN agent_registrations sa ON sa.id=k.agent_id AND sa.status='verified'
 JOIN agent_owner_bindings rb ON rb.agent_id=? AND rb.revoked_at IS NULL
 JOIN agent_registrations ra ON ra.id=rb.agent_id AND ra.status='verified'
 WHERE k.key_id=? AND k.agent_id=? AND k.algorithm='Ed25519' AND k.revoked_at IS NULL AND NOT EXISTS(SELECT 1 FROM social_blocks b WHERE (b.actor_did=sb.owner_did AND b.subject_did=rb.owner_did) OR (b.actor_did=rb.owner_did AND b.subject_did=sb.owner_did))`, e.RecipientAgentID, e.KeyID, e.SenderAgentID).Scan(&pub)
	if err != nil {
		return ErrTransportDenied
	}
	if err = VerifyEnvelope(e, ed25519.PublicKey(pub), now); err != nil {
		return err
	}
	var authorized int
	err = tx.QueryRowContext(ctx, `SELECT count(*) FROM agent_scoped_grants g
 JOIN agent_owner_bindings b ON b.agent_id=g.agent_id AND b.owner_did=g.owner_did AND b.revoked_at IS NULL
 JOIN agent_message_consents c ON c.sender_agent_id=g.agent_id AND c.recipient_agent_id=g.grantee_agent_id AND c.revoked_at IS NULL
 JOIN agent_owner_bindings rb ON rb.agent_id=c.recipient_agent_id AND rb.owner_did=c.recipient_owner_did AND rb.revoked_at IS NULL
 WHERE g.agent_id=? AND g.grantee_agent_id=? AND g.scope='message.preview' AND g.revoked_at IS NULL
 AND g.expires_at>? AND c.expires_at>?`, e.SenderAgentID, e.RecipientAgentID, now.Format(time.RFC3339Nano), now.Format(time.RFC3339Nano)).Scan(&authorized)
	if err != nil || authorized < 1 {
		return ErrTransportDenied
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO agent_interop_messages(message_id,sender_agent_id,recipient_agent_id,key_id,payload_digest,signature,issued_at,expires_at,status) VALUES(?,?,?,?,?,?,?,?,'pending')`, e.MessageID, e.SenderAgentID, e.RecipientAgentID, e.KeyID, e.PayloadDigest, e.Signature, e.IssuedAt, e.ExpiresAt)
	if err != nil {
		return ErrReplay
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return t.push(ctx)
}

// RevokeConsent binds the action to an already authenticated owner DID and
// checks recipient ownership in the same transaction as the revocation.
func (t *LocalTransport) RevokeConsent(ctx context.Context, owner, recipient, sender string) error {
	if !t.enabled {
		return ErrTransportDisabled
	}
	if t.db == nil || t.push == nil {
		return ErrTransportDenied
	}
	tx, err := t.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	now := t.now().UTC().Format(time.RFC3339Nano)
	res, err := tx.ExecContext(ctx, `UPDATE agent_message_consents SET revoked_at=?
 WHERE sender_agent_id=? AND recipient_agent_id=? AND recipient_owner_did=? AND revoked_at IS NULL
 AND EXISTS(SELECT 1 FROM agent_owner_bindings b JOIN agent_registrations a ON a.id=b.agent_id
 WHERE b.agent_id=? AND b.owner_did=? AND b.revoked_at IS NULL AND a.status='verified')`, now, sender, recipient, owner, recipient, owner)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil || n != 1 {
		return ErrTransportDenied
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return t.push(ctx)
}
