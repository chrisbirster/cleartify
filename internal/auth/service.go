package auth

import (
	"context"
	"net/http"

	"cleartify/internal/domain"
)

const SessionCookieName = "cleartify_session"

type Service interface {
	CurrentSession(ctx context.Context, r *http.Request) (domain.AuthSession, error)
}
