package object_test

import (
	"testing"

	"github.com/roadmap-thesis/backend/internal/domain/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPassword_Validate(t *testing.T) {
	t.Parallel()

	t.Run("ValidCharacters", func(t *testing.T) {
		t.Parallel()
		p := object.Password("")
		err := p.Validate("validPassword123")
		require.NoError(t, err)
	})

	t.Run("InvalidCharacters", func(t *testing.T) {
		t.Parallel()
		p := object.Password("")
		err := p.Validate("invalidPassword123😊")
		require.Error(t, err)
		assert.Equal(t, object.ErrPasswordInvalidCharacters, err)
	})
}

func TestPassword_Hash(t *testing.T) {
	t.Parallel()

	t.Run("HashSuccess", func(t *testing.T) {
		t.Parallel()
		p := object.Password("")
		hashedPassword, err := p.Hash("password123")
		require.NoError(t, err)
		assert.NotEmpty(t, hashedPassword)
	})

	t.Run("HashEmptyPassword", func(t *testing.T) {
		t.Parallel()
		p := object.Password("")
		hashedPassword, err := p.Hash("")
		require.Error(t, err)
		assert.Empty(t, hashedPassword)
		assert.Equal(t, object.ErrPasswordEmpty, err)
	})
}

func TestPassword_Compare(t *testing.T) {
	t.Parallel()

	t.Run("CompareSuccess", func(t *testing.T) {
		t.Parallel()
		p := object.Password("")
		hashedPassword, err := p.Hash("password123")
		require.NoError(t, err)

		isMatch := hashedPassword.Compare("password123")
		assert.True(t, isMatch)
	})

	t.Run("CompareFailure", func(t *testing.T) {
		t.Parallel()
		p := object.Password("")
		hashedPassword, err := p.Hash("password123")
		require.NoError(t, err)

		isMatch := hashedPassword.Compare("wrongpassword")
		assert.False(t, isMatch)
	})
}

func TestPassword_String(t *testing.T) {
	t.Parallel()

	t.Run("String", func(t *testing.T) {
		t.Parallel()
		p := object.Password("password123")
		assert.Equal(t, "password123", p.String())
	})
}
