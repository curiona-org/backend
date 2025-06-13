package app

import (
	"context"

	"github.com/curiona-org/backend/internal/app/io"
	"github.com/curiona-org/backend/internal/filter"
	"github.com/curiona-org/backend/pkg/interval"
)

func (app *application) ListCommunityRoadmaps(ctx context.Context, input io.ListCommunityRoadmapsInput) (io.ListCommunityRoadmapsOutput, error) {
	ctx, span := app.tracer.Start(ctx, "(*application.ListCommunityRoadmaps)")
	defer span.End()

	var count uint64
	var err error
	if input.Search != "" {
		count, err = app.repository.Roadmap.CountBySearching(ctx, input.AccountID, input.Search)
		if err != nil {
			return io.ListCommunityRoadmapsOutput{}, err
		}
	} else {
		count, err = app.repository.Roadmap.CountOmitAccountID(ctx, input.AccountID)
		if err != nil {
			return io.ListCommunityRoadmapsOutput{}, err
		}
	}

	filters := filter.New(input, count)
	roadmaps, err := app.repository.Roadmap.ListAll(ctx, filters)
	if err != nil {
		return io.ListCommunityRoadmapsOutput{}, err
	}

	output := io.ListCommunityRoadmapsOutput{
		Total:       filters.Paginator.Total,
		TotalPages:  filters.Paginator.TotalPages,
		CurrentPage: filters.Paginator.CurrentPage,
		Items:       make([]io.ListCommunityRoadmapsOutputItem, len(roadmaps)),
	}

	for idx, roadmap := range roadmaps {
		output.Items[idx] = io.ListCommunityRoadmapsOutputItem{
			ID:           roadmap.ID,
			Title:        roadmap.Title,
			Description:  roadmap.Description,
			Slug:         roadmap.Slug,
			TotalTopics:  roadmap.TotalTopics,
			CreatedAt:    roadmap.CreatedAt,
			UpdatedAt:    roadmap.UpdatedAt,
			IsBookmarked: roadmap.IsBookmarked,
			PersonalizationOpts: io.ListCommunityRoadmapsOutputItemPersonalizationOptions{
				DailyTimeAvailability: interval.FromDuration(roadmap.PersonalizationOptions.DailyTimeAvailability),
				TotalDuration:         interval.FromDuration(roadmap.PersonalizationOptions.TotalDuration),
				SkillLevel:            roadmap.PersonalizationOptions.SkillLevel.String(),
				AdditionalInfo:        roadmap.PersonalizationOptions.AdditionalInfo,
			},
			Creator: io.ListCommunityRoadmapsOutputItemCreator{
				ID:       roadmap.Account.ID,
				Name:     roadmap.Account.Profile.Name,
				Avatar:   roadmap.Account.Profile.Avatar,
				JoinedAt: roadmap.Account.CreatedAt,
			},
		}

		if input.AccountID != 0 && roadmap.Progression != nil {
			output.Items[idx].Progression = io.ListCommunityRoadmapsOutputItemProgression{
				TotalTopics:          roadmap.Progression.TotalTopics,
				TotalFinishedTopics:  roadmap.Progression.TotalFinishedTopics,
				CompletionPercentage: roadmap.Progression.CompletionPercentage(),
				IsFinished:           roadmap.Progression.IsFinished,
				FinishedAt:           roadmap.Progression.FinishedAt,
				CreatedAt:            roadmap.Progression.CreatedAt,
				UpdatedAt:            roadmap.Progression.UpdatedAt,
			}
		}
	}

	return output, nil
}
