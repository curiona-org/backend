package admin

import (
	"context"
	"errors"

	"github.com/curiona-org/backend/internal/admin/io"
	"github.com/curiona-org/backend/internal/cerrors"
	"github.com/curiona-org/backend/internal/domain"
	"github.com/curiona-org/backend/internal/filter"
	"github.com/curiona-org/backend/pkg/interval"
)

func (app *adminApplication) GetUser(ctx context.Context, input io.GetUserInput) (io.GetUserOutput, error) {
	ctx, span := app.tracer.Start(ctx, "(*adminApplication.GetUser)")
	defer span.End()

	account, err := app.repository.Account.GetByID(ctx, input.AccountID)
	if err != nil {
		if errors.Is(err, domain.ErrAccountNotFound) {
			return io.GetUserOutput{}, cerrors.ErrNotFound.Msg("account")
		}
		return io.GetUserOutput{}, err
	}

	count, err := app.repository.Roadmap.CountByAccountID(ctx, input.AccountID)
	if err != nil {
		return io.GetUserOutput{}, err
	}

	filters := filter.New(input, count)
	roadmaps, err := app.repository.Roadmap.ListByAccountID(ctx, filters)
	if err != nil && !errors.Is(err, domain.ErrRoadmapNotFound) {
		return io.GetUserOutput{}, err
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
	}

	output.Roadmaps = io.ListUserRoadmapsOutput{
		Total:       filters.Paginator.Total,
		TotalPages:  filters.Paginator.TotalPages,
		CurrentPage: filters.Paginator.CurrentPage,
		Items:       make([]io.ListUserRoadmapsOutputItem, len(roadmaps)),
	}

	for idx, roadmap := range roadmaps {
		output.Roadmaps.Items[idx] = io.ListUserRoadmapsOutputItem{
			ID:           roadmap.ID,
			Title:        roadmap.Title,
			Description:  roadmap.Description,
			Slug:         roadmap.Slug,
			TotalTopics:  roadmap.TotalTopics,
			CreatedAt:    roadmap.CreatedAt,
			UpdatedAt:    roadmap.UpdatedAt,
			IsBookmarked: roadmap.IsBookmarked,
			Progression: io.ListUserRoadmapsOutputItemProgression{
				TotalTopics:          roadmap.Progression.TotalTopics,
				TotalFinishedTopics:  roadmap.Progression.TotalFinishedTopics,
				CompletionPercentage: roadmap.Progression.CompletionPercentage(),
				CreatedAt:            roadmap.Progression.CreatedAt,
				UpdatedAt:            roadmap.Progression.UpdatedAt,
			},
			PersonalizationOpts: io.ListUserRoadmapsOutputItemPersonalizationOptions{
				DailyTimeAvailability: interval.FromDuration(roadmap.PersonalizationOptions.DailyTimeAvailability),
				TotalDuration:         interval.FromDuration(roadmap.PersonalizationOptions.TotalDuration),
				SkillLevel:            roadmap.PersonalizationOptions.SkillLevel.String(),
				AdditionalInfo:        roadmap.PersonalizationOptions.AdditionalInfo,
			},
		}
	}

	return output, nil
}
