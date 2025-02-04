package auth

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/roadmap-thesis/backend/pkg/config"
)

func CreateToken(id int) (string, error) {
	payload := NewPayload(id, config.JWTSecretExpiresIn()).Claims()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	return token.SignedString([]byte(config.JWTSecretKey()))
}

func VerifyToken(token string) (*Payload, error) {
	t, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(config.JWTSecretKey()), nil
	})
	if err != nil {
		return nil, err
	}

	claims := t.Claims.(jwt.MapClaims)

	return NewPayloadFromClaims(claims), nil
}
