package marketplace

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// FakeAccounting operates only over the existing fake-provider entitlement
// records. No payment provider, transfers, network or credentials.
var ErrMissingSaleSnapshot = errors.New("fake entitlement missing immutable sale snapshot")

type FakeAccounting struct {
	DB      *sql.DB
	Enabled bool
	FeeBPS  int64
	Now     func() time.Time
}

func (a FakeAccounting) time() string {
	if a.Now != nil {
		return a.Now().UTC().Format(time.RFC3339Nano)
	}
	return time.Now().UTC().Format(time.RFC3339Nano)
}
func (a FakeAccounting) check() error {
	if !a.Enabled || a.DB == nil || a.FeeBPS < 0 || a.FeeBPS > 10000 {
		return ErrDenied
	}
	return nil
}
func (a FakeAccounting) fee(gross int64) (int64, error) {
	if gross < 0 || gross > 1000000 {
		return 0, ErrDenied
	}
	return (gross*a.FeeBPS + 5000) / 10000, nil
}

// Reconcile builds immutable, idempotent sale and full reversal entries from
// committed fake entitlements. Re-running cannot duplicate ledger rows.
// The caller schedules this explicitly; it is not an external background job.
func (a FakeAccounting) Reconcile(ctx context.Context) (int, error) {
	if e := a.check(); e != nil {
		return 0, e
	}
	tx, e := a.DB.BeginTx(ctx, nil)
	if e != nil {
		return 0, e
	}
	defer tx.Rollback()
	var missing int
	e = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM marketplace_entitlements e LEFT JOIN fake_sale_snapshots s ON s.entitlement_id=e.id WHERE e.provider='fake' AND s.entitlement_id IS NULL`).Scan(&missing)
	if e != nil {
		return 0, e
	}
	if missing != 0 {
		return 0, fmt.Errorf("%w: %d", ErrMissingSaleSnapshot, missing)
	}
	rows, e := tx.QueryContext(ctx, `SELECT e.id,s.seller_did,s.price_minor,s.currency,e.status FROM marketplace_entitlements e JOIN fake_sale_snapshots s ON s.entitlement_id=e.id WHERE e.provider='fake' ORDER BY e.id`)
	if e != nil {
		return 0, e
	}
	type item struct {
		id, seller, currency, status string
		gross                        int64
	}
	items := []item{}
	for rows.Next() {
		var x item
		if e = rows.Scan(&x.id, &x.seller, &x.gross, &x.currency, &x.status); e != nil {
			rows.Close()
			return 0, e
		}
		items = append(items, x)
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return 0, e
	}
	inserted := 0
	for _, x := range items {
		var count int
		if e = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM marketplace_payment_audit WHERE entitlement_id=? AND event='fake_purchase'`, x.id).Scan(&count); e != nil {
			return 0, e
		}
		if count != 1 || x.gross <= 0 || x.currency != "USD" {
			return 0, ErrDenied
		}
		fee, e := a.fee(x.gross)
		if e != nil {
			return 0, e
		}
		now := a.time()
		res, e := tx.ExecContext(ctx, `INSERT INTO fake_seller_ledger(entitlement_id,entry_kind,seller_did,gross_minor,fee_minor,net_minor,currency,recorded_at) VALUES(?,'sale',?,?,?,?,?,?) ON CONFLICT(entitlement_id,entry_kind) DO NOTHING`, x.id, x.seller, x.gross, fee, x.gross-fee, x.currency, now)
		if e != nil {
			return 0, e
		}
		n, e := res.RowsAffected()
		if e != nil {
			return 0, e
		}
		inserted += int(n)
		// Existing rows are not trusted simply because ON CONFLICT ignored an
		// insert. Validate immutable sale identity and arithmetic on every pass,
		// including non-refunded entitlements. Historical fee BPS can differ.
		var recordedSeller, recordedCurrency string
		var recordedGross, recordedFee, recordedNet int64
		if e = tx.QueryRowContext(ctx, `SELECT seller_did,currency,gross_minor,fee_minor,net_minor FROM fake_seller_ledger WHERE entitlement_id=? AND entry_kind='sale'`, x.id).Scan(&recordedSeller, &recordedCurrency, &recordedGross, &recordedFee, &recordedNet); e != nil {
			return 0, e
		}
		if recordedSeller != x.seller || recordedCurrency != x.currency || recordedGross != x.gross || recordedFee < 0 || recordedFee > recordedGross || recordedNet != recordedGross-recordedFee {
			return 0, ErrDenied
		}
		if x.status == "active" {
			if e = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM fake_seller_ledger WHERE entitlement_id=? AND entry_kind='reversal'`, x.id).Scan(&count); e != nil {
				return 0, e
			}
			if count != 0 {
				return 0, ErrDenied
			}
		} else if x.status != "refunded" {
			return 0, ErrDenied
		}
		if x.status == "refunded" {
			if e = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM marketplace_payment_audit WHERE entitlement_id=? AND event='fake_refund'`, x.id).Scan(&count); e != nil {
				return 0, e
			}
			if count != 1 {
				return 0, ErrDenied
			}
			var saleSeller, saleCurrency string
			var saleGross, saleFee, saleNet int64
			e = tx.QueryRowContext(ctx, `SELECT seller_did,gross_minor,fee_minor,net_minor,currency FROM fake_seller_ledger WHERE entitlement_id=? AND entry_kind='sale'`, x.id).Scan(&saleSeller, &saleGross, &saleFee, &saleNet, &saleCurrency)
			if e != nil {
				return 0, e
			}
			if saleSeller != x.seller || saleCurrency != x.currency || saleGross != x.gross || saleGross-saleFee != saleNet {
				return 0, ErrDenied
			}
			res, e = tx.ExecContext(ctx, `INSERT INTO fake_seller_ledger(entitlement_id,entry_kind,seller_did,gross_minor,fee_minor,net_minor,currency,recorded_at) VALUES(?,'reversal',?,?,?,?,?,?) ON CONFLICT(entitlement_id,entry_kind) DO NOTHING`, x.id, saleSeller, -saleGross, -saleFee, -saleNet, saleCurrency, now)
			if e != nil {
				return 0, e
			}
			n, e = res.RowsAffected()
			if e != nil {
				return 0, e
			}
			inserted += int(n)
			// A pre-existing reversal must still exactly negate the original sale.
			var reverseSeller, reverseCurrency string
			var reverseGross, reverseFee, reverseNet int64
			if e = tx.QueryRowContext(ctx, `SELECT seller_did,currency,gross_minor,fee_minor,net_minor FROM fake_seller_ledger WHERE entitlement_id=? AND entry_kind='reversal'`, x.id).Scan(&reverseSeller, &reverseCurrency, &reverseGross, &reverseFee, &reverseNet); e != nil {
				return 0, e
			}
			if reverseSeller != recordedSeller || reverseCurrency != recordedCurrency || reverseGross != -recordedGross || reverseFee != -recordedFee || reverseNet != -recordedNet {
				return 0, ErrDenied
			}
		}
	}
	if e = tx.Commit(); e != nil {
		return 0, e
	}
	return inserted, nil
}
func (a FakeAccounting) SellerBalance(ctx context.Context, sellerDID string) (int64, error) {
	if e := a.check(); e != nil {
		return 0, e
	}
	var sum sql.NullInt64
	e := a.DB.QueryRowContext(ctx, `SELECT SUM(net_minor) FROM fake_seller_ledger WHERE seller_did=?`, sellerDID).Scan(&sum)
	return sum.Int64, e
}

// RunReconciliationJob is one bounded explicit local invocation, not a
// scheduler or an external provider job. Audit mismatches prevent ledger writes.
func (a FakeAccounting) RunReconciliationJob(ctx context.Context) (int, error) {
	if e := a.check(); e != nil {
		return 0, e
	}
	mismatch, e := (Service{DB: a.DB, Provider: FakeProvider{}, Enabled: true}).ReconcileFake(ctx)
	if e != nil {
		return 0, e
	}
	if mismatch != 0 {
		return 0, ErrDenied
	}
	return a.Reconcile(ctx)
}

// OpenDispute is buyer-only, deduplicated and independently audited.
func (a FakeAccounting) OpenDispute(ctx context.Context, buyerDID, entitlementID, reason string) error {
	if e := a.check(); e != nil {
		return e
	}
	if reason != "not_received" && reason != "not_as_described" && reason != "other" {
		return ErrDenied
	}
	tx, e := a.DB.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	var buyer, status, provider string
	e = tx.QueryRowContext(ctx, "SELECT buyer_did,status,provider FROM marketplace_entitlements WHERE id=?", entitlementID).Scan(&buyer, &status, &provider)
	if errors.Is(e, sql.ErrNoRows) {
		return ErrDenied
	}
	if e != nil {
		return e
	}
	if buyer != buyerDID || status != "active" || provider != "fake" {
		return ErrDenied
	}
	now := a.time()
	if _, e = tx.ExecContext(ctx, `INSERT INTO fake_payment_disputes(entitlement_id,requester_did,reason_key,status,opened_at) VALUES(?,?,?,'open',?)`, entitlementID, buyerDID, reason, now); e != nil {
		return ErrDenied
	}
	if _, e = tx.ExecContext(ctx, `INSERT INTO fake_payment_dispute_audit(entitlement_id,event,actor_did,recorded_at) VALUES(?,'opened',?,?)`, entitlementID, buyerDID, now); e != nil {
		return e
	}
	return tx.Commit()
}

// ResolveDispute is an operator primitive requiring independently authorized
// calling code. Refund authorization remains in the existing fake Service.
func (a FakeAccounting) ResolveDispute(ctx context.Context, operator, entitlementID string, refund bool) error {
	if e := a.check(); e != nil {
		return e
	}
	if operator == "" || entitlementID == "" {
		return ErrDenied
	}
	tx, e := a.DB.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	var status, buyer, receipt string
	e = tx.QueryRowContext(ctx, `SELECT d.status,e.buyer_did,e.provider_receipt FROM fake_payment_disputes d JOIN marketplace_entitlements e ON e.id=d.entitlement_id WHERE d.entitlement_id=? AND e.provider='fake'`, entitlementID).Scan(&status, &buyer, &receipt)
	if e != nil || status != "open" {
		return ErrDenied
	}
	decision := "resolved_no_refund"
	now := a.time()
	if refund {
		decision = "resolved_refund"
		var current string
		e = tx.QueryRowContext(ctx, `SELECT status FROM marketplace_entitlements WHERE id=?`, entitlementID).Scan(&current)
		if e != nil || current != "active" {
			return ErrDenied
		}
		if _, e = (FakeProvider{}).Refund(ctx, receipt); e != nil {
			return e
		}
		if _, e = tx.ExecContext(ctx, `UPDATE marketplace_entitlements SET status='refunded',refunded_at=? WHERE id=? AND status='active'`, now, entitlementID); e != nil {
			return e
		}
		audit, e := randomID()
		if e != nil {
			return e
		}
		if _, e = tx.ExecContext(ctx, `INSERT INTO marketplace_payment_audit(id,entitlement_id,event,actor_did,recorded_at) VALUES(?,?,'fake_refund',?,?)`, audit, entitlementID, operator, now); e != nil {
			return e
		}
	}
	if _, e = tx.ExecContext(ctx, `UPDATE fake_payment_disputes SET status=?,decided_at=? WHERE entitlement_id=? AND status='open'`, decision, now, entitlementID); e != nil {
		return e
	}
	if _, e = tx.ExecContext(ctx, `INSERT INTO fake_payment_dispute_audit(entitlement_id,event,actor_did,recorded_at) VALUES(?,?,?,?)`, entitlementID, decision, operator, now); e != nil {
		return e
	}
	_ = buyer
	return tx.Commit()
}
