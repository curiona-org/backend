package app

import (
	"context"
	"errors"

	"github.com/curiona-org/backend/internal/app/io"
	"github.com/curiona-org/backend/internal/domain"
	"github.com/curiona-org/backend/internal/filter"
	"github.com/curiona-org/backend/pkg/interval"
)

func (app *application) ListUserInProgressRoadmaps(ctx context.Context, input io.ListUserInProgressRoadmapsInput) (io.ListUserInProgressRoadmapsOutput, error) {
	ctx, span := app.tracer.Start(ctx, "(*application.ListUserInProgressRoadmaps)")
	defer span.End()

	count, err := app.repository.Roadmap.CountAccountInProgressRoadmaps(ctx, input.AccountID)
	if err != nil {
		return io.ListUserInProgressRoadmapsOutput{}, err
	}

	filters := filter.New(input, count)
	roadmaps, err := app.repository.Roadmap.ListAccountInProgressRoadmaps(ctx, input.AccountID, filters)
	if err != nil && !errors.Is(err, domain.ErrRoadmapNotFound) {
		return io.ListUserInProgressRoadmapsOutput{}, err
	}

	output := io.ListUserInProgressRoadmapsOutput{
		Total:       filters.Paginator.Total,
		TotalPages:  filters.Paginator.TotalPages,
		CurrentPage: filters.Paginator.CurrentPage,
		Items:       make([]io.ListUserInProgressRoadmapsOutputItem, len(roadmaps)),
	}

	for idx, roadmap := range roadmaps {
		output.Items[idx] = io.ListUserInProgressRoadmapsOutputItem{
			ID:           roadmap.ID,
			Title:        roadmap.Title,
			Description:  roadmap.Description,
			Slug:         roadmap.Slug,
			TotalTopics:  roadmap.TotalTopics,
			CreatedAt:    roadmap.CreatedAt,
			UpdatedAt:    roadmap.UpdatedAt,
			IsBookmarked: roadmap.IsBookmarked,
			Progression: io.ListUserInProgressRoadmapsOutputItemProgression{
				TotalTopics:          roadmap.Progression.TotalTopics,
				TotalFinishedTopics:  roadmap.Progression.TotalFinishedTopics,
				CompletionPercentage: roadmap.Progression.CompletionPercentage(),
				IsFinished:           roadmap.Progression.IsFinished,
				CreatedAt:            roadmap.Progression.CreatedAt,
				UpdatedAt:            roadmap.Progression.UpdatedAt,
			},
			PersonalizationOpts: io.ListUserInProgressRoadmapsOutputItemPersonalizationOptions{
				DailyTimeAvailability: interval.FromDuration(roadmap.PersonalizationOptions.DailyTimeAvailability),
				TotalDuration:         interval.FromDuration(roadmap.PersonalizationOptions.TotalDuration),
				SkillLevel:            roadmap.PersonalizationOptions.SkillLevel.String(),
				AdditionalInfo:        roadmap.PersonalizationOptions.AdditionalInfo,
			},
		}
	}

	return output, nil
}
