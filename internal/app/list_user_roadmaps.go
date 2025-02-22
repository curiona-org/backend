package app

import (
	"context"
	"errors"

	"github.com/roadmap-thesis/backend/internal/app/io"
	"github.com/roadmap-thesis/backend/internal/domain"
	"github.com/roadmap-thesis/backend/internal/domain/object"
)

func (app *application) ListUserRoadmaps(ctx context.Context, accountID int) (io.ListUserRoadmapsOutput, error) {
	ctx, span := app.tracer.Start(ctx, "(*application.ListUserRoadmaps)")
	defer span.End()

	roadmaps, err := app.repository.Roadmap.ListByAccountID(ctx, accountID)
	if err != nil && !errors.Is(err, domain.ErrRoadmapNotFound) {
		return io.ListUserRoadmapsOutput{}, err
	}

	output := io.ListUserRoadmapsOutput{
		TotalRoadmaps: len(roadmaps),
		Roadmaps:      []io.ListUserRoadmapsOutputRoadmap{},
	}

	if len(roadmaps) == 0 {
		return output, nil
	}

	for _, roadmap := range roadmaps {
		outputRoadmap := io.ListUserRoadmapsOutputRoadmap{
			ID:                   roadmap.ID,
			Title:                roadmap.Title,
			Description:          roadmap.Description,
			Slug:                 roadmap.Slug,
			TotalTopics:          roadmap.TotalTopics(),
			CompletionPercentage: roadmap.CompletionPercentage(),
			PersonalizationOpts: io.ListUserRoadmapsOutputPersonalizationOptions{
				DailyTimeAvailability: object.NewIntervalFromDuration(roadmap.PersonalizationOptions.DailyTimeAvailability),
				TotalDuration:         object.NewIntervalFromDuration(roadmap.PersonalizationOptions.TotalDuration),
				SkillLevel:            roadmap.PersonalizationOptions.SkillLevel.String(),
				AdditionalInfo:        roadmap.PersonalizationOptions.AdditionalInfo,
			},
			CreatedAt: roadmap.CreatedAt,
			UpdatedAt: roadmap.UpdatedAt,
		}

		output.Roadmaps = append(output.Roadmaps, outputRoadmap)
	}

	return output, nil
}
