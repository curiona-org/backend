package jwt

import (
	"fmt"
	"time"

	"maps"

	basejwt "github.com/golang-jwt/jwt/v5"
)

type JWT struct {
	secret    string
	expiresIn time.Duration
}

// MapClaims is a type alias for jwt.MapClaims compatible with the golang-jwt library
// and also a type alias for map[string]any.
type MapClaims = basejwt.MapClaims

func New(secret string, expiresIn time.Duration) JWT {
	return JWT{secret: secret, expiresIn: expiresIn}
}

func (j JWT) Marshal(claims map[string]any) (string, error) {
	token := basejwt.NewWithClaims(basejwt.SigningMethodHS256, MapClaims(claims))
	return token.SignedString([]byte(j.secret))
}

func (j JWT) Unmarshal(token string, out map[string]any) error {
	t, err := basejwt.Parse(token, func(t *basejwt.Token) (any, error) {
		if _, ok := t.Method.(*basejwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(j.secret), nil
	})
	if err != nil {
		return err
	}

	claims, ok := t.Claims.(MapClaims)
	if !ok || !t.Valid || t.Header["alg"] != "HS256" {
		return ErrInvalidToken
	}

	maps.Copy(out, claims)

	return nil
}
