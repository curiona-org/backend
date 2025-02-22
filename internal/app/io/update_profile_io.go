package io

import "time"

type UpdateProfileInput struct {
	AccountID int    `json:"-"`
	Name      string `json:"name" validate:"required"`
}

type UpdateProfileOutput struct {
	Name      string    `json:"name"`
	Avatar    string    `json:"avatar"`
	UpdatedAt time.Time `json:"updated_at"`
}
