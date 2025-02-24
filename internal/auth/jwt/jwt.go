package jwt

import (
	"errors"
	"fmt"
	"time"

	basejwt "github.com/golang-jwt/jwt/v5"
)

type JWT struct {
	secret    string
	expiresIn time.Duration
}

type MapClaims = basejwt.MapClaims

func NewJWT(secret string, expiresIn time.Duration) JWT {
	return JWT{secret: secret, expiresIn: expiresIn}
}

func (j JWT) Generate(claims basejwt.MapClaims) (string, error) {
	token := basejwt.NewWithClaims(basejwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secret))
}

func (j JWT) Parse(token string) (basejwt.MapClaims, error) {
	t, err := basejwt.Parse(token, func(t *basejwt.Token) (any, error) {
		if _, ok := t.Method.(*basejwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(j.secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := t.Claims.(basejwt.MapClaims)
	if !ok || !t.Valid || t.Header["alg"] != "HS256" {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

func (j JWT) ExpiresAt() time.Time {
	return time.Now().Add(j.expiresIn)
}

func (j JWT) ExpiresIn() time.Duration {
	return j.expiresIn
}
