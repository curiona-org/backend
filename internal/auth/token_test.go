package auth_test

import (
	"testing"
	"time"

	"github.com/curiona-org/backend/internal/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewToken(t *testing.T) {
	t.Parallel()

	secret := "test_secret"
	accountID := 1
	expiresIn := time.Hour

	token := auth.NewToken(secret, accountID, expiresIn)

	assert.NotNil(t, token)
	assert.Equal(t, accountID, token.AccountID())
	assert.Equal(t, "https://curiona.com", token.Issuer)
	assert.WithinDuration(t, time.Now(), token.IssuedAt(), time.Second)
	assert.WithinDuration(t, time.Now().Add(expiresIn), token.ExpiresAt(), time.Second)
}

func TestToken_Marshal(t *testing.T) {
	t.Parallel()

	secret := "test_secret"
	accountID := 1
	expiresIn := time.Hour

	token := auth.NewToken(secret, accountID, expiresIn)
	tokenStr, err := token.Marshal()

	require.NoError(t, err)
	assert.NotEmpty(t, tokenStr)
}

func TestTokenUnmarshal(t *testing.T) {
	t.Parallel()

	secret := "test_secret"
	accountID := 1
	expiresIn := time.Hour

	token := auth.NewToken(secret, accountID, expiresIn)
	tokenStr, err := token.Marshal()
	require.NoError(t, err)

	unmarshaledToken, err := auth.TokenUnmarshal(secret, tokenStr)
	require.NoError(t, err)

	assert.Equal(t, token.AccountID(), unmarshaledToken.AccountID())
	assert.Equal(t, token.Issuer, unmarshaledToken.Issuer)
	assert.WithinDuration(t, token.IssuedAt(), unmarshaledToken.IssuedAt(), time.Second)
	assert.WithinDuration(t, token.ExpiresAt(), unmarshaledToken.ExpiresAt(), time.Second)
}

func TestToken_IsExpired(t *testing.T) {
	t.Parallel()

	secret := "test_secret"
	accountID := 1
	expiresIn := time.Second * 1

	token := auth.NewToken(secret, accountID, expiresIn)
	assert.False(t, token.IsExpired())

	time.Sleep(time.Second * 2)
	assert.True(t, token.IsExpired())
}
