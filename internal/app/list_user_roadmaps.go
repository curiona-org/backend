package app

import (
	"context"
	"errors"

	"github.com/curiona-org/backend/internal/app/io"
	"github.com/curiona-org/backend/internal/domain"
	"github.com/curiona-org/backend/pkg/interval"
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
			TotalTopics:          roadmap.TotalTopics,
			TotalFinishedTopics:  roadmap.TotalFinishedTopics,
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

		output.Roadmaps = append(output.Roadmaps, outputRoadmap)
	}

	return output, nil
}
