package app

import (
	"context"
	"errors"

	"github.com/curiona-org/backend/internal/app/io"
	"github.com/curiona-org/backend/internal/cerrors"
	"github.com/curiona-org/backend/internal/domain"
)

func (app *application) DeleteUserRoadmap(ctx context.Context, input io.DeleteUserRoadmapInput) error {
	ctx, span := app.tracer.Start(ctx, "(*application.DeleteUserRoadmap)")
	defer span.End()

	err := app.repository.Roadmap.Update(ctx, input.Slug, func(roadmap *domain.Roadmap) (bool, error) {
		if roadmap.AccountID != input.AccountID {
			return false, cerrors.NotFound
		}

		roadmap.Delete()
		return true, nil
	})
	if err != nil {
		if errors.Is(err, domain.ErrRoadmapNotFound) {
			return cerrors.Wrap(cerrors.NotFound, err)
		}
		return err
	}

	return nil
}
