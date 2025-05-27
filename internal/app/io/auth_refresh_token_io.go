package io

import "time"

type AuthRefreshInput struct {
	Token string `json:"token" validate:"required"`

	UserAgent string `json:"-"`
	ClientIP  string `json:"-"`
}

type AuthRefreshOutput struct {
	AccessToken           string    `json:"access_token"`
	AccessTokenExpiresAt  time.Time `json:"access_token_expires_at"`
	RefreshToken          string    `json:"refresh_token"`
	RefreshTokenExpiresIn int       `json:"refresh_token_expires_in"` // in seconds
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at"`
}
