package marketplace

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrDenied = errors.New("marketplace action denied")
var ErrPayment = errors.New("fake payment rejected")

type Receipt struct {
	ID       string
	Provider string
}
type Provider interface {
	Charge(context.Context, string, int64, string) (Receipt, error)
	Refund(context.Context, string) (Receipt, error)
}

// FakeProvider has no keys, network, currency movement or external side effects.
type FakeProvider struct{}

func (FakeProvider) Charge(_ context.Context, id string, amount int64, currency string) (Receipt, error) {
	if id == "" || amount < 0 || amount > 1000000 || currency != "USD" {
		return Receipt{}, ErrPayment
	}
	return Receipt{ID: "fake-" + id, Provider: "fake"}, nil
}
func (FakeProvider) Refund(_ context.Context, receipt string) (Receipt, error) {
	if !strings.HasPrefix(receipt, "fake-") {
		return Receipt{}, ErrPayment
	}
	return Receipt{ID: receipt, Provider: "fake"}, nil
}

type Service struct {
	DB       *sql.DB
	Provider Provider
	Enabled  bool
	Now      func() time.Time
}

func (s Service) clock() string {
	if s.Now != nil {
		return s.Now().UTC().Format(time.RFC3339Nano)
	}
	return time.Now().UTC().Format(time.RFC3339Nano)
}
func randomID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
func (s Service) provider() (Provider, error) {
	if _, ok := s.Provider.(FakeProvider); !ok {
		return nil, ErrPayment
	}
	return FakeProvider{}, nil
}

// AuthorizeSeller checks the trusted server-supplied actor against both listing
// and published protocol ownership. No client-provided seller role is accepted.
func (s Service) AuthorizeSeller(ctx context.Context, actorDID, listingID string) error {
	if !s.Enabled || s.DB == nil || !strings.HasPrefix(actorDID, "did:") {
		return ErrDenied
	}
	var n int
	err := s.DB.QueryRowContext(ctx, `SELECT count(*) FROM marketplace_listings l JOIN published_protocols p ON p.id=l.protocol_id WHERE l.id=? AND l.creator_did=? AND p.creator_did=?`, listingID, actorDID, actorDID).Scan(&n)
	if err != nil {
		return err
	}
	if n != 1 {
		return ErrDenied
	}
	return nil
}

// Purchase creates a test-only entitlement; no payment is collected.
func (s Service) Purchase(ctx context.Context, actorBuyerDID, listingID string) (string, error) {
	if !s.Enabled || s.DB == nil || !strings.HasPrefix(actorBuyerDID, "did:") {
		return "", ErrDenied
	}
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	id, err := s.purchaseTx(ctx, tx, actorBuyerDID, listingID)
	if err != nil {
		return "", err
	}
	return id, tx.Commit()
}
func (s Service) purchaseTx(ctx context.Context, tx *sql.Tx, actorBuyerDID, listingID string) (string, error) {
	provider, err := s.provider()
	if err != nil {
		return "", err
	}

	var amount int64
	var currency string
	err = tx.QueryRowContext(ctx, `SELECT price_minor,currency FROM marketplace_listings WHERE id=? AND status='listed' AND pricing_kind='one_time'`, listingID).Scan(&amount, &currency)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrDenied
	}
	if err != nil {
		return "", err
	}
	if amount <= 0 || amount > 1000000 || currency != "USD" {
		return "", ErrPayment
	}
	var seller, owner string
	err = tx.QueryRowContext(ctx, `SELECT l.creator_did,p.creator_did FROM marketplace_listings l JOIN published_protocols p ON p.id=l.protocol_id JOIN social_profiles seller ON seller.did=l.creator_did AND seller.claimed=1 AND seller.public_visible=1
 WHERE l.id=? AND p.status='published'
 AND NOT EXISTS(SELECT 1 FROM creator_publications c WHERE c.protocol_id=p.id AND c.visibility='private')
 AND NOT EXISTS(SELECT 1 FROM creator_takedown_requests t WHERE t.protocol_id=p.id AND t.status='hidden')
 AND NOT EXISTS(SELECT 1 FROM social_impersonation_holds h WHERE h.subject_did=p.creator_did AND h.released_at IS NULL) AND NOT EXISTS(SELECT 1 FROM social_blocks b WHERE (b.actor_did=? AND b.subject_did=p.creator_did) OR (b.actor_did=p.creator_did AND b.subject_did=?))`, listingID, actorBuyerDID, actorBuyerDID).Scan(&seller, &owner)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrDenied
	}
	if err != nil {
		return "", err
	}
	if seller != owner {
		return "", ErrDenied
	}
	var existing, status string
	err = tx.QueryRowContext(ctx, `SELECT id,status FROM marketplace_entitlements WHERE buyer_did=? AND listing_id=?`, actorBuyerDID, listingID).Scan(&existing, &status)
	if err == nil {
		if status == "active" {
			return existing, nil
		}
		return "", ErrDenied
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	id, err := randomID()
	if err != nil {
		return "", err
	}
	receipt, err := provider.Charge(ctx, id, amount, currency)
	if err != nil {
		return "", err
	}
	if receipt.Provider != "fake" || receipt.ID != "fake-"+id {
		return "", ErrPayment
	}
	now := s.clock()
	_, err = tx.ExecContext(ctx, `INSERT INTO marketplace_entitlements(id,listing_id,buyer_did,provider,provider_receipt,status,created_at) VALUES(?,?,?,'fake',?,'active',?)`, id, listingID, actorBuyerDID, receipt.ID, now)
	if err != nil {
		return "", fmt.Errorf("entitlement insert: %w", err)
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO fake_sale_snapshots(entitlement_id,seller_did,price_minor,currency,captured_at) VALUES(?,?,?,?,?)`, id, seller, amount, currency, now); err != nil {
		return "", err
	}
	auditID, err := randomID()
	if err != nil {
		return "", err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO marketplace_payment_audit(id,entitlement_id,event,actor_did,recorded_at) VALUES(?,?,'fake_purchase',?,?)`, auditID, id, actorBuyerDID, now); err != nil {
		return "", err
	}
	return id, nil
}
func (s Service) Entitled(ctx context.Context, buyerDID, listingID string) (bool, error) {
	if !s.Enabled || s.DB == nil {
		return false, nil
	}
	var n int
	err := s.DB.QueryRowContext(ctx, `SELECT count(*) FROM marketplace_entitlements WHERE buyer_did=? AND listing_id=? AND status='active'`, buyerDID, listingID).Scan(&n)
	return n > 0, err
}

// Refund is restricted to the original buyer or the validated seller, not arbitrary callers.
func (s Service) Refund(ctx context.Context, actorDID, entitlementID string) error {
	if !s.Enabled || s.DB == nil {
		return ErrDenied
	}
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err = s.refundTx(ctx, tx, actorDID, entitlementID); err != nil {
		return err
	}
	return tx.Commit()
}
func (s Service) refundTx(ctx context.Context, tx *sql.Tx, actorDID, entitlementID string) error {
	provider, err := s.provider()
	if err != nil {
		return err
	}

	var buyer, seller, protocolOwner, receipt string
	err = tx.QueryRowContext(ctx, `SELECT e.buyer_did,l.creator_did,p.creator_did,e.provider_receipt FROM marketplace_entitlements e JOIN marketplace_listings l ON l.id=e.listing_id JOIN published_protocols p ON p.id=l.protocol_id WHERE e.id=? AND e.status='active'`, entitlementID).Scan(&buyer, &seller, &protocolOwner, &receipt)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrDenied
	}
	if err != nil {
		return err
	}
	if actorDID != buyer && (actorDID != seller || seller != protocolOwner) {
		return ErrDenied
	}
	if _, err = provider.Refund(ctx, receipt); err != nil {
		return err
	}
	now := s.clock()
	res, err := tx.ExecContext(ctx, `UPDATE marketplace_entitlements SET status='refunded',refunded_at=? WHERE id=? AND status='active'`, now, entitlementID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil || n != 1 {
		return ErrDenied
	}
	auditID, err := randomID()
	if err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO marketplace_payment_audit(id,entitlement_id,event,actor_did,recorded_at) VALUES(?,?,'fake_refund',?,?)`, auditID, entitlementID, actorDID, now); err != nil {
		return err
	}
	return nil
}

// PublicationAccess allows free/unlisted publications without a paid listing,
// but requires a currently active fake entitlement for every priced listing.
// This method is intentionally fake-only and never confers real-money access.
func (s Service) PublicationAccess(ctx context.Context, buyerDID, publicationID string) (bool, error) {
	if !s.Enabled || s.DB == nil || !strings.HasPrefix(buyerDID, "did:") {
		return false, ErrDenied
	}
	var paid int
	err := s.DB.QueryRowContext(ctx, `SELECT count(*) FROM marketplace_listings WHERE protocol_id=? AND pricing_kind='one_time' AND price_minor>0`, publicationID).Scan(&paid)
	if err != nil {
		return false, err
	}
	if paid == 0 {
		return true, nil
	}
	var count int
	err = s.DB.QueryRowContext(ctx, `SELECT count(*) FROM marketplace_entitlements e JOIN marketplace_listings l ON l.id=e.listing_id WHERE l.protocol_id=? AND l.price_minor>0 AND e.buyer_did=? AND e.provider='fake' AND e.status='active'`, publicationID, buyerDID).Scan(&count)
	return count > 0, err
}

// ReconcileFake verifies that fake active/refunded entitlements each have one
// corresponding purchase audit, and refunded records have one refund audit.
func (s Service) ReconcileFake(ctx context.Context) (int, error) {
	if s.DB == nil {
		return 0, ErrDenied
	}
	var mismatches int
	err := s.DB.QueryRowContext(ctx, `SELECT count(*) FROM marketplace_entitlements e WHERE e.provider='fake' AND (
 (SELECT count(*) FROM marketplace_payment_audit a WHERE a.entitlement_id=e.id AND a.event='fake_purchase')!=1
 OR (e.status='refunded' AND (SELECT count(*) FROM marketplace_payment_audit a WHERE a.entitlement_id=e.id AND a.event='fake_refund')!=1)
 OR (e.status='active' AND EXISTS(SELECT 1 FROM marketplace_payment_audit a WHERE a.entitlement_id=e.id AND a.event='fake_refund'))
 )`).Scan(&mismatches)
	return mismatches, err
}
