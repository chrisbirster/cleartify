package marketplace

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

var ErrIdempotencyConflict = errors.New("fake payment idempotency key conflict")

// Key reservation, fake side effects and result are committed atomically.
// FakeProvider is local-only; no external payment can run in this transaction.
type KeyedOperations struct {
	DB      *sql.DB
	Enabled bool
}

func (k KeyedOperations) run(ctx context.Context, actor, key, op, target string) (string, error) {
	if !k.Enabled || k.DB == nil || !strings.HasPrefix(actor, "did:") || len(key) < 8 || len(key) > 128 || target == "" {
		return "", ErrDenied
	}
	tx, e := k.DB.BeginTx(ctx, nil)
	if e != nil {
		return "", e
	}
	defer tx.Rollback()
	var oldOp, oldTarget, result string
	e = tx.QueryRowContext(ctx, `SELECT operation,target_id,result_id FROM fake_payment_idempotency WHERE actor_did=? AND request_key=?`, actor, key).Scan(&oldOp, &oldTarget, &result)
	if e != nil && !errors.Is(e, sql.ErrNoRows) {
		return "", e
	}
	if e == nil {
		if oldOp != op || oldTarget != target {
			return "", ErrIdempotencyConflict
		}
		if result != "" {
			return result, nil
		}
	} else {
		if _, e = tx.ExecContext(ctx, `INSERT INTO fake_payment_idempotency(actor_did,request_key,operation,target_id,result_id,created_at) VALUES(?,?,?,?,'',?)`, actor, key, op, target, time.Now().UTC().Format(time.RFC3339Nano)); e != nil {
			return "", e
		}
	}
	svc := Service{DB: k.DB, Provider: FakeProvider{}, Enabled: true}
	if op == "purchase" {
		result, e = svc.purchaseTx(ctx, tx, actor, target)
	} else {
		result = target
		e = svc.refundTx(ctx, tx, actor, target)
		if errors.Is(e, ErrDenied) {
			// Recover a historical reservation whose refund committed before its result.
			var n int
			queryErr := tx.QueryRowContext(ctx, `SELECT count(*) FROM marketplace_entitlements e JOIN marketplace_listings l ON l.id=e.listing_id JOIN published_protocols p ON p.id=l.protocol_id
    WHERE e.id=? AND e.provider='fake' AND e.status='refunded' AND (e.buyer_did=? OR (l.creator_did=? AND p.creator_did=l.creator_did))
    AND (SELECT count(*) FROM marketplace_payment_audit a WHERE a.entitlement_id=e.id AND a.event='fake_refund')=1`, target, actor, actor).Scan(&n)
			if queryErr != nil {
				return "", queryErr
			}
			if n == 1 {
				e = nil
			}
		}
	}
	if e != nil {
		return "", e
	}
	res, e := tx.ExecContext(ctx, `UPDATE fake_payment_idempotency SET result_id=? WHERE actor_did=? AND request_key=? AND result_id=''`, result, actor, key)
	if e != nil {
		return "", e
	}
	n, e := res.RowsAffected()
	if e != nil {
		return "", e
	}
	if n != 1 {
		return "", ErrIdempotencyConflict
	}
	if e = tx.Commit(); e != nil {
		return "", e
	}
	return result, nil
}
func (k KeyedOperations) Purchase(ctx context.Context, actor, key, listing string) (string, error) {
	return k.run(ctx, actor, key, "purchase", listing)
}
func (k KeyedOperations) Refund(ctx context.Context, actor, key, id string) error {
	_, e := k.run(ctx, actor, key, "refund", id)
	return e
}
