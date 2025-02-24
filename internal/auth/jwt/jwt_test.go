package jwt_test

import (
	"testing"
	"time"

	"github.com/curiona-org/backend/internal/auth/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Parallel()

	secret := "test_secret"
	expiresIn := time.Hour

	j := jwt.New(secret, expiresIn)

	assert.NotNil(t, j)
}

func TestJWT_Marshal(t *testing.T) {
	t.Parallel()

	secret := "test_secret"
	expiresIn := time.Hour
	j := jwt.New(secret, expiresIn)

	claims := jwt.MapClaims{
		"sub":   "1234567890",
		"name":  "John Doe",
		"admin": true,
		"exp":   time.Now().Add(expiresIn).Unix(),
	}

	token, err := j.Marshal(claims)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestJWT_Unmarshal(t *testing.T) {
	t.Parallel()

	secret := "test_secret"
	expiresIn := time.Hour
	j := jwt.New(secret, expiresIn)

	claims := jwt.MapClaims{
		"sub":   "1234567890",
		"name":  "John Doe",
		"admin": true,
		"exp":   time.Now().Add(expiresIn).Unix(),
	}

	token, err := j.Marshal(claims)
	require.NoError(t, err)

	parsedClaims := make(jwt.MapClaims)
	err = j.Unmarshal(token, parsedClaims)
	require.NoError(t, err)
	assert.Equal(t, claims["sub"], parsedClaims["sub"])
	assert.Equal(t, claims["name"], parsedClaims["name"])
	assert.Equal(t, claims["admin"], parsedClaims["admin"])
}

func TestJWT_UnmarshalInvalidToken(t *testing.T) {
	t.Parallel()

	secret := "test_secret"
	expiresIn := time.Hour
	j := jwt.New(secret, expiresIn)

	err := j.Unmarshal("invalid_token", nil)
	assert.Error(t, err)
}
