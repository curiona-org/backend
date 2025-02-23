package app

import (
	"context"
	"errors"

	"github.com/roadmap-thesis/backend/internal/app/io"
	"github.com/roadmap-thesis/backend/internal/cerrors"
	"github.com/roadmap-thesis/backend/internal/domain"
)

func (app *application) MarkTopicAsIncomplete(ctx context.Context, input io.MarkTopicInput) error {
	ctx, span := app.tracer.Start(ctx, "(*application.MarkTopicAsIncomplete)")
	defer span.End()

	err := app.repository.Topic.Update(ctx, input.Slug, func(topic *domain.Topic) (bool, error) {
		if topic.AccountID != input.AccountID {
			return false, cerrors.NotFound()
		}

		if !topic.Finished {
			return false, nil
		}

		topic.MarkAsIncomplete()
		return true, nil
	})
	if err != nil {
		if errors.Is(err, domain.ErrTopicNotFound) {
			return cerrors.Wrap(cerrors.NotFound(), err)
		}
		return err
	}

	return nil
}
