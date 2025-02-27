package admin

import (
	"context"

	"github.com/curiona-org/backend/internal/admin/io"
)

func (app *adminApplication) EditUser(ctx context.Context, input io.EditUserInput) (io.EditUserOutput, error) {
	_ = ctx
	_ = input
	return io.EditUserOutput{}, nil
}
