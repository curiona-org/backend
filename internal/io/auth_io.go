package io

import "time"

type AuthInput struct {
	Name       string `json:"name" validate:"omitempty"`
	Avatar     string `json:"avatar" validate:"omitempty,url"`
	Email      string `json:"email" validate:"required_without=OAuthToken,omitempty,email"`
	Password   string `json:"password" validate:"required_without=OAuthToken,omitempty,min=6"`
	OAuthToken string `json:"oauth_token" validate:"required_without_all=Email Password,omitempty"`

	IgnorePasswordCheck bool `json:"-"` // used to skip password check in some cases
}

type AuthOutput struct {
	Created bool              `json:"created"`
	Token   string            `json:"token"`
	Account AuthOutputAccount `json:"account"`
}

type AuthOutputAccount struct {
	ID       int       `json:"id"`
	Email    string    `json:"email"`
	Name     string    `json:"name"`
	Avatar   string    `json:"avatar"`
	JoinedAt time.Time `json:"joined_at"`
}
