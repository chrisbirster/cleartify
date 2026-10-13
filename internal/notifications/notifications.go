package notifications

import (
	"context"
	"errors"
	"log/slog"
)

var ErrUnavailable = errors.New("notification delivery is not configured")

type Email struct {
	To      string
	Subject string
	Text    string
}

type SMS struct {
	To   string
	Text string
}

type Sender interface {
	SendEmail(context.Context, Email) error
	SendSMS(context.Context, SMS) error
}

type DisabledSender struct{}

func (DisabledSender) SendEmail(context.Context, Email) error { return ErrUnavailable }
func (DisabledSender) SendSMS(context.Context, SMS) error     { return ErrUnavailable }

// DebugSender is for local development only. It intentionally prints message bodies,
// which can include beta activation secrets. Never enable it in production.
type DebugSender struct{}

func (DebugSender) SendEmail(_ context.Context, message Email) error {
	slog.Warn("DEBUG notification email", "to", message.To, "subject", message.Subject, "text", message.Text)
	return nil
}

func (DebugSender) SendSMS(_ context.Context, message SMS) error {
	slog.Warn("DEBUG notification sms", "to", message.To, "text", message.Text)
	return nil
}
