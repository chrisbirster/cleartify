package beta

import (
	"context"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/chrisbirster/cleartify/internal/database"
	"github.com/chrisbirster/cleartify/internal/notifications"
	"github.com/chrisbirster/cleartify/internal/security"
	"github.com/chrisbirster/cleartify/internal/waitlist"
)

type fakeSender struct {
	emails []notifications.Email
	sms    []notifications.SMS
}

func (f *fakeSender) SendEmail(_ context.Context, message notifications.Email) error {
	f.emails = append(f.emails, message)
	return nil
}

func (f *fakeSender) SendSMS(_ context.Context, message notifications.SMS) error {
	f.sms = append(f.sms, message)
	return nil
}

func TestInviteStartVerifyCreatesPhoneFirstAccount(t *testing.T) {
	ctx := context.Background()
	store, err := database.Open(ctx, database.Config{Path: ":memory:"})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer store.Close()
	if err := store.Migrate(ctx); err != nil {
		t.Fatalf("migrate database: %v", err)
	}

	if _, err := waitlist.NewService(store).Join(ctx, waitlist.Signup{
		Name: "Chris", Email: "chris@example.com", Phone: "(717) 555-1212",
		EmailUpdates: true, Source: "test",
	}); err != nil {
		t.Fatalf("join waitlist: %v", err)
	}

	tokens, err := security.NewTokens("test-secret-that-is-definitely-longer-than-32-bytes")
	if err != nil {
		t.Fatalf("new token manager: %v", err)
	}
	sender := &fakeSender{}
	service := NewService(store, tokens, sender, "http://localhost:3000")

	if _, err := service.Invite(ctx, "CHRIS@example.com"); err != nil {
		t.Fatalf("invite: %v", err)
	}
	if len(sender.emails) != 1 {
		t.Fatalf("emails = %d, want 1", len(sender.emails))
	}

	rawInviteToken := extractInviteToken(t, sender.emails[0].Text)
	var storedInviteHash string
	if err := store.DB.QueryRowContext(ctx,
		`SELECT email_token_hash FROM beta_invites LIMIT 1`,
	).Scan(&storedInviteHash); err != nil {
		t.Fatalf("read invite hash: %v", err)
	}
	if storedInviteHash == rawInviteToken || storedInviteHash == "" {
		t.Fatal("raw invite token must not be stored")
	}

	started, err := service.Start(ctx, rawInviteToken)
	if err != nil {
		t.Fatalf("start activation: %v", err)
	}
	if started.ActivationToken == "" || started.PhoneHint != "•••1212" {
		t.Fatalf("unexpected start result: %#v", started)
	}
	if len(sender.sms) != 1 {
		t.Fatalf("sms after start = %d, want 1", len(sender.sms))
	}

	code := extractCode(t, sender.sms[0].Text)
	var storedCodeHash string
	if err := store.DB.QueryRowContext(ctx,
		`SELECT challenge_code_hash FROM beta_invites LIMIT 1`,
	).Scan(&storedCodeHash); err != nil {
		t.Fatalf("read challenge hash: %v", err)
	}
	if storedCodeHash == code || storedCodeHash == "" {
		t.Fatal("raw challenge code must not be stored")
	}

	if _, err := service.Verify(ctx, started.ActivationToken, "99999"); err != ErrInvalidCode && code != "99999" {
		t.Fatalf("wrong code error = %v, want ErrInvalidCode", err)
	}

	activated, err := service.Verify(ctx, started.ActivationToken, code)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if activated.UserID == "" {
		t.Fatal("expected activated user id")
	}

	var users, phones, starters int
	if err := store.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&users); err != nil {
		t.Fatalf("count users: %v", err)
	}
	if err := store.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM phone_identities`).Scan(&phones); err != nil {
		t.Fatalf("count phone identities: %v", err)
	}
	if err := store.DB.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM protocol_installations WHERE source = 'starter'`,
	).Scan(&starters); err != nil {
		t.Fatalf("count starter protocols: %v", err)
	}
	if users != 1 || phones != 1 || starters != len(starterProtocols) {
		t.Fatalf("users=%d phones=%d starters=%d", users, phones, starters)
	}

	var status string
	if err := store.DB.QueryRowContext(ctx,
		`SELECT status FROM waitlist_entries LIMIT 1`,
	).Scan(&status); err != nil {
		t.Fatalf("read waitlist status: %v", err)
	}
	if status != "activated" {
		t.Fatalf("waitlist status = %q, want activated", status)
	}
	if len(sender.sms) != 2 || !strings.Contains(sender.sms[1].Text, "I'm Clear") {
		t.Fatalf("expected welcome text after activation, sms=%#v", sender.sms)
	}
}

func TestInviteLinkCanOnlyStartOnce(t *testing.T) {
	ctx := context.Background()
	store, err := database.Open(ctx, database.Config{Path: ":memory:"})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer store.Close()
	if err := store.Migrate(ctx); err != nil {
		t.Fatalf("migrate database: %v", err)
	}

	if _, err := waitlist.NewService(store).Join(ctx, waitlist.Signup{
		Name: "Chris", Email: "chris@example.com", Phone: "+17175551212",
	}); err != nil {
		t.Fatalf("join waitlist: %v", err)
	}
	tokens, _ := security.NewTokens("test-secret-that-is-definitely-longer-than-32-bytes")
	sender := &fakeSender{}
	service := NewService(store, tokens, sender, "http://localhost:3000")

	if _, err := service.Invite(ctx, "chris@example.com"); err != nil {
		t.Fatalf("invite: %v", err)
	}
	token := extractInviteToken(t, sender.emails[0].Text)
	if _, err := service.Start(ctx, token); err != nil {
		t.Fatalf("first start: %v", err)
	}
	if _, err := service.Start(ctx, token); err != ErrInviteUsed {
		t.Fatalf("second start error = %v, want ErrInviteUsed", err)
	}
}

func extractInviteToken(t *testing.T, body string) string {
	t.Helper()
	re := regexp.MustCompile(`https?://[^\s]+/beta/activate\?token=([^\s]+)`)
	match := re.FindStringSubmatch(body)
	if len(match) != 2 {
		t.Fatalf("invite URL token not found in %q", body)
	}
	token, err := url.QueryUnescape(match[1])
	if err != nil {
		t.Fatalf("decode invite token: %v", err)
	}
	return token
}

func extractCode(t *testing.T, body string) string {
	t.Helper()
	re := regexp.MustCompile(`\b([0-9]{5})\b`)
	match := re.FindStringSubmatch(body)
	if len(match) != 2 {
		t.Fatalf("challenge code not found in %q", body)
	}
	return match[1]
}
