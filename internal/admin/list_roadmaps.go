package admin

import (
	"context"

	"github.com/curiona-org/backend/internal/admin/io"
	"github.com/curiona-org/backend/internal/filter"
	"github.com/curiona-org/backend/pkg/interval"
)

func (app *adminApplication) ListRoadmaps(ctx context.Context, input io.ListRoadmapsInput) (io.ListRoadmapsOutput, error) {
	ctx, span := app.tracer.Start(ctx, "(*adminApplication.ListUsers)")
	defer span.End()

	totalItems, err := app.repository.Roadmap.Count(ctx)
	if err != nil {
		return io.ListRoadmapsOutput{}, err
	}

	filters := filter.New(input, totalItems)
	roadmaps, err := app.repository.Roadmap.ListAll(ctx, filters)
	if err != nil {
		return io.ListRoadmapsOutput{}, err
	}

	output := io.ListRoadmapsOutput{
		Total:       filters.Paginator.Total,
		TotalPages:  filters.Paginator.TotalPages,
		CurrentPage: filters.Paginator.CurrentPage,
		Items:       make([]io.ListRoadmapsOutputItem, len(roadmaps)),
	}

	for idx, roadmap := range roadmaps {
		output.Items[idx] = io.ListRoadmapsOutputItem{
			ID:                   roadmap.ID,
			Title:                roadmap.Title,
			Description:          roadmap.Description,
			Slug:                 roadmap.Slug,
			TotalTopics:          roadmap.TotalTopics,
			TotalFinishedTopics:  roadmap.Progression.TotalFinishedTopics,
			CompletionPercentage: roadmap.CompletionPercentage(),
			CreatedAt:            roadmap.CreatedAt,
			UpdatedAt:            roadmap.UpdatedAt,
			PersonalizationOpts: io.ListRoadmapsOutputItemPersonalizationOptions{
				DailyTimeAvailability: interval.FromDuration(roadmap.PersonalizationOptions.DailyTimeAvailability),
				TotalDuration:         interval.FromDuration(roadmap.PersonalizationOptions.TotalDuration),
				SkillLevel:            roadmap.PersonalizationOptions.SkillLevel.String(),
				AdditionalInfo:        roadmap.PersonalizationOptions.AdditionalInfo,
			},
			Creator: io.ListRoadmapsOutputItemUser{
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
