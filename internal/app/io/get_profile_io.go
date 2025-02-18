package io

import "time"

type GetProfileOutput struct {
	ID       int       `json:"id"`
	Email    string    `json:"email"`
	Name     string    `json:"name"`
	Avatar   string    `json:"avatar"`
	JoinedAt time.Time `json:"joined_at"`
}
