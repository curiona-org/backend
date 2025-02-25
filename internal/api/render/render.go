package render

import (
	"bytes"
	"context"
	"sync"

	"github.com/curiona-org/backend/internal/logger"
	"github.com/rs/zerolog"
)

type Renderer struct {
	logger *zerolog.Logger
	pool   *sync.Pool
}

func New(ctx context.Context) *Renderer {
	r := &Renderer{
		logger: logger.FromContext(ctx),
		pool: &sync.Pool{
			New: func() any {
				return bytes.NewBuffer(make([]byte, 0, 1024))
			},
		},
	}

	return r
}
