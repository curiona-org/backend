package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWT struct {
	secret    string
	expiresIn time.Duration
}

func NewJWT(secret string, expiresIn time.Duration) Authorizer {
	return JWT{secret: secret, expiresIn: expiresIn}
}

func (j JWT) Generate(id int) (string, error) {
	payload := NewPayload(id, j.expiresIn).Claims()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	return token.SignedString([]byte(j.secret))
}

func (j JWT) Parse(token string) (*Payload, error) {
	t, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(j.secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims := t.Claims.(jwt.MapClaims)

	return NewPayloadFromClaims(claims), nil
}
