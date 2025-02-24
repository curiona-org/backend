package app

import (
	"context"

	"github.com/curiona-org/backend/internal/app/io"
)

func (app *application) GetProfile(ctx context.Context, accountID int) (io.GetProfileOutput, error) {
	ctx, span := app.tracer.Start(ctx, "(*application.GetProfile)")
	defer span.End()

	account, err := app.repository.Account.GetByID(ctx, accountID)
	if err != nil {
		return io.GetProfileOutput{}, err
	}

	return io.GetProfileOutput{
		ID:       account.ID,
		Method:   account.Method,
		Email:    account.Email,
		Name:     account.Profile.Name,
		Avatar:   account.Profile.Avatar,
		JoinedAt: account.CreatedAt,
	}, nil
}
