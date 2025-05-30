package app

import (
	"context"
	"errors"

	"github.com/curiona-org/backend/internal/app/io"
	"github.com/curiona-org/backend/internal/domain"
	"github.com/curiona-org/backend/internal/filter"
	"github.com/curiona-org/backend/pkg/interval"
)

func (app *application) ListUserFinishedRoadmaps(ctx context.Context, input io.ListUserFinishedRoadmapsInput) (io.ListUserFinishedRoadmapsOutput, error) {
	ctx, span := app.tracer.Start(ctx, "(*application.ListUserFinishedRoadmaps)")
	defer span.End()

	count, err := app.repository.Roadmap.CountAccountFinishedRoadmaps(ctx, input.AccountID)
	if err != nil {
		return io.ListUserFinishedRoadmapsOutput{}, err
	}

	filters := filter.New(input, count)
	roadmaps, err := app.repository.Roadmap.ListAccountFinishedRoadmaps(ctx, input.AccountID, filters)
	if err != nil && !errors.Is(err, domain.ErrRoadmapNotFound) {
		return io.ListUserFinishedRoadmapsOutput{}, err
	}

	output := io.ListUserFinishedRoadmapsOutput{
		Total:       filters.Paginator.Total,
		TotalPages:  filters.Paginator.TotalPages,
		CurrentPage: filters.Paginator.CurrentPage,
		Items:       make([]io.ListUserFinishedRoadmapsOutputItem, len(roadmaps)),
	}

	for idx, roadmap := range roadmaps {
		output.Items[idx] = io.ListUserFinishedRoadmapsOutputItem{
			ID:           roadmap.ID,
			Title:        roadmap.Title,
			Description:  roadmap.Description,
			Slug:         roadmap.Slug,
			TotalTopics:  roadmap.TotalTopics,
			CreatedAt:    roadmap.CreatedAt,
			UpdatedAt:    roadmap.UpdatedAt,
			IsBookmarked: roadmap.IsBookmarked,
			Progression: io.ListUserFinishedRoadmapsOutputItemProgression{
				TotalTopics:          roadmap.Progression.TotalTopics,
				TotalFinishedTopics:  roadmap.Progression.TotalFinishedTopics,
				CompletionPercentage: roadmap.Progression.CompletionPercentage(),
				IsFinished:           roadmap.Progression.IsFinished,
				CreatedAt:            roadmap.Progression.CreatedAt,
				UpdatedAt:            roadmap.Progression.UpdatedAt,
			},
			PersonalizationOpts: io.ListUserFinishedRoadmapsOutputItemPersonalizationOptions{
				DailyTimeAvailability: interval.FromDuration(roadmap.PersonalizationOptions.DailyTimeAvailability),
				TotalDuration:         interval.FromDuration(roadmap.PersonalizationOptions.TotalDuration),
				SkillLevel:            roadmap.PersonalizationOptions.SkillLevel.String(),
				AdditionalInfo:        roadmap.PersonalizationOptions.AdditionalInfo,
			},
		}
	}

	return output, nil
}
