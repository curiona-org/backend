package app

import (
	"context"

	"github.com/roadmap-thesis/backend/internal/app/io"
	"github.com/roadmap-thesis/backend/internal/domain"
)

func (app *application) UpdateProfile(ctx context.Context, input io.UpdateProfileInput) (io.UpdateProfileOutput, error) {
	ctx, span := app.tracer.Start(ctx, "(*application.GetProfile)")
	defer span.End()

	var profileOutput *domain.Profile
	err := app.repository.Profile.Update(ctx, input.AccountID, func(profile *domain.Profile) (bool, error) {
		profile.Update(input.Name)

		profileOutput = profile

		return true, nil
	})
	if err != nil {
		return io.UpdateProfileOutput{}, err
	}

	output := io.UpdateProfileOutput{
		Name:      profileOutput.Name,
		Avatar:    profileOutput.Avatar,
		UpdatedAt: profileOutput.UpdatedAt,
	}

	return output, nil
}
