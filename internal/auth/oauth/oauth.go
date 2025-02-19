package oauth

import (
	"context"
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
