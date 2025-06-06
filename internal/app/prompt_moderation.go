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

	// Validate if user already hit the limit of generating roadmaps or suspended
	account, err := app.repository.Account.GetByID(ctx, input.AccountID)
	if err != nil {
		return io.PromptModerationOutput{}, cerrors.ErrUnauthorized
	}

	if account.IsSuspended {
		return io.PromptModerationOutput{}, cerrors.ErrForbidden
	}

	if !account.IsAdmin {
		// Check if the account has reached the maximum number of generated roadmaps by
		// checking the number of unfinished roadmaps.
		accountRoadmapsCount, err := app.repository.Roadmap.CountUnfinishedRoadmapsByAccountID(ctx, input.AccountID)
		if err != nil {
			return io.PromptModerationOutput{}, err
		}

		if accountRoadmapsCount >= uint64(account.Profile.MaxGeneratedRoadmaps) {
			return io.PromptModerationOutput{}, cerrors.ErrLLMMaximumRoadmapGenerationReached
		}
	}

	flagged, err := app.llm.Moderate(ctx, input.Prompt)
	if err != nil {
		return io.PromptModerationOutput{}, cerrors.ErrLLMProviderUnavailable
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
