package app

import (
	"context"
	"errors"

	"github.com/roadmap-thesis/backend/internal/app/io"
	"github.com/roadmap-thesis/backend/internal/apperrors"
	"github.com/roadmap-thesis/backend/internal/domain"
)

func (app *application) MarkTopicAsIncomplete(ctx context.Context, input io.MarkTopicInput) error {
	ctx, span := app.tracer.Start(ctx, "(*application.MarkTopicAsIncomplete)")
	defer span.End()

	err := app.repository.Topic.Update(ctx, input.Slug, func(topic *domain.Topic) (bool, error) {
		if topic.AccountID != input.AccountID {
			return false, apperrors.NotFound()
		}

		if !topic.Finished {
			return false, nil
		}

		topic.MarkAsIncomplete()
		return true, nil
	})
	if err != nil {
		if errors.Is(err, domain.ErrTopicNotFound) {
			return apperrors.Wrap(apperrors.NotFound(), err)
		}
		return err
	}

	return nil
}
