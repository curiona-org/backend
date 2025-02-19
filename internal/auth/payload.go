package auth

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type authContextKey string

const (
	// AuthCtxKey is the key used to store the auth payload in a context.
	AuthCtxKey authContextKey = "auth_payload"
)

func (k authContextKey) String() string {
	return string(k)
}

// Payload represents a token payload.
type Payload struct {
	ID        int       `json:"id"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// NewPayload creates a new payload.
func NewPayload(id int, expiresIn time.Duration) *Payload {
	return &Payload{
		ID:        id,
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(expiresIn),
	}
}

// NewPayloadFromClaims creates a new payload from jwt claims.
func NewPayloadFromClaims(claims jwt.MapClaims) *Payload {
	iat, ok := claims["iat"].(float64)
	if !ok {
		return nil
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return nil
	}

	id, ok := claims["id"].(float64)
	if !ok {
		return nil
	}

	return &Payload{
		ID:        int(id),
		IssuedAt:  time.Unix(int64(iat), 0),
		ExpiresAt: time.Unix(int64(exp), 0),
	}
}

// FromContext extracts the auth payload.
func FromContext(ctx context.Context) *Payload {
	payload, ok := ctx.Value(AuthCtxKey).(*Payload)
	if !ok {
		return nil
	}

	return payload
}

// Claims returns the jwt token claims.
func (p *Payload) Claims() jwt.Claims {
	return jwt.MapClaims{
		"id":  p.ID,
		"iat": p.IssuedAt.Unix(),
		"exp": p.ExpiresAt.Unix(),
	}
}

// Valid checks if the payload is valid.
func (p *Payload) Valid() bool {
	return time.Now().Before(p.ExpiresAt)
}
