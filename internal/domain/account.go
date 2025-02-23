package domain

import (
	"errors"
	"time"

	"github.com/curiona-org/backend/internal/domain/object"
)

const (
	AccountTable = "accounts"
)

var (
	ErrAccountNotFound = errors.New("account not found")
)

type Account struct {
	ID       int
	Email    string
	Password object.Password
	Provider object.AccountProvider

	Profile  *Profile
	Roadmaps []*Roadmap

	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewAccount(email, plainPassword string, provider object.AccountProvider, profile *Profile) (*Account, error) {
	password := object.Password(plainPassword)

	if err := password.Validate(plainPassword); err != nil {
		return nil, err
	}

	hash, err := password.Hash(plainPassword)
	if err != nil {
		return nil, err
	}

	account := &Account{
		Email:     email,
		Password:  hash,
		Provider:  provider,
		Profile:   profile,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return account, nil
}

func (e *Account) IsZero() bool {
	return e.ID == 0 &&
		e.Email == "" &&
		e.Password == "" &&
		(e.Profile == nil || e.Profile.IsZero()) &&
		len(e.Roadmaps) == 0 &&
		e.CreatedAt.IsZero() &&
		e.UpdatedAt.IsZero()
}

func (e *Account) CheckPassword(password string) bool {
	return e.Password.Compare(password)
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
