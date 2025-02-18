package admin

import (
	"context"

	"github.com/roadmap-thesis/backend/internal/admin/io"
)

func (app *application) EditUser(ctx context.Context, input io.EditUserInput) (io.EditUserOutput, error) {
	_ = ctx
	_ = input
	return io.EditUserOutput{}, nil
}
