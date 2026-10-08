package beta

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/chrisbirster/cleartify/internal/database"
	"github.com/chrisbirster/cleartify/internal/notifications"
	"github.com/chrisbirster/cleartify/internal/security"
)

var (
	ErrInviteNotFound      = errors.New("beta invite not found")
	ErrInviteExpired       = errors.New("beta invite expired")
	ErrInviteUsed          = errors.New("beta invite already used")
	ErrActivationNotFound  = errors.New("activation session not found")
	ErrActivationExpired   = errors.New("activation session expired")
	ErrInvalidCode         = errors.New("invalid beta code")
	ErrLocked              = errors.New("beta activation locked")
	ErrAlreadyActivated    = errors.New("beta invite already activated")
	ErrResendTooSoon       = errors.New("wait before requesting another code")
	ErrPhoneAlreadyActive  = errors.New("phone number already belongs to an active user")
)

const (
	emailInviteTTL    = 7 * 24 * time.Hour
	activationTTL     = 30 * time.Minute
	challengeTTL      = 10 * time.Minute
	resendCooldown    = 60 * time.Second
	maxCodeAttempts   = 5
)

var starterProtocols = []string{
	"remember-this",
	"follow-through",
	"weekly-reset",
}

type Service struct {
	db     *database.Store
	tokens *security.Tokens
	sender notifications.Sender
	appURL string
	now    func() time.Time
}

type InviteResult struct {
	InviteID string
}

type StartResult struct {
	ActivationToken string
	PhoneHint       string
	ExpiresIn       int
}

type ActivationResult struct {
	UserID string
}

func NewService(db *database.Store, tokens *security.Tokens, sender notifications.Sender, appURL string) *Service {
	return &Service{
		db: db,
		tokens: tokens,
		sender: sender,
		appURL: strings.TrimRight(appURL, "/"),
		now: time.Now,
	}
}

func (s *Service) Invite(ctx context.Context, email string) (InviteResult, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	var entry struct {
		ID    string
		Name  string
		Email string
		Phone string
	}
	err := s.db.DB.QueryRowContext(ctx, `
		SELECT id, name, email, phone_e164
		FROM waitlist_entries
		WHERE email_normalized = ?
		LIMIT 1
	`, email).Scan(&entry.ID, &entry.Name, &entry.Email, &entry.Phone)
	if errors.Is(err, sql.ErrNoRows) {
		return InviteResult{}, ErrInviteNotFound
	}
	if err != nil {
		return InviteResult{}, fmt.Errorf("load waitlist entry: %w", err)
	}

	rawToken, err := s.tokens.RandomURLToken(32)
	if err != nil {
		return InviteResult{}, err
	}
	inviteID, err := s.tokens.RandomURLToken(18)
	if err != nil {
		return InviteResult{}, err
	}

	now := s.now().UTC()
	expires := now.Add(emailInviteTTL)

	tx, err := s.db.DB.BeginTx(ctx, nil)
	if err != nil {
		return InviteResult{}, fmt.Errorf("begin invite transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
		UPDATE beta_invites
		SET status = 'revoked', updated_at = ?
		WHERE waitlist_entry_id = ? AND status IN ('sent', 'challenge')
	`, stamp(now), entry.ID); err != nil {
		return InviteResult{}, fmt.Errorf("revoke previous invites: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO beta_invites (
			id, waitlist_entry_id, status,
			email_token_hash, email_token_expires_at,
			challenge_attempts, max_challenge_attempts,
			created_at, updated_at
		) VALUES (?, ?, 'sent', ?, ?, 0, ?, ?, ?)
	`, inviteID, entry.ID, s.tokens.Digest(rawToken), stamp(expires), maxCodeAttempts, stamp(now), stamp(now)); err != nil {
		return InviteResult{}, fmt.Errorf("create beta invite: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE waitlist_entries SET status = 'invited', updated_at = ? WHERE id = ?
	`, stamp(now), entry.ID); err != nil {
		return InviteResult{}, fmt.Errorf("mark waitlist invited: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return InviteResult{}, fmt.Errorf("commit beta invite: %w", err)
	}
	if err := s.db.Push(ctx); err != nil {
		return InviteResult{}, err
	}

	activationURL := s.appURL + "/beta/activate?token=" + url.QueryEscape(rawToken)
	message := notifications.Email{
		To: entry.Email,
		Subject: "You're in — meet Clear",
		Text: "Hey " + entry.Name + ",

You're invited to the Cleartify beta.

Activate your invite:
" + activationURL + "

The link expires in 7 days.",
	}
	if err := s.sender.SendEmail(ctx, message); err != nil {
		_, _ = s.db.DB.ExecContext(ctx, `
			UPDATE beta_invites SET status = 'delivery_failed', updated_at = ? WHERE id = ?
		`, stamp(s.now().UTC()), inviteID)
		_ = s.db.Push(ctx)
		return InviteResult{}, fmt.Errorf("send beta invitation: %w", err)
	}

	return InviteResult{InviteID: inviteID}, nil
}

func (s *Service) Start(ctx context.Context, rawEmailToken string) (StartResult, error) {
	if strings.TrimSpace(rawEmailToken) == "" {
		return StartResult{}, ErrInviteNotFound
	}

	var invite struct {
		ID          string
		Status      string
		ExpiresAt   string
		UsedAt      sql.NullString
		Name        string
		Phone       string
	}
	err := s.db.DB.QueryRowContext(ctx, `
		SELECT i.id, i.status, i.email_token_expires_at, i.email_token_used_at,
		       w.name, w.phone_e164
		FROM beta_invites i
		JOIN waitlist_entries w ON w.id = i.waitlist_entry_id
		WHERE i.email_token_hash = ?
		LIMIT 1
	`, s.tokens.Digest(rawEmailToken)).Scan(
		&invite.ID, &invite.Status, &invite.ExpiresAt, &invite.UsedAt,
		&invite.Name, &invite.Phone,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return StartResult{}, ErrInviteNotFound
	}
	if err != nil {
		return StartResult{}, fmt.Errorf("load beta invite: %w", err)
	}
	if invite.Status == "activated" {
		return StartResult{}, ErrAlreadyActivated
	}
	if invite.UsedAt.Valid {
		return StartResult{}, ErrInviteUsed
	}
	if expired(invite.ExpiresAt, s.now()) {
		return StartResult{}, ErrInviteExpired
	}
	if invite.Status != "sent" {
		return StartResult{}, ErrInviteNotFound
	}

	activationToken, err := s.tokens.RandomURLToken(32)
	if err != nil {
		return StartResult{}, err
	}
	code, err := s.tokens.FiveDigitCode()
	if err != nil {
		return StartResult{}, err
	}

	now := s.now().UTC()
	activationExpires := now.Add(activationTTL)
	codeExpires := now.Add(challengeTTL)

	result, err := s.db.DB.ExecContext(ctx, `
		UPDATE beta_invites
		SET status = 'challenge',
		    email_token_used_at = ?,
		    activation_token_hash = ?,
		    activation_expires_at = ?,
		    challenge_code_hash = ?,
		    challenge_expires_at = ?,
		    challenge_attempts = 0,
		    challenge_sent_at = ?,
		    updated_at = ?
		WHERE id = ? AND email_token_used_at IS NULL AND status = 'sent'
	`, stamp(now), s.tokens.Digest(activationToken), stamp(activationExpires),
		s.tokens.Digest(code), stamp(codeExpires), stamp(now), stamp(now), invite.ID)
	if err != nil {
		return StartResult{}, fmt.Errorf("start beta activation: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return StartResult{}, fmt.Errorf("read activation update: %w", err)
	}
	if affected != 1 {
		return StartResult{}, ErrInviteUsed
	}
	if err := s.db.Push(ctx); err != nil {
		return StartResult{}, err
	}

	if err := s.sender.SendSMS(ctx, notifications.SMS{
		To: invite.Phone,
		Text: "Your Cleartify beta code is " + code + ". It expires in 10 minutes.",
	}); err != nil {
		return StartResult{}, fmt.Errorf("send beta challenge: %w", err)
	}

	return StartResult{
		ActivationToken: activationToken,
		PhoneHint: maskPhone(invite.Phone),
		ExpiresIn: int(challengeTTL.Seconds()),
	}, nil
}

func (s *Service) Resend(ctx context.Context, activationToken string) (StartResult, error) {
	session, err := s.loadActivation(ctx, activationToken)
	if err != nil {
		return StartResult{}, err
	}
	if session.Status == "activated" {
		return StartResult{}, ErrAlreadyActivated
	}
	if session.Status == "locked" {
		return StartResult{}, ErrLocked
	}
	if expired(session.ActivationExpiresAt, s.now()) {
		return StartResult{}, ErrActivationExpired
	}
	if session.ChallengeSentAt.Valid {
		sentAt, err := time.Parse(time.RFC3339Nano, session.ChallengeSentAt.String)
		if err == nil && s.now().Before(sentAt.Add(resendCooldown)) {
			return StartResult{}, ErrResendTooSoon
		}
	}

	code, err := s.tokens.FiveDigitCode()
	if err != nil {
		return StartResult{}, err
	}
	now := s.now().UTC()
	if _, err := s.db.DB.ExecContext(ctx, `
		UPDATE beta_invites
		SET challenge_code_hash = ?, challenge_expires_at = ?,
		    challenge_attempts = 0, challenge_sent_at = ?, updated_at = ?
		WHERE id = ? AND status = 'challenge'
	`, s.tokens.Digest(code), stamp(now.Add(challengeTTL)), stamp(now), stamp(now), session.ID); err != nil {
		return StartResult{}, fmt.Errorf("refresh beta challenge: %w", err)
	}
	if err := s.db.Push(ctx); err != nil {
		return StartResult{}, err
	}
	if err := s.sender.SendSMS(ctx, notifications.SMS{
		To: session.Phone,
		Text: "Your new Cleartify beta code is " + code + ". It expires in 10 minutes.",
	}); err != nil {
		return StartResult{}, fmt.Errorf("resend beta challenge: %w", err)
	}

	return StartResult{
		ActivationToken: activationToken,
		PhoneHint: maskPhone(session.Phone),
		ExpiresIn: int(challengeTTL.Seconds()),
	}, nil
}

func (s *Service) Verify(ctx context.Context, activationToken, code string) (ActivationResult, error) {
	session, err := s.loadActivation(ctx, activationToken)
	if err != nil {
		return ActivationResult{}, err
	}
	if session.Status == "activated" {
		return ActivationResult{}, ErrAlreadyActivated
	}
	if session.Status == "locked" || session.Attempts >= session.MaxAttempts {
		return ActivationResult{}, ErrLocked
	}
	if expired(session.ActivationExpiresAt, s.now()) || expired(session.ChallengeExpiresAt, s.now()) {
		return ActivationResult{}, ErrActivationExpired
	}

	if !s.tokens.Matches(strings.TrimSpace(code), session.ChallengeCodeHash) {
		attempts := session.Attempts + 1
		status := "challenge"
		if attempts >= session.MaxAttempts {
			status = "locked"
		}
		if _, err := s.db.DB.ExecContext(ctx, `
			UPDATE beta_invites
			SET challenge_attempts = ?, status = ?, updated_at = ?
			WHERE id = ?
		`, attempts, status, stamp(s.now().UTC()), session.ID); err != nil {
			return ActivationResult{}, fmt.Errorf("record challenge attempt: %w", err)
		}
		_ = s.db.Push(ctx)
		if status == "locked" {
			return ActivationResult{}, ErrLocked
		}
		return ActivationResult{}, ErrInvalidCode
	}

	var existingUser string
	err = s.db.DB.QueryRowContext(ctx, `
		SELECT user_id FROM phone_identities WHERE phone_e164 = ? LIMIT 1
	`, session.Phone).Scan(&existingUser)
	if err == nil {
		return ActivationResult{}, ErrPhoneAlreadyActive
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return ActivationResult{}, fmt.Errorf("check active phone: %w", err)
	}

	userID, err := s.tokens.RandomURLToken(18)
	if err != nil {
		return ActivationResult{}, err
	}
	phoneID, err := s.tokens.RandomURLToken(18)
	if err != nil {
		return ActivationResult{}, err
	}

	now := s.now().UTC()
	tx, err := s.db.DB.BeginTx(ctx, nil)
	if err != nil {
		return ActivationResult{}, fmt.Errorf("begin activation transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO users (id, display_name, primary_email, status, created_at, updated_at)
		VALUES (?, ?, ?, 'active', ?, ?)
	`, userID, session.Name, session.EmailNormalized, stamp(now), stamp(now)); err != nil {
		return ActivationResult{}, fmt.Errorf("create user: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO phone_identities (id, user_id, phone_e164, verified_at, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, phoneID, userID, session.Phone, stamp(now), stamp(now)); err != nil {
		return ActivationResult{}, fmt.Errorf("create phone identity: %w", err)
	}

	for _, key := range starterProtocols {
		installID, err := s.tokens.RandomURLToken(18)
		if err != nil {
			return ActivationResult{}, err
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO protocol_installations (
				id, user_id, protocol_key, version, source, enabled, created_at
			) VALUES (?, ?, ?, 1, 'starter', 1, ?)
		`, installID, userID, key, stamp(now)); err != nil {
			return ActivationResult{}, fmt.Errorf("install starter protocol %s: %w", key, err)
		}
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE beta_invites
		SET status = 'activated', activated_user_id = ?, activated_at = ?, updated_at = ?
		WHERE id = ? AND status = 'challenge'
	`, userID, stamp(now), stamp(now), session.ID); err != nil {
		return ActivationResult{}, fmt.Errorf("mark invite activated: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE waitlist_entries
		SET status = 'activated', updated_at = ?
		WHERE id = ?
	`, stamp(now), session.WaitlistEntryID); err != nil {
		return ActivationResult{}, fmt.Errorf("mark waitlist activated: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return ActivationResult{}, fmt.Errorf("commit activation: %w", err)
	}
	if err := s.db.Push(ctx); err != nil {
		return ActivationResult{}, err
	}

	if err := s.sender.SendSMS(ctx, notifications.SMS{
		To: session.Phone,
		Text: "Hey " + session.Name + ". I'm Clear. You can text me like a person. What's something you've been putting off?",
	}); err != nil {
		slog.Warn("beta activated but welcome sms failed", "user_id", userID, "error", err)
	}

	return ActivationResult{UserID: userID}, nil
}

type activationSession struct {
	ID                  string
	WaitlistEntryID     string
	Status              string
	ActivationExpiresAt string
	ChallengeCodeHash   string
	ChallengeExpiresAt  string
	ChallengeSentAt     sql.NullString
	Attempts            int
	MaxAttempts         int
	Name                string
	EmailNormalized     string
	Phone               string
}

func (s *Service) loadActivation(ctx context.Context, token string) (activationSession, error) {
	var v activationSession
	if strings.TrimSpace(token) == "" {
		return v, ErrActivationNotFound
	}
	err := s.db.DB.QueryRowContext(ctx, `
		SELECT i.id, i.waitlist_entry_id, i.status,
		       i.activation_expires_at, i.challenge_code_hash,
		       i.challenge_expires_at, i.challenge_sent_at,
		       i.challenge_attempts, i.max_challenge_attempts,
		       w.name, w.email_normalized, w.phone_e164
		FROM beta_invites i
		JOIN waitlist_entries w ON w.id = i.waitlist_entry_id
		WHERE i.activation_token_hash = ?
		LIMIT 1
	`, s.tokens.Digest(token)).Scan(
		&v.ID, &v.WaitlistEntryID, &v.Status,
		&v.ActivationExpiresAt, &v.ChallengeCodeHash,
		&v.ChallengeExpiresAt, &v.ChallengeSentAt,
		&v.Attempts, &v.MaxAttempts,
		&v.Name, &v.EmailNormalized, &v.Phone,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return v, ErrActivationNotFound
	}
	if err != nil {
		return v, fmt.Errorf("load activation session: %w", err)
	}
	return v, nil
}

func expired(value string, now time.Time) bool {
	at, err := time.Parse(time.RFC3339Nano, value)
	return err != nil || !now.Before(at)
}

func stamp(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}

func maskPhone(phone string) string {
	if len(phone) <= 4 {
		return "••••"
	}
	return "•••" + phone[len(phone)-4:]
}
