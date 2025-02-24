package io

import (
	"time"

	"github.com/curiona-org/backend/internal/domain/object"
)

type GetProfileOutput struct {
	ID       int                    `json:"id"`
	Provider object.AccountProvider `json:"provider"`
	Email    string                 `json:"email"`
	Name     string                 `json:"name"`
	Avatar   string                 `json:"avatar"`
	JoinedAt time.Time              `json:"joined_at"`
}
