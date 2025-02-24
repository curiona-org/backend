package jwt_test

import (
	"testing"
	"time"

	"github.com/curiona-org/backend/internal/auth/jwt"
	basejwt "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewJWT(t *testing.T) {
	t.Parallel()

	secret := "test_secret"
	expiresIn := time.Hour

	j := jwt.NewJWT(secret, expiresIn)

	assert.NotNil(t, j)
	assert.Equal(t, expiresIn, j.ExpiresIn())
}

func TestJWT_Generate(t *testing.T) {
	t.Parallel()

	secret := "test_secret"
	expiresIn := time.Hour
	j := jwt.NewJWT(secret, expiresIn)

	claims := basejwt.MapClaims{
		"sub":   "1234567890",
		"name":  "John Doe",
		"admin": true,
		"exp":   time.Now().Add(expiresIn).Unix(),
	}

	token, err := j.Generate(claims)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestJWT_Parse(t *testing.T) {
	t.Parallel()

	secret := "test_secret"
	expiresIn := time.Hour
	j := jwt.NewJWT(secret, expiresIn)

	claims := basejwt.MapClaims{
		"sub":   "1234567890",
		"name":  "John Doe",
		"admin": true,
		"exp":   time.Now().Add(expiresIn).Unix(),
	}

	token, err := j.Generate(claims)
	require.NoError(t, err)

	parsedClaims, err := j.Parse(token)
	require.NoError(t, err)
	assert.Equal(t, claims["sub"], parsedClaims["sub"])
	assert.Equal(t, claims["name"], parsedClaims["name"])
	assert.Equal(t, claims["admin"], parsedClaims["admin"])
}

func TestJWT_ParseInvalidToken(t *testing.T) {
	t.Parallel()

	secret := "test_secret"
	expiresIn := time.Hour
	j := jwt.NewJWT(secret, expiresIn)

	_, err := j.Parse("invalid_token")
	assert.Error(t, err)
}

func TestJWT_ExpiresAt(t *testing.T) {
	t.Parallel()

	secret := "test_secret"
	expiresIn := time.Hour
	j := jwt.NewJWT(secret, expiresIn)

	expectedExpiresAt := time.Now().Add(expiresIn)
	assert.WithinDuration(t, expectedExpiresAt, j.ExpiresAt(), time.Second)
}
