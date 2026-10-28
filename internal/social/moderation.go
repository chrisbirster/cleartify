package social

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/chrisbirster/cleartify/internal/security"
)

var ErrInvalidModeration = errors.New("invalid moderation request")
var ErrUnlistedSubject = errors.New("subject not in local profile registry")

type Moderation struct {
	DB     *sql.DB
	Tokens *security.Tokens
	Push   func(context.Context) error
	Now    func() time.Time
}

func (m *Moderation) now() string {
	if m.Now != nil {
		return m.Now().UTC().Format(time.RFC3339Nano)
	}
	return time.Now().UTC().Format(time.RFC3339Nano)
}
func (m *Moderation) validate(ctx context.Context, actor, subject string) error {
	if !strings.HasPrefix(actor, "did:") || !strings.HasPrefix(subject, "did:") || actor == subject || len(subject) > 256 {
		return ErrInvalidModeration
	}
	var found int
	err := m.DB.QueryRowContext(ctx, "SELECT 1 FROM social_profiles WHERE did=?", subject).Scan(&found)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrUnlistedSubject
	}
	return err
}
func (m *Moderation) Block(ctx context.Context, actor, subject string, blocked bool) error {
	if err := m.validate(ctx, actor, subject); err != nil {
		return err
	}
	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if blocked {
		if _, err = tx.ExecContext(ctx, "INSERT INTO social_blocks(actor_did,subject_did,created_at) VALUES(?,?,?) ON CONFLICT(actor_did,subject_did) DO NOTHING", actor, subject, m.now()); err != nil {
			return err
		}
		// Unblocking must not silently restore an automatic first-friend edge.
		if _, err = tx.ExecContext(ctx, `INSERT INTO social_default_follow_optouts(follower_did,subject_did,opted_out_at) VALUES(?,?,?),(?,?,?) ON CONFLICT(follower_did,subject_did) DO NOTHING`, actor, subject, m.now(), subject, actor, m.now()); err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, `DELETE FROM social_follows WHERE (follower_did=? AND subject_did=?) OR (follower_did=? AND subject_did=?)`, actor, subject, subject, actor); err != nil {
			return err
		}
	} else {
		if _, err = tx.ExecContext(ctx, "DELETE FROM social_blocks WHERE actor_did=? AND subject_did=?", actor, subject); err != nil {
			return err
		}
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	if m.Push != nil {
		return m.Push(ctx)
	}
	return nil
}
func (m *Moderation) Report(ctx context.Context, actor, subject, reason string) (string, error) {
	if err := m.validate(ctx, actor, subject); err != nil {
		return "", err
	}
	switch reason {
	case "spam", "harassment", "impersonation", "other":
	default:
		return "", ErrInvalidModeration
	}
	id, err := m.Tokens.RandomURLToken(18)
	if err != nil {
		return "", err
	}
	_, err = m.DB.ExecContext(ctx, "INSERT INTO social_reports(id,reporter_did,subject_did,reason,created_at) VALUES(?,?,?,?,?)", id, actor, subject, reason, m.now())
	if err != nil {
		return "", err
	}
	return id, nil
}
