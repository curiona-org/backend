package app

import (
	"context"

	"github.com/curiona-org/backend/internal/app/io"
	"github.com/curiona-org/backend/internal/filter"
	"github.com/curiona-org/backend/pkg/interval"
)

func (app *application) ListBookmarkedRoadmaps(ctx context.Context, input io.ListBookmarkedRoadmapsInput) (io.ListBookmarkedRoadmapsOutput, error) {
	ctx, span := app.tracer.Start(ctx, "(*application.ListBookmarkedRoadmaps)")
	defer span.End()

	var count uint64
	var err error
	if input.Search != "" {
		count, err = app.repository.Bookmark.CountBySearching(ctx, input.AccountID, input.Search)
		if err != nil {
			return io.ListBookmarkedRoadmapsOutput{}, err
		}
	} else {
		count, err = app.repository.Bookmark.Count(ctx, input.AccountID)
		if err != nil {
			return io.ListBookmarkedRoadmapsOutput{}, err
		}
	}

	filters := filter.New(input, count)
	roadmaps, err := app.repository.Bookmark.ListBookmarkedRoadmaps(ctx, filters)
	if err != nil {
		return io.ListBookmarkedRoadmapsOutput{}, err
	}

	output := io.ListBookmarkedRoadmapsOutput{
		Total:       filters.Paginator.Total,
		TotalPages:  filters.Paginator.TotalPages,
		CurrentPage: filters.Paginator.CurrentPage,
		Items:       make([]io.ListBookmarkedRoadmapsOutputItem, len(roadmaps)),
	}

	for idx, roadmap := range roadmaps {
		output.Items[idx] = io.ListBookmarkedRoadmapsOutputItem{
			ID:           roadmap.ID,
			Title:        roadmap.Title,
			Description:  roadmap.Description,
			Slug:         roadmap.Slug,
			TotalTopics:  roadmap.TotalTopics,
			CreatedAt:    roadmap.CreatedAt,
			UpdatedAt:    roadmap.UpdatedAt,
			IsBookmarked: roadmap.IsBookmarked,
			Progression: io.ListBookmarkedRoadmapsOutputItemProgression{
				TotalTopics:          roadmap.Progression.TotalTopics,
				TotalFinishedTopics:  roadmap.Progression.TotalFinishedTopics,
				CompletionPercentage: roadmap.Progression.CompletionPercentage(),
				IsFinished:           roadmap.Progression.IsFinished,
				FinishedAt:           roadmap.Progression.FinishedAt,
				CreatedAt:            roadmap.Progression.CreatedAt,
				UpdatedAt:            roadmap.Progression.UpdatedAt,
			},
			PersonalizationOpts: io.ListBookmarkedRoadmapsOutputItemPersonalizationOptions{
				DailyTimeAvailability: interval.FromDuration(roadmap.PersonalizationOptions.DailyTimeAvailability),
				TotalDuration:         interval.FromDuration(roadmap.PersonalizationOptions.TotalDuration),
				SkillLevel:            roadmap.PersonalizationOptions.SkillLevel.String(),
				AdditionalInfo:        roadmap.PersonalizationOptions.AdditionalInfo,
			},
			Creator: io.ListBookmarkedRoadmapsOutputItemUser{
				ID:          roadmap.Account.ID,
				Method:      roadmap.Account.Method,
				Email:       roadmap.Account.Email,
				Name:        roadmap.Account.Profile.Name,
				Avatar:      roadmap.Account.Profile.Avatar,
				IsSuspended: roadmap.Account.IsSuspended,
				JoinedAt:    roadmap.Account.CreatedAt,
			},
		}
	}

	return output, nil
}
