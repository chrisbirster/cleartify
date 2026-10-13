package waitlist

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/chrisbirster/cleartify/internal/database"
)

var (
	ErrPhoneInUse = errors.New("phone number is already attached to another waitlist entry")
	e164Pattern   = regexp.MustCompile(`^\+[1-9][0-9]{7,14}$`)
)

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

type Signup struct {
	Name         string
	Email        string
	Phone        string
	EmailUpdates bool
	Source       string
	AcceptPolicies bool
}

type Service struct {
 ConsentKeysEnabled bool
	db *database.Store
}

func NewService(db *database.Store) *Service {
	return &Service{db: db}
}

const CurrentPolicyVersion = "2026-10-12-draft"

func (s *Service) Join(ctx context.Context, input Signup) (string, error) {
	if !input.AcceptPolicies { return "", &ValidationError{Field:"acceptPolicies",Message:"accept terms and privacy notice"} }
	name := strings.TrimSpace(input.Name)
	if name == "" || len(name) > 100 {
		return "", &ValidationError{Field: "name", Message: "enter a name between 1 and 100 characters"}
	}

	email, err := normalizeEmail(input.Email)
	if err != nil {
		return "", &ValidationError{Field: "email", Message: "enter a valid email address"}
	}

	phone, err := normalizePhone(input.Phone)
	if err != nil {
		return "", &ValidationError{Field: "phone", Message: "enter a valid mobile number"}
	}

	source := strings.TrimSpace(input.Source)
	if source == "" {
		source = "landing"
	}
	if len(source) > 64 {
		source = source[:64]
	}

	now := time.Now().UTC().Format(time.RFC3339Nano)
	tx, err := s.db.DB.BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("begin waitlist transaction: %w", err)
	}
	defer tx.Rollback()

	var otherID string
	err = tx.QueryRowContext(ctx,
		`SELECT id FROM waitlist_entries
		 WHERE phone_e164 = ? AND email_normalized <> ?
		 LIMIT 1`,
		phone, email,
	).Scan(&otherID)
	switch {
	case err == nil:
		return "", ErrPhoneInUse
	case !errors.Is(err, sql.ErrNoRows):
		return "", fmt.Errorf("check phone uniqueness: %w", err)
	}

	id, err := randomID()
	if err != nil {
		return "", err
	}

	emailUpdates := 0
	if input.EmailUpdates {
		emailUpdates = 1
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO waitlist_entries (
			id, name, email, email_normalized, phone_e164,
			email_updates, status, source, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, 'waiting', ?, ?, ?)
		ON CONFLICT(email_normalized) DO UPDATE SET
			name = excluded.name,
			email = excluded.email,
			phone_e164 = excluded.phone_e164,
			email_updates = excluded.email_updates,
			source = excluded.source,
			updated_at = excluded.updated_at
	`, id, name, strings.TrimSpace(input.Email), email, phone, emailUpdates, source, now, now)
	if err != nil {
		return "", fmt.Errorf("upsert waitlist entry: %w", err)
	}

	if err := tx.QueryRowContext(ctx,
		`SELECT id FROM waitlist_entries WHERE email_normalized = ?`,
		email,
	).Scan(&id); err != nil {
		return "", fmt.Errorf("load waitlist entry id: %w", err)
	}

	consentID, err := randomID()
	if err != nil {
		return "", err
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO consents (
			id, waitlist_entry_id, kind, granted, source, recorded_at
		) VALUES (?, ?, 'email_marketing', ?, ?, ?)
	`, consentID, id, emailUpdates, source, now); err != nil {
		return "", fmt.Errorf("record marketing consent: %w", err)
	}

	acceptanceID,err := randomID()
	if err != nil {return "",err}
	if _,err := tx.ExecContext(ctx, `INSERT INTO policy_acceptances (id,waitlist_entry_id,policy_version,accepted_at) VALUES(?,?,?,?)`, acceptanceID,id,CurrentPolicyVersion,now);err != nil {return "",fmt.Errorf("record policy acceptance: %w",err)}
	if s.ConsentKeysEnabled {
  for _,record:=range []struct{key,version string;accepted int}{
   {"policy_acceptance","policy_version_pending",1},
   {"email_marketing","marketing_consent_version_pending",emailUpdates},
  }{
   consentKeyID,e:=randomID();if e!=nil{return "",e}
   _,e=tx.ExecContext(ctx,`INSERT INTO signup_consent_keys(waitlist_entry_id,consent_key,consent_version_key,accepted,recorded_at) VALUES(?,?,?,?,?)`,id,record.key,record.version,record.accepted,now)
   _=consentKeyID
   if e!=nil{return "",fmt.Errorf("record consent key: %w",e)}
  }
 }
 if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("commit waitlist transaction: %w", err)
	}
	if err := s.db.Push(ctx); err != nil {
		return "", err
	}

	return id, nil
}

func normalizeEmail(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > 254 {
		return "", errors.New("invalid email")
	}
	addr, err := mail.ParseAddress(value)
	if err != nil || !strings.EqualFold(addr.Address, value) {
		return "", errors.New("invalid email")
	}
	parts := strings.Split(addr.Address, "@")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", errors.New("invalid email")
	}
	return strings.ToLower(addr.Address), nil
}

func normalizePhone(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", errors.New("invalid phone")
	}

	var b strings.Builder
	for i, r := range value {
		switch {
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '+' && i == 0:
			b.WriteRune(r)
		case r == ' ' || r == '-' || r == '(' || r == ')' || r == '.':
			continue
		default:
			return "", errors.New("invalid phone")
		}
	}
	phone := b.String()
	if strings.HasPrefix(phone, "00") {
		phone = "+" + strings.TrimPrefix(phone, "00")
	}
	if !strings.HasPrefix(phone, "+") && len(phone) == 10 {
		phone = "+1" + phone
	}
	if !e164Pattern.MatchString(phone) {
		return "", errors.New("invalid phone")
	}
	return phone, nil
}

func randomID() (string, error) {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", fmt.Errorf("generate id: %w", err)
	}
	return hex.EncodeToString(buf[:]), nil
}
