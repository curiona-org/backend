package logger

import (
	"os"

	"github.com/curiona-org/backend/internal/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
)

func Init() {
	// Setup logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	switch {
	case config.IsDevelopment():
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Logger = log.
			Output(zerolog.ConsoleWriter{Out: os.Stderr}).
			With().
			Caller().
			Stack().
			Logger()
	case config.IsStaging():
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Logger = log.
			With().
			Caller().
			Stack().
			Logger()
	default:
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
		log.Logger = log.
			With().
			Caller().
			Stack().
			Logger().
			Sample(&zerolog.BasicSampler{
				N: 2,
			})
	}
}
