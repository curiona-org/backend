package app

import (
	"context"
	"errors"

	"github.com/curiona-org/backend/internal/app/io"
	"github.com/curiona-org/backend/internal/cerrors"
	"github.com/curiona-org/backend/internal/domain"
)

func (app *application) UnbookmarkRoadmap(ctx context.Context, input io.BookmarkRoadmapInput) error {
	ctx, span := app.tracer.Start(ctx, "(*application.UnbookmarkRoadmap)")
	defer span.End()

	err := app.repository.Bookmark.Delete(ctx, input.AccountID, input.Slug)
	if err != nil {
		if errors.Is(err, domain.ErrRoadmapNotFound) || errors.Is(err, domain.ErrBookmarkNotFound) {
			return cerrors.ErrNotFound.Msg("bookmark")
		}

		return err
	}

	return nil
}
