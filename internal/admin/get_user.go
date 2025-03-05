package admin

import (
	"context"

	"github.com/curiona-org/backend/internal/admin/io"
	"github.com/curiona-org/backend/pkg/interval"
)

func (app *adminApplication) GetUser(ctx context.Context, accountID int) (io.GetUserOutput, error) {
	ctx, span := app.tracer.Start(ctx, "(*adminApplication.GetUser)")
	defer span.End()

	account, err := app.repository.Account.GetByID(ctx, accountID)
	if err != nil {
		return io.GetUserOutput{}, err
	}

	roadmaps, err := app.repository.Roadmap.ListByAccountID(ctx, accountID)
	if err != nil {
		return io.GetUserOutput{}, err
	}

	var roadmapsOutput []io.GetUserOutputRoadmap
	for _, roadmap := range roadmaps {
		roadmapOutput := io.GetUserOutputRoadmap{
			ID:                   roadmap.ID,
			Title:                roadmap.Title,
			Description:          roadmap.Description,
			Slug:                 roadmap.Slug,
			TotalTopics:          roadmap.TotalTopics,
			TotalFinishedTopics:  roadmap.TotalFinishedTopics,
			CompletionPercentage: roadmap.CompletionPercentage(),
			CreatedAt:            roadmap.CreatedAt,
			UpdatedAt:            roadmap.UpdatedAt,
			PersonalizationOpts: io.GetUserOutputRoadmapPersonalizationOptions{
				DailyTimeAvailability: interval.FromDuration(roadmap.PersonalizationOptions.DailyTimeAvailability),
				TotalDuration:         interval.FromDuration(roadmap.PersonalizationOptions.TotalDuration),
				SkillLevel:            roadmap.PersonalizationOptions.SkillLevel.String(),
				AdditionalInfo:        roadmap.PersonalizationOptions.AdditionalInfo,
			},
		}

		roadmapsOutput = append(roadmapsOutput, roadmapOutput)
	}

	output := io.GetUserOutput{
		ID:          account.ID,
		Method:      account.Method,
		Email:       account.Email,
		Name:        account.Profile.Name,
		Avatar:      account.Profile.Avatar,
		IsSuspended: account.IsSuspended,
		IsAdmin:     account.IsAdmin,
		JoinedAt:    account.CreatedAt,
		Roadmaps:    roadmapsOutput,
	}

	return output, nil
}
