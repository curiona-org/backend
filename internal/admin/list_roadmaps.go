package admin

import (
	"context"

	"github.com/curiona-org/backend/internal/admin/io"
)

func (app *application) ListRoadmaps(ctx context.Context) (io.ListRoadmapsOutput, error) {
	_ = ctx
	return io.ListRoadmapsOutput{}, nil
}
