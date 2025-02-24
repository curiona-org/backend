package auth

import (
	"context"
	"errors"
	"time"

	"github.com/curiona-org/backend/internal/auth/jwt"
	"github.com/curiona-org/backend/pkg/cerrors"
)

type Token interface {
	Marshal() (string, error)
	IsExpired() bool

	AccountID() int
	Issuer() string
	IssuedAt() time.Time
	ExpiresAt() time.Time
	ExpiresIn() time.Duration
}

type token struct {
	secret string

	accountID int
	issuer    string
	issuedAt  time.Time
	expiresAt time.Time
}

var _ Token = (*token)(nil)

func NewToken(secret string, accountID int, expiresIn time.Duration) Token {
	token := &token{
		secret:    secret,
		accountID: accountID,
		issuer:    "https://curiona.com",
		issuedAt:  time.Now(),
		expiresAt: time.Now().Add(expiresIn),
	}

	return token
}

func FromContext(ctx context.Context) Token {
	return ctx.Value(ContextKey).(Token)
}

// Marshal generates a JWT token from the token struct
func (t *token) Marshal() (string, error) {
	claims := jwt.MapClaims{
		"account_id": t.accountID,
		"iss":        t.issuer,
		"iat":        t.issuedAt.Unix(),
		"exp":        t.expiresAt.Unix(),
	}

	jwt := jwt.NewJWT(t.secret, t.expiresAt.Sub(t.issuedAt))
	return jwt.Generate(claims)
}

// TokenUnmarshal parses a JWT token and returns a token struct
func TokenUnmarshal(secret, tokenStr string) (*token, error) {
	jwt := jwt.NewJWT(secret, 0)
	claims, err := jwt.Parse(tokenStr)
	if err != nil {
		return nil, err
	}

	accountID, ok := claims["account_id"].(float64)
	if !ok {
		return nil, cerrors.Wrap(cerrors.InvalidCredentials(), errors.New("token malformed"))
	}

	issuer, ok := claims["iss"].(string)
	if !ok {
		return nil, cerrors.Wrap(cerrors.InvalidCredentials(), errors.New("token malformed"))
	}

	issuedAt, ok := claims["iat"].(float64)
	if !ok {
		return nil, cerrors.Wrap(cerrors.InvalidCredentials(), errors.New("token malformed"))
	}

	expiresAt, ok := claims["exp"].(float64)
	if !ok {
		return nil, cerrors.Wrap(cerrors.InvalidCredentials(), errors.New("token malformed"))
	}

	return &token{
		accountID: int(accountID),
		issuer:    issuer,
		issuedAt:  time.Unix(int64(issuedAt), 0),
		expiresAt: time.Unix(int64(expiresAt), 0),
	}, nil
}

func (t *token) AccountID() int {
	return t.accountID
}

func (t *token) Issuer() string {
	return t.issuer
}

func (t *token) IssuedAt() time.Time {
	return t.issuedAt
}

func (t *token) ExpiresAt() time.Time {
	return t.expiresAt
}

func (t *token) ExpiresIn() time.Duration {
	return t.expiresAt.Sub(t.issuedAt)
}

func (t *token) IsExpired() bool {
	return time.Now().After(t.expiresAt)
}
