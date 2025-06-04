package app

import (
	"context"
	"errors"

	"github.com/curiona-org/backend/internal/app/io"
	"github.com/curiona-org/backend/internal/domain"
	"github.com/curiona-org/backend/internal/filter"
)

func (app *application) RateRoadmap(ctx context.Context, input io.RateRoadmapInput) error {
	ctx, span := app.tracer.Start(ctx, "(*application.RateRoadmap)")
	defer span.End()

	roadmap, err := app.repository.Roadmap.GetBySlug(ctx, filter.Filters{
		AccountID: input.AccountID,
		Slug:      input.Slug,
	})
	if err != nil {
		return err
	}

	progression, err := app.repository.Roadmap.GetRoadmapProgression(ctx, input.AccountID, roadmap.ID)
	if err != nil && !errors.Is(err, domain.ErrRoadmapProgressionNotFound) {
		return err
	}

	rating := domain.NewRating(
		roadmap.ID,
		input.AccountID,
		progression.TotalTopics,
		progression.TotalFinishedTopics,
		input.Rating,
		input.Comment,
	)

	if err := app.repository.Rating.RateRoadmap(ctx, rating); err != nil {
		return err
	}

	return nil
}
