package io

import (
	"time"

	"github.com/curiona-org/backend/internal/auth"
)

type AuthInput struct {
	Name       string `json:"name" validate:"omitempty"`
	Avatar     string `json:"avatar" validate:"omitempty,url"`
	Email      string `json:"email" validate:"required_without=OAuthToken,omitempty,email"`
	Password   string `json:"password" validate:"required_without=OAuthToken,omitempty,min=6"`
	OAuthToken string `json:"oauth_token" validate:"required_without_all=Email Password,omitempty"`

	UserAgent string `json:"-"`
	ClientIP  string `json:"-"`

	Method              auth.Method `json:"-"`
	IgnorePasswordCheck bool        `json:"-"` // used to skip password check in some cases
}

type AuthOutput struct {
	Created               bool              `json:"created"`
	AccessToken           string            `json:"access_token"`
	AccessTokenExpiresAt  time.Time         `json:"access_token_expires_at"`
	RefreshToken          string            `json:"-"`
	RefreshTokenExpiresIn int               `json:"-"` // in seconds
	RefreshTokenExpiresAt time.Time         `json:"-"`
	Account               AuthOutputAccount `json:"account"`
}

type AuthOutputAccount struct {
	ID       int       `json:"id"`
	Email    string    `json:"email"`
	Name     string    `json:"name"`
	Avatar   string    `json:"avatar"`
	JoinedAt time.Time `json:"joined_at"`
}
