package admin

import (
	"context"
	"errors"

	"github.com/curiona-org/backend/internal/admin/io"
	"github.com/curiona-org/backend/internal/domain"
	"github.com/curiona-org/backend/internal/filter"
)

func (app *adminApplication) ListRoadmapRatings(ctx context.Context, input io.ListRoadmapRatingsInput) (io.ListRoadmapRatingsOutput, error) {
	ctx, span := app.tracer.Start(ctx, "(*application.ListRoadmapRatings)")
	defer span.End()

	var count uint64
	var err error
	if input.Search != "" {
		count, err = app.repository.Rating.CountRoadmapRatingsBySearching(ctx, input.ID, input.Search)
		if err != nil {
			return io.ListRoadmapRatingsOutput{}, err
		}
	} else {
		count, err = app.repository.Rating.CountRoadmapRatings(ctx, input.ID)
		if err != nil {
			return io.ListRoadmapRatingsOutput{}, err
		}
	}

	filters := filter.New(input, count)
	ratings, err := app.repository.Rating.GetRoadmapRatings(ctx, filters)
	if err != nil {
		if errors.Is(err, domain.ErrRatingNotFound) {
			return io.ListRoadmapRatingsOutput{}, nil
		}
		return io.ListRoadmapRatingsOutput{}, err
	}

	output := io.ListRoadmapRatingsOutput{
		Total:       filters.Paginator.Total,
		TotalPages:  filters.Paginator.TotalPages,
		CurrentPage: filters.Paginator.CurrentPage,
		Items:       make([]io.ListRoadmapRatingsOutputItem, len(ratings)),
	}

	for i, rating := range ratings {
		output.Items[i] = io.ListRoadmapRatingsOutputItem{
			IsRated:                        !rating.IsZero(),
			RoadmapID:                      rating.RoadmapID,
			AccountID:                      rating.Account.ID,
			ProgressionTotalTopics:         rating.ProgressionTotalTopics,
			ProgressionTotalFinishedTopics: rating.ProgressionTotalFinishedTopics,
			Rating:                         rating.Rating,
			Comment:                        rating.Comment,
			User: io.ListRoadmapRatingsOutputItemUser{
				ID:          rating.Account.ID,
				Method:      rating.Account.Method,
				Email:       rating.Account.Email,
				Name:        rating.Account.Profile.Name,
				Avatar:      rating.Account.Profile.Avatar,
				IsSuspended: rating.Account.IsSuspended,
				JoinedAt:    rating.Account.CreatedAt,
			},
			CreatedAt: rating.CreatedAt,
			UpdatedAt: rating.UpdatedAt,
		}
	}

	return output, nil
}
