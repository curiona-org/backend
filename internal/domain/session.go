package domain

import (
	"context"
	"errors"
	"time"
)

const (
	SessionTable = "sessions"
)

var (
	ErrSessionNotFound  = errors.New("session not found")
	ErrSessionIsBlocked = errors.New("session is blocked")
	ErrSessionExpired   = errors.New("session expired")
)

// Session represents a user session.
type Session struct {
	ID           int
	AccountID    int
	RefreshToken string
	UserAgent    string
	ClientIP     string
	Blocked      bool
	ExpiresAt    time.Time
	CreatedAt    time.Time
}

type SessionRepository interface {
	GetByAccountID(ctx context.Context, accountID int) (Session, error)
	Save(ctx context.Context, input *Session) (Session, error)
	RotateRefreshToken(ctx context.Context, refreshToken string, updateFn func(context.Context, *Session) (bool, error)) error
	Delete(ctx context.Context, id int) error
}

// NewSession creates a new session for the given account.
func NewSession(accountID int, refreshToken, userAgent, clientIP string, expiresAt time.Time) *Session {
	return &Session{
		AccountID:    accountID,
		RefreshToken: refreshToken,
		UserAgent:    userAgent,
		ClientIP:     clientIP,
		Blocked:      false,
		ExpiresAt:    expiresAt,
		CreatedAt:    time.Now(),
	}
}

// MarkAsBlocked marks the session as blocked.
func (s *Session) MarkAsBlocked() {
	s.Blocked = true
}
