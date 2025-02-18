package book

import (
	"context"
	"errors"
)

type Volume struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Pages       int    `json:"pages"`
}

type Client interface {
	Search(ctx context.Context, query string) ([]*Volume, error)
}

type Source string

const (
	GoogleBooks Source = "google"
)

func NewAPI(source Source) (Client, error) {
	switch source {
	case GoogleBooks:
		return NewGoogleBooksClient(), nil
	default:
		return nil, errors.New("invalid book api source")
	}
}
