package admin

import (
	"context"

	"github.com/curiona-org/backend/internal/admin/io"
	"github.com/curiona-org/backend/internal/domain/object"
	"github.com/curiona-org/backend/pkg/pagination"
)

func (app *adminApplication) ListRoadmaps(ctx context.Context, input io.ListRoadmapsInput) (io.ListRoadmapsOutput, error) {
	ctx, span := app.tracer.Start(ctx, "(*adminApplication.ListUsers)")
	defer span.End()

	count, err := app.repository.Roadmap.Count(ctx)
	if err != nil {
		return io.ListRoadmapsOutput{}, err
	}

	pagination := pagination.NewOffsetPaginator(input.Page, input.Limit, count)
	roadmaps, err := app.repository.Roadmap.ListAll(ctx, pagination)
	if err != nil {
		return io.ListRoadmapsOutput{}, err
	}

	output := io.ListRoadmapsOutput{
		Total:       pagination.Total,
		TotalPages:  pagination.TotalPages,
		CurrentPage: pagination.CurrentPage,
		Items:       make([]io.ListRoadmapsOutputItem, len(roadmaps)),
	}

	for idx, roadmap := range roadmaps {
		output.Items[idx] = io.ListRoadmapsOutputItem{
			ID:                   roadmap.ID,
			Title:                roadmap.Title,
			Description:          roadmap.Description,
			Slug:                 roadmap.Slug,
			TotalTopics:          roadmap.TotalTopics,
			CompletionPercentage: roadmap.CompletionPercentage(),
			PersonalizationOpts: io.ListRoadmapsOutputItemPersonalizationOptions{
				DailyTimeAvailability: object.NewIntervalFromDuration(roadmap.PersonalizationOptions.DailyTimeAvailability),
				TotalDuration:         object.NewIntervalFromDuration(roadmap.PersonalizationOptions.TotalDuration),
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
			CreatedAt: roadmap.CreatedAt,
			UpdatedAt: roadmap.UpdatedAt,
		}
	}

	return output, nil
}
