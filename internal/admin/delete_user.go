package admin

import (
	"context"

	"github.com/curiona-org/backend/internal/admin/io"
)

func (app *application) DeleteUser(ctx context.Context, input io.DeleteUserInput) (io.DeleteUserOutput, error) {
	_ = ctx
	_ = input
	return io.DeleteUserOutput{}, nil
}
