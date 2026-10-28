package agents

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"github.com/chrisbirster/cleartify/internal/policy"
	"time"
)

var ErrOwnershipCeremony = errors.New("agent ownership verification denied")

type OwnershipVerifier interface {
	Verify(context.Context, []byte, string, string) bool
}

// OwnershipProofMessage binds a domain, agent, authenticated owner, candidate
// Ed25519 public key and single-use nonce. The private key never reaches server.
func OwnershipProofMessage(agent, owner, nonce string, key []byte) []byte {
	return []byte("cleartify-agent-ownership-v2\\n" + agent + "\\n" + owner + "\\n" + hex.EncodeToString(key) + "\\n" + nonce)
}

type Ed25519OwnershipVerifier struct{}

func (Ed25519OwnershipVerifier) Verify(_ context.Context, key []byte, challenge, proof string) bool {
	// challenge stores the complete contextual signing message as hex.
	message, err := hex.DecodeString(challenge)
	if err != nil || len(key) != ed25519.PublicKeySize {
		return false
	}
	sig, err := hex.DecodeString(proof)
	return err == nil && len(sig) == ed25519.SignatureSize && ed25519.Verify(ed25519.PublicKey(key), message, sig)
}

type HostAdapter interface {
	Challenge(context.Context, string, string) (string, error)
	Health(context.Context, string) (bool, error)
}
type LocalTestAdapter struct{}

func (LocalTestAdapter) Challenge(_ context.Context, agent, nonce string) (string, error) {
	if agent == "" || nonce == "" {
		return "", ErrOwnershipCeremony
	}
	return "local-test:" + agent + ":" + nonce, nil
}
func (LocalTestAdapter) Health(_ context.Context, id string) (bool, error) { return id != "", nil }

type Ownership struct {
	DB       *sql.DB
	Push     func(context.Context) error
	Verifier OwnershipVerifier
}

func randomHex() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
func (o Ownership) Start(ctx context.Context, actor, agentID string, key []byte) (string, error) {
	if o.DB == nil || o.Verifier == nil || o.Push == nil || len(key) != 32 || len(agentID) == 0 || len(actor) < 5 {
		return "", ErrOwnershipCeremony
	}
	var owner, status string
	if err := o.DB.QueryRowContext(ctx, "SELECT owner_did,status FROM agent_registrations WHERE id=?", agentID).Scan(&owner, &status); err != nil || owner != actor || status == "disabled" {
		return "", ErrOwnershipCeremony
	}
	raw, err := randomHex()
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256([]byte(raw))
	expiry := time.Now().UTC().Add(policy.AgentOwnershipChallengeTTL).Format(time.RFC3339Nano)
	_, err = o.DB.ExecContext(ctx, `INSERT INTO agent_owner_challenges(nonce_hash,agent_id,owner_did,public_key,expires_at) VALUES(?,?,?,?,?)`, hex.EncodeToString(digest[:]), agentID, actor, key, expiry)
	if err != nil {
		return "", err
	}
	if err = o.Push(ctx); err != nil {
		return "", err
	}
	return raw, nil
}
func (o Ownership) Confirm(ctx context.Context, actor, agentID, nonce, proof string) error {
	if o.DB == nil || o.Push == nil || o.Verifier == nil || len(nonce) != 64 || len(proof) != 128 {
		return ErrOwnershipCeremony
	}
	digest := sha256.Sum256([]byte(nonce))
	hash := hex.EncodeToString(digest[:])
	tx, err := o.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var owner, expires string
	var key []byte
	var consumed sql.NullString
	err = tx.QueryRowContext(ctx, `SELECT owner_did,public_key,expires_at,consumed_at FROM agent_owner_challenges WHERE nonce_hash=? AND agent_id=?`, hash, agentID).Scan(&owner, &key, &expires, &consumed)
	if err != nil || owner != actor || consumed.Valid {
		return ErrOwnershipCeremony
	}
	deadline, err := time.Parse(time.RFC3339Nano, expires)
	if err != nil || !time.Now().UTC().Before(deadline) {
		return ErrOwnershipCeremony
	}
	var liveOwner, status string
	if err = tx.QueryRowContext(ctx, "SELECT owner_did,status FROM agent_registrations WHERE id=?", agentID).Scan(&liveOwner, &status); err != nil || liveOwner != actor || status == "disabled" {
		return ErrOwnershipCeremony
	}
	if !o.Verifier.Verify(ctx, key, hex.EncodeToString(OwnershipProofMessage(agentID, actor, nonce, key)), proof) {
		return ErrOwnershipCeremony
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	result, err := tx.ExecContext(ctx, `UPDATE agent_owner_challenges SET consumed_at=? WHERE nonce_hash=? AND consumed_at IS NULL`, now, hash)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil || n != 1 {
		return ErrOwnershipCeremony
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO agent_owner_bindings(agent_id,owner_did,agent_public_key,proof_signature,verified_at,revoked_at) VALUES(?,?,?,?,?,NULL) ON CONFLICT(agent_id) DO UPDATE SET owner_did=excluded.owner_did,agent_public_key=excluded.agent_public_key,proof_signature=excluded.proof_signature,verified_at=excluded.verified_at,revoked_at=NULL`, agentID, actor, key, []byte(proof), now)
	if err != nil {
		return err
	}
	id, err := randomHex()
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO agent_owner_audit(id,agent_id,owner_did,action,recorded_at) VALUES(?,?,?,'ed25519_verified',?)`, id, agentID, actor, now)
	if err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return o.Push(ctx)
}

// Promote is an operator-reviewed transition, separate from proving key
// possession. The HTTP caller must independently authenticate the operator.
func (o Ownership) Promote(ctx context.Context, agentID string) error {
	if o.DB == nil || o.Push == nil {
		return ErrOwnershipCeremony
	}
	tx, err := o.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	res, err := tx.ExecContext(ctx, `UPDATE agent_registrations SET status='verified' WHERE id=? AND status='unverified'
 AND EXISTS(SELECT 1 FROM agent_owner_bindings b JOIN social_profiles p ON p.did=b.owner_did
 WHERE b.agent_id=agent_registrations.id AND b.owner_did=agent_registrations.owner_did AND b.revoked_at IS NULL
 AND p.claimed=1 AND p.public_visible=1 AND p.kind='person')`, agentID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return ErrOwnershipCeremony
	}
	id, err := randomHex()
	if err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO agent_owner_audit(id,agent_id,owner_did,action,recorded_at) SELECT ?,id,owner_did,'operator_activated',? FROM agent_registrations WHERE id=?`, id, time.Now().UTC().Format(time.RFC3339Nano), agentID); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return o.Push(ctx)
}
