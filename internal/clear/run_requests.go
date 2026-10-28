package clear

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"github.com/chrisbirster/cleartify/internal/protocols"
	"time"
)

var ErrNotPublished = errors.New("protocol is not published or installed")
var ErrNotPending = errors.New("run request is not pending or no longer authorized")

type Service struct {
	DB             *sql.DB
	Push           func(context.Context) error
	HoldVisibility bool
}

// Only a canonical publication ID can authorize use. Historical slug-only
// installations carry no creator proof and must be explicitly reinstalled.
func installedSpec(ctx context.Context, tx *sql.Tx, user, id string, holdVisibility bool) (string, error) {
	access := protocols.AccessiblePublicationSQL
	if holdVisibility {
		access += " AND " + protocols.PublicationHoldSQL
	}
	var spec, digest string
	err := tx.QueryRowContext(ctx, `SELECT published_protocols.spec_json,i.spec_sha256 FROM published_protocols
 JOIN protocol_installations i ON i.publication_id=published_protocols.id AND i.version=published_protocols.version
 WHERE published_protocols.id=? AND i.user_id=? AND i.enabled=1
 AND NOT EXISTS(SELECT 1 FROM identity_links l JOIN social_blocks b ON
 ((b.actor_did=l.social_did AND b.subject_did=published_protocols.creator_did) OR (b.actor_did=published_protocols.creator_did AND b.subject_did=l.social_did))
 WHERE l.clear_user_id=i.user_id AND l.status='active') AND `+access+` LIMIT 1`, id, user).Scan(&spec, &digest)
	if err == nil && (digest == "" || digest != specDigest(spec)) {
		return "", sql.ErrNoRows
	}
	return spec, err
}
func specDigest(spec string) string {
	h := sha256.Sum256([]byte(spec))
	return hex.EncodeToString(h[:])
}
func (s Service) Request(ctx context.Context, userID, publicationID string) (string, error) {
	if userID == "" || publicationID == "" {
		return "", ErrNotPublished
	}
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	spec, err := installedSpec(ctx, tx, userID, publicationID, s.HoldVisibility)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotPublished
	}
	if err != nil {
		return "", err
	}
	idBytes := make([]byte, 18)
	if _, err = rand.Read(idBytes); err != nil {
		return "", err
	}
	id := hex.EncodeToString(idBytes)
	_, err = tx.ExecContext(ctx, `INSERT INTO clear_run_requests(id,user_id,publication_id,status,requested_at,spec_sha256) VALUES(?,?,?,'pending_approval',?,?)`, id, userID, publicationID, time.Now().UTC().Format(time.RFC3339Nano), specDigest(spec))
	if err != nil {
		return "", err
	}
	if err = tx.Commit(); err != nil {
		return "", err
	}
	if s.Push != nil {
		if err = s.Push(ctx); err != nil {
			return "", err
		}
	}
	return id, nil
}
func (s Service) Decide(ctx context.Context, userID, requestID string, approve bool) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var pub, digest string
	err = tx.QueryRowContext(ctx, `SELECT publication_id,spec_sha256 FROM clear_run_requests WHERE id=? AND user_id=? AND status='pending_approval'`, requestID, userID).Scan(&pub, &digest)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotPending
	}
	if err != nil {
		return err
	}
	if approve {
		spec, e := installedSpec(ctx, tx, userID, pub, s.HoldVisibility)
		if errors.Is(e, sql.ErrNoRows) {
			return ErrNotPending
		}
		if e != nil {
			return e
		}
		if digest == "" || digest != specDigest(spec) {
			return ErrNotPending
		}
	}
	status := "rejected"
	if approve {
		status = "approved"
	}
	result, err := tx.ExecContext(ctx, `UPDATE clear_run_requests SET status=?,decided_at=? WHERE id=? AND user_id=? AND status='pending_approval'`, status, time.Now().UTC().Format(time.RFC3339Nano), requestID, userID)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return ErrNotPending
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	if s.Push != nil {
		return s.Push(ctx)
	}
	return nil
}
