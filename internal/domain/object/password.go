package object

import (
	"errors"
	"unicode"

	"github.com/roadmap-thesis/backend/internal/apperrors"
	"github.com/roadmap-thesis/backend/internal/crypto"
)

var (
	// ErrPasswordInvalidCharacters is returned when the password contains invalid characters.
	ErrPasswordInvalidCharacters = apperrors.Wrap(apperrors.InvalidData(), errors.New("password invalid characters"))

	// ErrPasswordEmpty is returned when the password is empty.
	ErrPasswordEmpty = apperrors.Wrap(apperrors.InvalidData(), errors.New("password empty"))
)

// Password can be plain/hashed.
type Password string

// Validate validates a plain password.
func (p Password) Validate(plain string) error {
	if plain == "" {
		return ErrPasswordEmpty
	}

	if !p.validateCharacters(plain) {
		return ErrPasswordInvalidCharacters
	}

	return nil
}

func (p Password) validateCharacters(plain string) bool {
	for _, char := range plain {
		if char > unicode.MaxASCII {
			return false
		}
	}

	return true
}

// Hash generates a hash for the password.
func (p Password) Hash(plain string) (Password, error) {
	if err := p.Validate(plain); err != nil {
		return "", err
	}

	hash, err := crypto.BcryptHash(plain)
	if err != nil {
		return "", err
	}

	return Password(hash), nil
}

// Compare compares the password with the hash.
func (p Password) Compare(plain string) bool {
	return crypto.BcryptCompare(string(p), plain)
}

func (p Password) String() string {
	return string(p)
}
