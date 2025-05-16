package app

import (
	"context"
	"errors"

	"github.com/curiona-org/backend/internal/app/io"
	"github.com/curiona-org/backend/internal/cerrors"
	"github.com/curiona-org/backend/internal/domain"
)

func (app *application) BookmarkRoadmap(ctx context.Context, input io.BookmarkRoadmapInput) error {
	ctx, span := app.tracer.Start(ctx, "(*application.BookmarkRoadmap)")
	defer span.End()

	err := app.repository.Bookmark.Save(ctx, input.AccountID, input.Slug)
	if err != nil {
		if errors.Is(err, domain.ErrRoadmapNotFound) {
			return cerrors.ErrNotFound.Msg("roadmap")
		}

		return err
	}

	return nil
}
