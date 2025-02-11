package auth

import "time"

type Token interface {
	Generate(id int) (string, error)
	Parse(token string) (*Payload, error)
	ExpiresAt() time.Time
	ExpiresIn() time.Duration
}
