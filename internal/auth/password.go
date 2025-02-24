package auth

import (
	"errors"
	"unicode"

	"github.com/curiona-org/backend/internal/cerrors"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrPasswordInvalidCharacters is returned when the password contains invalid characters.
	ErrPasswordInvalidCharacters = cerrors.Wrap(cerrors.InvalidData(), errors.New("password invalid characters"))

	// ErrPasswordEmpty is returned when the password is empty.
	ErrPasswordEmpty = cerrors.Wrap(cerrors.InvalidData(), errors.New("password empty"))
)

const (
	// CostFactor is the amount of work required to check a password against a bcrypt.
	CostFactor = 10
)

// Password represents a plain password.
type Password string

func NewPassword(password string) Password {
	return Password(password)
}

// Validate validates a plain password.
func (p Password) Validate() error {
	if p == "" {
		return ErrPasswordEmpty
	}

	if !p.validateCharacters() {
		return ErrPasswordInvalidCharacters
	}

	return nil
}

func (p Password) validateCharacters() bool {
	for _, char := range p {
		if char > unicode.MaxASCII {
			return false
		}
	}

	return true
}

// Hash generates a hash for the password.
func (p Password) Hash() (string, error) {
	if err := p.Validate(); err != nil {
		return "", err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(p), 10)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

// Compare returns wether the plaintext password matches the password digest.
func (p Password) Compare(digest string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(digest), []byte(p))
	switch {
	case err == nil:
		return true
	case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
		return false
	default:
		log.Error().Err(err).Msg("failed to compare password")
		return false
	}
}

func (p Password) String() string {
	return string(p)
}
