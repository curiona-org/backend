package admin

import (
	"context"

	"github.com/roadmap-thesis/backend/internal/admin/io"
)

func (app *application) ListRoadmaps(ctx context.Context) (io.ListRoadmapsOutput, error) {
	_ = ctx
	return io.ListRoadmapsOutput{}, nil
}
