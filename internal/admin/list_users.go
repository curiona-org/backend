package admin

import (
	"context"

	"github.com/curiona-org/backend/internal/admin/io"
	"github.com/curiona-org/backend/pkg/pagination"
)

func (app *adminApplication) ListUsers(ctx context.Context, input io.ListUsersInput) (io.ListUsersOutput, error) {
	ctx, span := app.tracer.Start(ctx, "(*adminApplication.ListUsers)")
	defer span.End()

	count, err := app.repository.Account.Count(ctx)
	if err != nil {
		return io.ListUsersOutput{}, err
	}

	pagination := pagination.NewOffsetPaginator(input.Page, input.Limit, count)
	users, err := app.repository.Account.GetAll(ctx, pagination)
	if err != nil {
		return io.ListUsersOutput{}, err
	}

	output := io.ListUsersOutput{
		Total:       pagination.Total,
		TotalPages:  pagination.TotalPages,
		CurrentPage: pagination.CurrentPage,
		Items:       make([]io.ListUsersOutputItem, len(users)),
	}

	for idx, user := range users {
		output.Items[idx] = io.ListUsersOutputItem{
			ID:          user.ID,
			Method:      user.Method,
			Email:       user.Email,
			Name:        user.Profile.Name,
			Avatar:      user.Profile.Avatar,
			IsSuspended: user.IsSuspended,
			JoinedAt:    user.CreatedAt,
		}
	}

	return output, nil
}
