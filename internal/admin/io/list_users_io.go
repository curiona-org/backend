package io

import (
	"time"

	"github.com/curiona-org/backend/internal/auth"
	"github.com/curiona-org/backend/internal/filter"
)

type ListUsersInput = filter.Params
type ListUsersOutput = filter.FilteredList[ListUsersOutputItem]

type ListUsersOutputItem struct {
	ID          int         `json:"id"`
	Method      auth.Method `json:"method"`
	Email       string      `json:"email"`
	Name        string      `json:"name"`
	Avatar      string      `json:"avatar"`
	IsSuspended bool        `json:"is_suspended"`
	IsAdmin     bool        `json:"is_admin"`
	JoinedAt    time.Time   `json:"joined_at"`
}
