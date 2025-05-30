package app

import (
	"context"
	"errors"

	"github.com/curiona-org/backend/internal/app/io"
	"github.com/curiona-org/backend/internal/domain"
	"github.com/curiona-org/backend/internal/filter"
	"github.com/curiona-org/backend/pkg/interval"
)

func (app *application) ListUserOnProgressRoadmaps(ctx context.Context, input io.ListUserOnProgressRoadmapsInput) (io.ListUserOnProgressRoadmapsOutput, error) {
	ctx, span := app.tracer.Start(ctx, "(*application.ListUserOnProgressRoadmaps)")
	defer span.End()

	count, err := app.repository.Roadmap.CountAccountOnProgressRoadmaps(ctx, input.AccountID)
	if err != nil {
		return io.ListUserOnProgressRoadmapsOutput{}, err
	}

	filters := filter.New(input, count)
	roadmaps, err := app.repository.Roadmap.ListAccountOnProgressRoadmaps(ctx, input.AccountID, filters)
	if err != nil && !errors.Is(err, domain.ErrRoadmapNotFound) {
		return io.ListUserOnProgressRoadmapsOutput{}, err
	}

	output := io.ListUserOnProgressRoadmapsOutput{
		Total:       filters.Paginator.Total,
		TotalPages:  filters.Paginator.TotalPages,
		CurrentPage: filters.Paginator.CurrentPage,
		Items:       make([]io.ListUserOnProgressRoadmapsOutputItem, len(roadmaps)),
	}

	for idx, roadmap := range roadmaps {
		output.Items[idx] = io.ListUserOnProgressRoadmapsOutputItem{
			ID:           roadmap.ID,
			Title:        roadmap.Title,
			Description:  roadmap.Description,
			Slug:         roadmap.Slug,
			TotalTopics:  roadmap.TotalTopics,
			CreatedAt:    roadmap.CreatedAt,
			UpdatedAt:    roadmap.UpdatedAt,
			IsBookmarked: roadmap.IsBookmarked,
			Progression: io.ListUserOnProgressRoadmapsOutputItemProgression{
				TotalTopics:          roadmap.Progression.TotalTopics,
				TotalFinishedTopics:  roadmap.Progression.TotalFinishedTopics,
				CompletionPercentage: roadmap.Progression.CompletionPercentage(),
				IsFinished:           roadmap.Progression.IsFinished,
				CreatedAt:            roadmap.Progression.CreatedAt,
				UpdatedAt:            roadmap.Progression.UpdatedAt,
			},
			PersonalizationOpts: io.ListUserOnProgressRoadmapsOutputItemPersonalizationOptions{
				DailyTimeAvailability: interval.FromDuration(roadmap.PersonalizationOptions.DailyTimeAvailability),
				TotalDuration:         interval.FromDuration(roadmap.PersonalizationOptions.TotalDuration),
				SkillLevel:            roadmap.PersonalizationOptions.SkillLevel.String(),
				AdditionalInfo:        roadmap.PersonalizationOptions.AdditionalInfo,
			},
		}
	}

	return output, nil
}
