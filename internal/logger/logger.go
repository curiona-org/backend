package logger

import (
	"os"

	"github.com/roadmap-thesis/backend/internal/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
)

func Init() {
	// Setup logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	if config.IsDevelopment() {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Logger = log.
			Output(zerolog.ConsoleWriter{Out: os.Stderr}).
			With().
			Caller().
			Stack().
			Logger()
	} else if config.IsStaging() {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Logger = log.
			With().
			Caller().
			Stack().
			Logger()
	} else {
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
