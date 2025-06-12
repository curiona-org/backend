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

	totalGenerated, err := app.repository.Roadmap.CountByAccountID(ctx, account.ID)
	if err != nil {
		return io.GetProfileOutput{}, err
	}

	totalInProgress, err := app.repository.Roadmap.CountAccountInProgressRoadmaps(ctx, account.ID)
	if err != nil {
		return io.GetProfileOutput{}, err
	}

	totalFinished, err := app.repository.Roadmap.CountAccountFinishedRoadmaps(ctx, account.ID)
	if err != nil {
		return io.GetProfileOutput{}, err
	}

	totalBookmarked, err := app.repository.Bookmark.Count(ctx, account.ID)
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
		Statistics: io.GetProfileOutputStatistics{
			TotalGeneratedRoadmaps:  totalGenerated,
			TotalInProgressRoadmaps: totalInProgress,
			TotalFinishedRoadmaps:   totalFinished,
			TotalBookmarkedRoadmaps: totalBookmarked,
		},
	}, nil
}
