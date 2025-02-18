package admin

import (
	"context"

	"github.com/roadmap-thesis/backend/internal/admin/io"
)

func (app *application) Statistics(ctx context.Context) (io.StatisticsOutput, error) {
	_ = ctx
	return io.StatisticsOutput{}, nil
}
