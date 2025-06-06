package app

import (
	"context"

	"github.com/curiona-org/backend/internal/app/io"
	"github.com/curiona-org/backend/internal/cerrors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (app *application) PromptModeration(ctx context.Context, input io.PromptModerationInput) (io.PromptModerationOutput, error) {
	ctx, span := app.tracer.Start(ctx, "(*application.PromptModeration)", trace.WithAttributes(
		attribute.String("input.prompt", input.Prompt),
	))
	defer span.End()

	flagged, err := app.llm.Moderate(ctx, input.Prompt)
	if err != nil {
		return io.PromptModerationOutput{}, err
	}

	if flagged {
		span.SetAttributes(attribute.Bool("output.flagged", true))
		return io.PromptModerationOutput{
			Flagged: true,
			Reason:  cerrors.ErrLLMFlaggedContentDetected.Message(),
		}, nil
	}

	span.SetAttributes(attribute.Bool("output.flagged", false))
	return io.PromptModerationOutput{
		Flagged: false,
		Reason:  "",
	}, nil
}
