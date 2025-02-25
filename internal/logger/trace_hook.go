package logger

import (
	"github.com/curiona-org/backend/pkg/tracing"
	"github.com/rs/zerolog"
)

type TraceHook struct{}

func (h TraceHook) Run(e *zerolog.Event, _ zerolog.Level, _ string) {
	ctx := e.GetCtx()
	spanID := tracing.GetTraceID(ctx)
	if spanID == "" {
		return
	}
	e.Str("span_id", spanID)
}
