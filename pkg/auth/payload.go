package auth

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	AuthCtxKey = "identity"
)

type Payload struct {
	ID        int       `json:"id"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

func NewPayload(id int, expiresIn time.Duration) *Payload {
	return &Payload{
		ID:        id,
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(expiresIn),
	}
}

func NewPayloadFromClaims(claims jwt.MapClaims) *Payload {
	iat := int64(claims["iat"].(float64))
	exp := int64(claims["exp"].(float64))
	return &Payload{
		ID:        int(claims["id"].(float64)),
		IssuedAt:  time.Unix(iat, 0),
		ExpiresAt: time.Unix(exp, 0),
	}
}

// FromContext extracts the auth payload
func FromContext(ctx context.Context) *Payload {
	return ctx.Value(AuthCtxKey).(*Payload)
}

func (p *Payload) Claims() jwt.Claims {
	return jwt.MapClaims{
		"id":  p.ID,
		"iat": p.IssuedAt.Unix(),
		"exp": p.ExpiresAt.Unix(),
	}
}

func (p *Payload) Valid() bool {
	return time.Now().Before(p.ExpiresAt)
}
