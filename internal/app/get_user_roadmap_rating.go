package app

import (
	"context"
	"errors"

	"github.com/curiona-org/backend/internal/app/io"
	"github.com/curiona-org/backend/internal/domain"
)

func (app *application) GetUserRoadmapRating(ctx context.Context, input io.GetUserRoadmapRatingInput) (io.GetUserRoadmapRatingOutput, error) {
	ctx, span := app.tracer.Start(ctx, "(*application.GetUserRoadmapRating)")
	defer span.End()

	rating, err := app.repository.Rating.GetRoadmapRatingByAccountID(ctx, input.AccountID, input.Slug)
	if err != nil {
		if errors.Is(err, domain.ErrRatingNotFound) {
			return io.GetUserRoadmapRatingOutput{}, nil
		}
		return io.GetUserRoadmapRatingOutput{}, err
	}

	return io.GetUserRoadmapRatingOutput{
		IsRated:                        true,
		RoadmapID:                      rating.RoadmapID,
		ProgressionTotalTopics:         rating.ProgressionTotalTopics,
		ProgressionTotalFinishedTopics: rating.ProgressionTotalFinishedTopics,
		Rating:                         rating.Rating,
		Comment:                        rating.Comment,
		CreatedAt:                      rating.CreatedAt,
		UpdatedAt:                      rating.UpdatedAt,
	}, nil
}
