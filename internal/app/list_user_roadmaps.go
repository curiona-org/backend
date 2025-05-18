package app

import (
	"context"
	"errors"

	"github.com/curiona-org/backend/internal/app/io"
	"github.com/curiona-org/backend/internal/domain"
	"github.com/curiona-org/backend/pkg/filter"
	"github.com/curiona-org/backend/pkg/interval"
)

func (app *application) ListUserRoadmaps(ctx context.Context, input io.ListUserRoadmapsInput) (io.ListUserRoadmapsOutput, error) {
	ctx, span := app.tracer.Start(ctx, "(*application.ListUserRoadmaps)")
	defer span.End()

	count, err := app.repository.Roadmap.CountAccountRoadmaps(ctx, input.AccountID)
	if err != nil {
		return io.ListUserRoadmapsOutput{}, err
	}

	filters := filter.New(input, count)
	roadmaps, err := app.repository.Roadmap.ListByAccountID(ctx, filters)
	if err != nil && !errors.Is(err, domain.ErrRoadmapNotFound) {
		return io.ListUserRoadmapsOutput{}, err
	}

	output := io.ListUserRoadmapsOutput{
		Total:       filters.Paginator.Total,
		TotalPages:  filters.Paginator.TotalPages,
		CurrentPage: filters.Paginator.CurrentPage,
		Items:       make([]io.ListUserRoadmapsOutputItem, len(roadmaps)),
	}

	for idx, roadmap := range roadmaps {
		output.Items[idx] = io.ListUserRoadmapsOutputItem{
			ID:                   roadmap.ID,
			Title:                roadmap.Title,
			Description:          roadmap.Description,
			Slug:                 roadmap.Slug,
			TotalTopics:          roadmap.TotalTopics,
			TotalFinishedTopics:  roadmap.Progression.TotalFinishedTopics,
			CompletionPercentage: roadmap.CompletionPercentage(),
			CreatedAt:            roadmap.CreatedAt,
			UpdatedAt:            roadmap.UpdatedAt,
			PersonalizationOpts: io.ListUserRoadmapsOutputPersonalizationOptions{
				DailyTimeAvailability: interval.FromDuration(roadmap.PersonalizationOptions.DailyTimeAvailability),
				TotalDuration:         interval.FromDuration(roadmap.PersonalizationOptions.TotalDuration),
				SkillLevel:            roadmap.PersonalizationOptions.SkillLevel.String(),
				AdditionalInfo:        roadmap.PersonalizationOptions.AdditionalInfo,
			},
		}
	}

	return output, nil
}
