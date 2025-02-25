package render

import (
	"context"

	"github.com/curiona-org/backend/internal/logger"
	"github.com/rs/zerolog"
)

type Renderer struct {
	logger *zerolog.Logger
}

func New(ctx context.Context) *Renderer {
	r := &Renderer{
		logger: logger.FromContext(ctx),
	}

	return r
}
