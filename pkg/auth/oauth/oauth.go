package oauth

import (
	"context"
	"errors"
)

type User struct {
	Sub    string `json:"sub"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

type Client interface {
	Verify(ctx context.Context, token string) (User, error)
}

type Provider string

const (
	Google Provider = "google"
)

func NewProvider(provider Provider) (Client, error) {
	switch provider {
	case Google:
		return NewGoogleProvider(), nil
	default:
		return nil, errors.New("unknown oauth provider")
	}
}
