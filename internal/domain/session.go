package domain

import (
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

// Renew renews the session with a new refresh token and expiration time.
func (s *Session) Renew(refreshToken, userAgent, clientIP string, expiresAt time.Time) {
	s.RefreshToken = refreshToken
	s.UserAgent = userAgent
	s.ClientIP = clientIP
	s.Blocked = false
	s.ExpiresAt = expiresAt
}
