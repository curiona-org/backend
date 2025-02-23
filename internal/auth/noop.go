package auth

import "time"

type Noop struct{}

var _ Token

func NewNoop() Token {
	return Noop{}
}

func (n Noop) Generate(id int) (string, error) {
	_ = id
	return "", nil
}

func (n Noop) Parse(token string) (*Payload, error) {
	_ = token
	return nil, nil
}

func (n Noop) ExpiresAt() time.Time {
	return time.Time{}
}

func (n Noop) ExpiresIn() time.Duration {
	return 0
}
