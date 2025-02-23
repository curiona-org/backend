package auth_test

import (
	"testing"
	"time"

	"github.com/roadmap-thesis/backend/pkg/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuth_Token(t *testing.T) {
	t.Parallel()
	t.Run("CreateToken", func(t *testing.T) {
		t.Parallel()
		id := 1

		jwt := auth.NewJWT("secret", time.Hour)
		token, err := jwt.Generate(id)
		require.NoError(t, err)
		assert.NotEmpty(t, token)
	})

	t.Run("VerifyToken", func(t *testing.T) {
		t.Parallel()
		id := 1

		jwt := auth.NewJWT("secret", time.Hour)
		token, err := jwt.Generate(id)
		require.NoError(t, err)
		assert.NotEmpty(t, token)

		payload, err := jwt.Parse(token)
		require.NoError(t, err)
		assert.NotNil(t, payload)
		assert.Equal(t, id, payload.ID)
	})

	t.Run("ExpiredToken", func(t *testing.T) {
		t.Parallel()
		id := 1

		// Create a token with a short expiration time
		jwt := auth.NewJWT("secret", time.Second*1)
		token, err := jwt.Generate(id)
		require.NoError(t, err)
		assert.NotEmpty(t, token)

		// Wait for the token to expire
		time.Sleep(time.Second * 2)

		payload, err := jwt.Parse(token)
		require.Error(t, err)
		assert.Nil(t, payload)
	})
}
