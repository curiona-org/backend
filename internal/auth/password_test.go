package auth_test

import (
	"testing"

	"github.com/curiona-org/backend/internal/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPassword_Validate(t *testing.T) {
	t.Parallel()

	t.Run("ValidCharacters", func(t *testing.T) {
		t.Parallel()
		p := auth.NewPassword("validPassword123")
		err := p.Validate()
		require.NoError(t, err)
	})

	t.Run("InvalidCharacters", func(t *testing.T) {
		t.Parallel()
		p := auth.NewPassword("invalidPassword123😊")
		err := p.Validate()
		require.Error(t, err)
		assert.Equal(t, auth.ErrPasswordInvalidCharacters, err)
	})
}

func TestPassword_Hash(t *testing.T) {
	t.Parallel()

	t.Run("HashSuccess", func(t *testing.T) {
		t.Parallel()
		p := auth.NewPassword("password123")
		hashedPassword, err := p.Hash()
		require.NoError(t, err)
		assert.NotEmpty(t, hashedPassword)
	})

	t.Run("HashEmptyPassword", func(t *testing.T) {
		t.Parallel()
		p := auth.Password("")
		hashedPassword, err := p.Hash()
		require.Error(t, err)
		assert.Empty(t, hashedPassword)
		assert.Equal(t, auth.ErrPasswordEmpty, err)
	})
}

func TestPassword_Compare(t *testing.T) {
	t.Parallel()

	t.Run("CompareSuccess", func(t *testing.T) {
		t.Parallel()
		p := auth.NewPassword("password123")
		p2 := auth.NewPassword("password123")
		hashedPassword, err := p.Hash()
		require.NoError(t, err)

		isMatch := p2.Compare(hashedPassword)
		assert.True(t, isMatch)
	})

	t.Run("CompareFailure", func(t *testing.T) {
		t.Parallel()
		p := auth.NewPassword("password123")
		p2 := auth.NewPassword("password123123")
		hashedPassword, err := p.Hash()
		require.NoError(t, err)

		isMatch := p2.Compare(hashedPassword)
		assert.False(t, isMatch)
	})
}

func TestPassword_String(t *testing.T) {
	t.Parallel()

	t.Run("String", func(t *testing.T) {
		t.Parallel()
		p := auth.NewPassword("password123")
		assert.Equal(t, "password123", p.String())
	})
}
