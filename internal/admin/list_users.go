package admin

import (
	"context"

	"github.com/roadmap-thesis/backend/internal/admin/io"
)

func (app *application) ListUsers(ctx context.Context) (io.ListUsersOutput, error) {
	_ = ctx
	return io.ListUsersOutput{}, nil
}
