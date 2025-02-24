package auth

import (
	"context"
	"errors"
	"time"

	"github.com/curiona-org/backend/internal/auth/jwt"
)

var (
	// ErrTokenInvalid is returned when the token is invalid.
	ErrTokenInvalid = errors.New("token is invalid")

	// ErrTokenExpired is returned when the token is expired.
	ErrTokenExpired = errors.New("token is expired")
)

type Token struct {
	secret string

	encoder TokenEncoder

	AccountID int
	Issuer    string
	IssuedAt  time.Time
	ExpiresAt time.Time
}

func NewToken(secret string, accountID int, expiresIn time.Duration) *Token {
	now := time.Now()
	token := &Token{
		secret:    secret,
		encoder:   jwt.New(secret, expiresIn),
		AccountID: accountID,
		Issuer:    "https://curiona.com",
		IssuedAt:  now,
		ExpiresAt: now.Add(expiresIn),
	}

	return token
}

// TokenFromContext retrieves the token from a context using the context key
func TokenFromContext(ctx context.Context) *Token {
	return ctx.Value(ContextKey).(*Token)
}

// Marshal generates a JWT token from the token struct
func (t *Token) Marshal() (string, error) {
	claims := map[string]any{
		"account_id": t.AccountID,
		"iss":        t.Issuer,
		"iat":        t.IssuedAt.Unix(),
		"exp":        t.ExpiresAt.Unix(),
	}

	return t.encoder.Marshal(claims)
}

// Unmarshal parses a JWT token and returns a token struct
func (t *Token) Unmarshal(tokenStr string) (*Token, error) {
	claims := make(map[string]any)
	err := t.encoder.Unmarshal(tokenStr, claims)
	if err != nil {
		return nil, err
	}

	accountID, ok := claims["account_id"].(float64)
	if !ok {
		return nil, ErrTokenInvalid
	}

	issuer, ok := claims["iss"].(string)
	if !ok {
		return nil, ErrTokenInvalid
	}

	issuedAt, ok := claims["iat"].(float64)
	if !ok {
		return nil, ErrTokenInvalid
	}

	expiresAt, ok := claims["exp"].(float64)
	if !ok {
		return nil, ErrTokenInvalid
	}

	token := &Token{
		AccountID: int(accountID),
		Issuer:    issuer,
		IssuedAt:  time.Unix(int64(issuedAt), 0),
		ExpiresAt: time.Unix(int64(expiresAt), 0),
	}

	if token.IsExpired() {
		return nil, ErrTokenExpired
	}

	return token, nil
}

func (t *Token) ExpiresIn() time.Duration {
	return t.ExpiresAt.Sub(t.IssuedAt)
}

func (t *Token) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}
