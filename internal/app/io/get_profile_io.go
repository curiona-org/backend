package io

import (
	"time"

	"github.com/curiona-org/backend/internal/auth"
)

type GetProfileOutput struct {
	ID       int         `json:"id"`
	Method   auth.Method `json:"method"`
	Email    string      `json:"email"`
	Name     string      `json:"name"`
	Avatar   string      `json:"avatar"`
	JoinedAt time.Time   `json:"joined_at"`
}
