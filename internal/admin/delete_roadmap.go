package admin

import (
	"context"

	"github.com/curiona-org/backend/internal/admin/io"
)

func (app *application) DeleteRoadmap(ctx context.Context, input io.DeleteRoadmapInput) (io.DeleteRoadmapOutput, error) {
	_ = ctx
	_ = input
	return io.DeleteRoadmapOutput{}, nil
}
