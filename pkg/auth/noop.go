package auth

import "time"

type Noop struct{}

func NewNoop() Token {
	return Noop{}
}

func (n Noop) Generate(id int) (string, error) {
	return "", nil
}

func (n Noop) Parse(token string) (*Payload, error) {
	return nil, nil
}

func (n Noop) ExpiresAt() time.Time {
	return time.Time{}
}

func (n Noop) ExpiresIn() time.Duration {
	return 0
}
