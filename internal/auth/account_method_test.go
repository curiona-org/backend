package auth_test

import (
	"testing"

	"github.com/curiona-org/backend/internal/auth"
	"github.com/stretchr/testify/assert"
)

func TestMethod_IsEmail(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		method   auth.Method
		expected bool
	}{
		{"EmailMethod", auth.MethodEmail, true},
		{"GoogleMethod", auth.MethodGoogle, false},
		{"EmptyMethod", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := tc.method.IsEmail()
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestMethod_IsGoogle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		method   auth.Method
		expected bool
	}{
		{"EmailMethod", auth.MethodEmail, false},
		{"GoogleMethod", auth.MethodGoogle, true},
		{"EmptyMethod", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := tc.method.IsGoogle()
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestMethod_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		method   auth.Method
		expected string
	}{
		{"EmailMethod", auth.MethodEmail, "email"},
		{"GoogleMethod", auth.MethodGoogle, "google"},
		{"EmptyMethod", "", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := tc.method.String()
			assert.Equal(t, tc.expected, result)
		})
	}
}
