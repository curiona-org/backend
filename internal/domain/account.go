package domain

import (
	"errors"
	"time"

	"github.com/curiona-org/backend/internal/auth"
)

const (
	AccountTable = "accounts"
)

var (
	ErrAccountNotFound = errors.New("account not found")
)

type Account struct {
	ID             int
	Email          string
	password       auth.Password
	PasswordDigest string
	Method         auth.Method

	Profile  *Profile
	Roadmaps []*Roadmap

	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewAccount(email string, password auth.Password, provider auth.Method, profile *Profile) (*Account, error) {
	account := &Account{
		Email:     email,
		password:  password,
		Method:    provider,
		Profile:   profile,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return account, nil
}

func (e *Account) IsZero() bool {
	return e.ID == 0 &&
		e.Email == "" &&
		e.PasswordDigest == "" &&
		(e.Profile == nil || e.Profile.IsZero()) &&
		len(e.Roadmaps) == 0 &&
		e.CreatedAt.IsZero() &&
		e.UpdatedAt.IsZero()
}

// HashPassword hashes the password.
func (e *Account) HashPassword() error {
	hash, err := e.password.Hash()
	if err != nil {
		return err
	}
	e.PasswordDigest = hash
	return nil
}

func (e *Account) CheckPassword(password auth.Password) bool {
	return password.Compare(e.PasswordDigest)
}

func (e *Account) SetProfile(profile *Profile) {
	e.Profile = profile
}

func (e *Account) UpdateEmail(email string) {
	e.Email = email
	e.UpdateChangelog()
}

func (e *Account) UpdateChangelog() {
	e.UpdatedAt = time.Now()
}
