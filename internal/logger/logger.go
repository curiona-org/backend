package logger

import (
	"context"
	"os"
	"sync"

	"github.com/curiona-org/backend/internal/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
)

var (
	once   sync.Once
	logger zerolog.Logger
)

func Get() zerolog.Logger {
	once.Do(func() {
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
		zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

		switch {
		case config.IsDevelopment():
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
			logger = log.
				Output(zerolog.ConsoleWriter{Out: os.Stderr}).
				With().
				Caller().
				Stack().
				Logger()
		case config.IsStaging():
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
			logger = log.
				With().
				Caller().
				Stack().
				Logger()
		default:
			zerolog.SetGlobalLevel(zerolog.InfoLevel)
			logger = log.
				With().
				Caller().
				Stack().
				Logger().
				Sample(&zerolog.BasicSampler{
					N: 10,
				})
		}
	})

	return logger
}

func FromContext(ctx context.Context) *zerolog.Logger {
	return zerolog.Ctx(ctx)
}
