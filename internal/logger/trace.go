package logger

import (
	"github.com/curiona-org/backend/pkg/tracing"
	"github.com/rs/zerolog"
)

type TraceHook struct{}

func (h TraceHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	ctx := e.GetCtx()
	spanId := tracing.GetTraceID(ctx)
	if spanId == "" {
		return
	}
	e.Str("span_id", spanId)
}
