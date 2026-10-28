package beta

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/chrisbirster/cleartify/internal/database"
	"github.com/chrisbirster/cleartify/internal/notifications"
	"github.com/chrisbirster/cleartify/internal/security"
	"github.com/chrisbirster/cleartify/internal/waitlist"
)

func activationFixture(t *testing.T) (*Service, *database.Store, *fakeSender, StartResult, string) {
	t.Helper()
	ctx := context.Background()
	db, err := database.Open(ctx, database.Config{Path: filepath.Join(t.TempDir(), "activation.db")})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err = db.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	_, err = waitlist.NewService(db).Join(ctx, waitlist.Signup{
		Name: "Original", Email: "activation@example.test", Phone: "+15550000001", AcceptPolicies: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	tokens, err := security.NewTokens("test-only-activation-secret-000000000000")
	if err != nil {
		t.Fatal(err)
	}
	sender := &fakeSender{}
	s := NewService(db, tokens, sender, "https://app.example.test")
	if _, err = s.Invite(ctx, "activation@example.test"); err != nil {
		t.Fatal(err)
	}
	started, err := s.Start(ctx, extractInviteToken(t, sender.emails[0].Text))
	if err != nil {
		t.Fatal(err)
	}
	return s, db, sender, started, extractCode(t, sender.sms[0].Text)
}

func assertNoActivatedAccount(t *testing.T, db *database.Store) {
	t.Helper()
	for _, table := range []string{"users", "phone_identities", "protocol_installations"} {
		var count int
		if err := db.DB.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count); err != nil || count != 0 {
			t.Fatalf("%s has %d rows after rejected activation: %v", table, count, err)
		}
	}
}

func TestRevokedChallengeCannotVerifyOrResend(t *testing.T) {
	s, db, sender, started, code := activationFixture(t)
	ctx := context.Background()
	if _, err := s.Invite(ctx, "activation@example.test"); err != nil {
		t.Fatal(err)
	}
	for _, attempt := range []string{code, "00000"} {
		if _, err := s.Verify(ctx, started.ActivationToken, attempt); !errors.Is(err, ErrActivationNotFound) {
			t.Fatalf("revoked challenge accepted or changed: %v", err)
		}
	}
	s.now = func() time.Time { return time.Now().Add(2 * time.Minute) }
	if _, err := s.Resend(ctx, started.ActivationToken); !errors.Is(err, ErrActivationNotFound) {
		t.Fatalf("revoked resend: %v", err)
	}
	if len(sender.sms) != 1 {
		t.Fatal("revoked challenge sent another SMS")
	}
	var status string
	if err := db.DB.QueryRow("SELECT status FROM beta_invites WHERE activation_token_hash=?", s.tokens.Digest(started.ActivationToken)).Scan(&status); err != nil || status != "revoked" {
		t.Fatalf("revoked state overwritten: %s %v", status, err)
	}
	assertNoActivatedAccount(t, db)
}

func TestActivationUsesChallengedIdentityAfterWaitlistMutationAndRestart(t *testing.T) {
	s, db, sender, started, code := activationFixture(t)
	// Even a historical/manual edit must not retarget an already-issued challenge.
	_, err := db.DB.Exec(`UPDATE waitlist_entries SET phone_e164='+15550000002',name='Changed'`)
	if err != nil {
		t.Fatal(err)
	}
	var path string
	var seq int
	var schema string
	if err = db.DB.QueryRow("PRAGMA database_list").Scan(&seq, &schema, &path); err != nil {
		t.Fatal(err)
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := database.Open(context.Background(), database.Config{Path: path})
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	s.db = reopened
	result, err := s.Verify(context.Background(), started.ActivationToken, code)
	if err != nil {
		t.Fatal(err)
	}
	var phone, name string
	if err = reopened.DB.QueryRow(`SELECT p.phone_e164,u.display_name FROM phone_identities p JOIN users u ON u.id=p.user_id WHERE u.id=?`, result.UserID).Scan(&phone, &name); err != nil {
		t.Fatal(err)
	}
	if phone != sender.sms[0].To || phone != "+15550000001" || name != "Original" {
		t.Fatalf("activation retargeted identity: phone=%s name=%s", phone, name)
	}
}

func TestActivationRollsBackWhenInviteCannotBeClaimed(t *testing.T) {
	s, db, _, started, code := activationFixture(t)
	// Simulate an activation state update losing a race, without nondeterministic scheduling.
	_, err := db.DB.Exec(`CREATE TRIGGER skip_activation BEFORE UPDATE ON beta_invites
		WHEN NEW.status='activated' BEGIN SELECT RAISE(IGNORE); END`)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.Verify(context.Background(), started.ActivationToken, code); err == nil {
		t.Fatal("zero-row activation claim reported success")
	}
	assertNoActivatedAccount(t, db)
}

func TestResendDoesNotForgiveFailedAttempts(t *testing.T) {
	s, db, sender, started, _ := activationFixture(t)
	ctx := context.Background()
	for i := 0; i < maxCodeAttempts-1; i++ {
		if _, err := s.Verify(ctx, started.ActivationToken, "00000"); !errors.Is(err, ErrInvalidCode) {
			t.Fatal(err)
		}
	}
	s.now = func() time.Time { return time.Now().Add(2 * time.Minute) }
	if _, err := s.Resend(ctx, started.ActivationToken); err != nil {
		t.Fatal(err)
	}
	if len(sender.sms) != 2 {
		t.Fatal("expected replacement challenge")
	}
	if _, err := s.Verify(ctx, started.ActivationToken, "00000"); !errors.Is(err, ErrLocked) {
		t.Fatalf("resend reset the failed-attempt budget: %v", err)
	}
	if _, err := s.Verify(ctx, started.ActivationToken, extractCode(t, sender.sms[1].Text)); !errors.Is(err, ErrLocked) {
		t.Fatalf("locked challenge accepted: %v", err)
	}
	assertNoActivatedAccount(t, db)
}

func TestLegacyChallengeRequiresFreshInvitation(t *testing.T) {
	s, db, sender, started, code := activationFixture(t)
	ctx := context.Background()
	// Pre-upgrade invitations have no trustworthy identity snapshot.
	if _, err := db.DB.ExecContext(ctx, `DELETE FROM beta_invite_identities`); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Verify(ctx, started.ActivationToken, code); !errors.Is(err, ErrActivationNotFound) {
		t.Fatalf("legacy verify: %v", err)
	}
	assertNoActivatedAccount(t, db)
	if _, err := s.Invite(ctx, "activation@example.test"); err != nil {
		t.Fatal(err)
	}
	fresh, err := s.Start(ctx, extractInviteToken(t, sender.emails[len(sender.emails)-1].Text))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Verify(ctx, fresh.ActivationToken, extractCode(t, sender.sms[len(sender.sms)-1].Text)); err != nil {
		t.Fatal(err)
	}
}

func TestInvitedIdentityCannotBeUpdated(t *testing.T) {
	_, db, _, _, _ := activationFixture(t)
	if _, err := db.DB.Exec(`UPDATE beta_invite_identities SET phone_e164='+15550000002'`); err == nil {
		t.Fatal("identity snapshot was mutable")
	}
}

type failingChallengeSender struct {
	*fakeSender
	fail bool
}

func (f *failingChallengeSender) SendSMS(ctx context.Context, message notifications.SMS) error {
	if f.fail {
		return errors.New("injected delivery outage")
	}
	return f.fakeSender.SendSMS(ctx, message)
}
func TestActivationCanResumeAfterDeliveryFailureWithoutResettingBudget(t *testing.T) {
	s, db, sender, started, _ := activationFixture(t)
	ctx := context.Background()
	if _, e := s.Verify(ctx, started.ActivationToken, "00000"); !errors.Is(e, ErrInvalidCode) {
		t.Fatal(e)
	}
	raw := extractInviteToken(t, sender.emails[0].Text)
	resumed, e := s.Start(ctx, raw)
	if e != nil || resumed.ActivationToken != started.ActivationToken {
		t.Fatal("refresh recovery", resumed, e)
	}
	var attempts int
	var before, after string
	if e = db.DB.QueryRowContext(ctx, `SELECT challenge_attempts,activation_expires_at FROM beta_invites WHERE status='challenge'`).Scan(&attempts, &before); e != nil || attempts != 1 {
		t.Fatal("recovery reset attempts", attempts, e)
	}
	if len(sender.sms) != 1 {
		t.Fatal("recovery sent another SMS")
	}
	if _, e = s.Start(ctx, raw); e != nil {
		t.Fatal(e)
	}
	if e = db.DB.QueryRowContext(ctx, `SELECT activation_expires_at FROM beta_invites WHERE status='challenge'`).Scan(&after); e != nil || after != before {
		t.Fatal("resume extended expiry", before, after, e)
	}
	// Reissue through the ordinary invite flow, then fail the first challenge delivery.
	if _, e = s.Invite(ctx, "activation@example.test"); e != nil {
		t.Fatal(e)
	}
	failed := &failingChallengeSender{fakeSender: sender, fail: true}
	s.sender = failed
	raw = extractInviteToken(t, sender.emails[len(sender.emails)-1].Text)
	if _, e = s.Start(ctx, raw); e == nil {
		t.Fatal("delivery failure ignored")
	}
	resumed, e = s.Start(ctx, raw)
	if e != nil || resumed.ActivationToken == "" {
		t.Fatal("failed delivery stranded activation", resumed, e)
	}
	if e = db.DB.QueryRowContext(ctx, `SELECT activation_expires_at FROM beta_invites WHERE status='challenge'`).Scan(&after); e != nil {
		t.Fatal(e)
	}
	s.now = func() time.Time { return time.Now().Add(2 * time.Minute) }
	failed.fail = false
	if _, e = s.Resend(ctx, resumed.ActivationToken); e != nil {
		t.Fatal(e)
	}
	if _, e = s.Verify(ctx, resumed.ActivationToken, extractCode(t, sender.sms[len(sender.sms)-1].Text)); e != nil {
		t.Fatal(e)
	}

}
