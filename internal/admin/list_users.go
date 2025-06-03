package admin

import (
	"context"

	"github.com/curiona-org/backend/internal/admin/io"
	"github.com/curiona-org/backend/internal/filter"
)

func (app *adminApplication) ListUsers(ctx context.Context, input io.ListUsersInput) (io.ListUsersOutput, error) {
	ctx, span := app.tracer.Start(ctx, "(*adminApplication.ListUsers)")
	defer span.End()

	var count uint64
	var err error
	if input.Search != "" {
		count, err = app.repository.Account.CountBySearching(ctx, input.Search)
		if err != nil {
			return io.ListUsersOutput{}, err
		}
	} else {
		count, err = app.repository.Account.Count(ctx)
		if err != nil {
			return io.ListUsersOutput{}, err
		}
	}

	filters := filter.New(input, count)

	filters.Options = map[string]any{
		"admin.with_total_roadmaps": true,
	}

	users, err := app.repository.Account.ListAll(ctx, filters)
	if err != nil {
		return io.ListUsersOutput{}, err
	}

	output := io.ListUsersOutput{
		Total:       filters.Paginator.Total,
		TotalPages:  filters.Paginator.TotalPages,
		CurrentPage: filters.Paginator.CurrentPage,
		Items:       make([]io.ListUsersOutputItem, len(users)),
	}

	for idx, user := range users {
		output.Items[idx] = io.ListUsersOutputItem{
			ID:            user.ID,
			Method:        user.Method,
			Email:         user.Email,
			Name:          user.Profile.Name,
			Avatar:        user.Profile.Avatar,
			TotalRoadmaps: user.TotalRoadmaps,
			IsSuspended:   user.IsSuspended,
			IsAdmin:       user.IsAdmin,
			JoinedAt:      user.CreatedAt,
		}
	}

	return output, nil
}
