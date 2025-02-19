package app

import (
	"context"

	"github.com/roadmap-thesis/backend/internal/app/io"
	"github.com/roadmap-thesis/backend/pkg/auth"
	"go.opentelemetry.io/otel/attribute"
)

func (app *application) GetProfile(ctx context.Context) (io.GetProfileOutput, error) {
	ctx, span := app.tracer.Start(ctx, "(*application.GetProfile)")
	defer span.End()

	auth := auth.FromContext(ctx)

	span.SetAttributes(attribute.Int("account_id", auth.ID))

	account, err := app.repository.Account.GetByID(ctx, auth.ID)
	if err != nil {
		return io.GetProfileOutput{}, err
	}

	return io.GetProfileOutput{
		ID:       account.ID,
		Email:    account.Email,
		Name:     account.Profile.Name,
		Avatar:   account.Profile.Avatar,
		JoinedAt: account.CreatedAt,
	}, nil
}
