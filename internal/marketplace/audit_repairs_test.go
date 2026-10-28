package marketplace

import (
	"context"
	"errors"
	"testing"
)

func TestSellerEligibilityUnlistingAndIdempotentFailureRecovery(t *testing.T) {
	db, original := setupReconciliationJobTest(t)
	ctx := context.Background()
	svc := Service{DB: db, Provider: FakeProvider{}, Enabled: true}
	for _, bad := range []string{`UPDATE social_profiles SET claimed=0`, `UPDATE published_protocols SET status='draft'`} {
		if _, e := db.ExecContext(ctx, bad); e != nil {
			t.Fatal(e)
		}
		if _, e := svc.Purchase(ctx, "did:plc:newbuyer", "l"); !errors.Is(e, ErrDenied) {
			t.Fatalf("ineligible seller accepted: %v", e)
		}
		if _, e := db.ExecContext(ctx, `UPDATE social_profiles SET claimed=1`); e != nil {
			t.Fatal(e)
		}
		if _, e := db.ExecContext(ctx, `UPDATE published_protocols SET status='published'`); e != nil {
			t.Fatal(e)
		}
	}
	if _, e := db.ExecContext(ctx, `UPDATE marketplace_listings SET status='unlisted'`); e != nil {
		t.Fatal(e)
	}
	if ok, e := svc.PublicationAccess(ctx, "did:plc:outsider", "p"); e != nil || ok {
		t.Fatal("unlisting granted free access", ok, e)
	}
	if ok, e := svc.PublicationAccess(ctx, "did:plc:buyer", "p"); e != nil || !ok {
		t.Fatal("unlisting removed purchased access", ok, e)
	}
	if e := svc.Refund(ctx, "did:plc:buyer", original); e != nil {
		t.Fatal(e)
	}
	if ok, e := svc.PublicationAccess(ctx, "did:plc:buyer", "p"); e != nil || ok {
		t.Fatal("refund restored access", ok, e)
	}
	if _, e := db.ExecContext(ctx, `UPDATE marketplace_listings SET status='listed'`); e != nil {
		t.Fatal(e)
	}
	keyed := KeyedOperations{DB: db, Enabled: true}
	if _, e := db.ExecContext(ctx, `CREATE TRIGGER fail_key_result BEFORE UPDATE ON fake_payment_idempotency BEGIN SELECT RAISE(ABORT,'injected result failure'); END`); e != nil {
		t.Fatal(e)
	}
	if _, e := keyed.Purchase(ctx, "did:plc:newbuyer", "audit-key-123", "l"); e == nil {
		t.Fatal("result failure ignored")
	}
	var n int
	if e := db.QueryRowContext(ctx, `SELECT count(*) FROM marketplace_entitlements WHERE buyer_did='did:plc:newbuyer'`).Scan(&n); e != nil || n != 0 {
		t.Fatal("payment committed without result", n, e)
	}
	if _, e := db.ExecContext(ctx, `DROP TRIGGER fail_key_result`); e != nil {
		t.Fatal(e)
	}
	id, e := keyed.Purchase(ctx, "did:plc:newbuyer", "audit-key-123", "l")
	if e != nil {
		t.Fatal("retry blocked", e)
	}
	again, e := keyed.Purchase(ctx, "did:plc:newbuyer", "audit-key-123", "l")
	if e != nil || again != id {
		t.Fatal("retry duplicated purchase", again, e)
	}
	refundID, e := keyed.Purchase(ctx, "did:plc:refundbuyer", "refund-purchase-123", "l")
	if e != nil {
		t.Fatal(e)
	}
	if _, e = db.ExecContext(ctx, `CREATE TRIGGER fail_key_result BEFORE UPDATE ON fake_payment_idempotency BEGIN SELECT RAISE(ABORT,'injected result failure'); END`); e != nil {
		t.Fatal(e)
	}
	if e = keyed.Refund(ctx, "did:plc:refundbuyer", "refund-key-123", refundID); e == nil {
		t.Fatal("refund result failure ignored")
	}
	var state string
	if e = db.QueryRowContext(ctx, `SELECT status FROM marketplace_entitlements WHERE id=?`, refundID).Scan(&state); e != nil || state != "active" {
		t.Fatal("refund committed without result", state, e)
	}
	if _, e = db.ExecContext(ctx, `DROP TRIGGER fail_key_result`); e != nil {
		t.Fatal(e)
	}
	for i := 0; i < 2; i++ {
		if e = keyed.Refund(ctx, "did:plc:refundbuyer", "refund-key-123", refundID); e != nil {
			t.Fatal("refund retry blocked", e)
		}
	}
	a := FakeAccounting{DB: db, Enabled: true}
	if _, e = a.Reconcile(ctx); e != nil {
		t.Fatal(e)
	}
	if _, e = db.ExecContext(ctx, `INSERT INTO fake_seller_ledger(entitlement_id,entry_kind,seller_did,gross_minor,fee_minor,net_minor,currency,recorded_at) VALUES(?,'reversal','did:plc:seller',-101,0,-101,'USD','now')`, id); e != nil {
		t.Fatal(e)
	}
	if _, e = a.Reconcile(ctx); !errors.Is(e, ErrDenied) {
		t.Fatal("active entitlement reversal accepted", e)
	}
}
