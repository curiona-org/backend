package logger

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
)

func Init(debug bool) {
	// Setup logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Logger = log.
			Output(zerolog.ConsoleWriter{Out: os.Stderr}).
			With().
			Caller().
			Stack().
			Logger()
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
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
