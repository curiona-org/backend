package app

import (
	"context"
	"errors"

	"github.com/curiona-org/backend/internal/app/io"
	"github.com/curiona-org/backend/internal/cerrors"
	"github.com/curiona-org/backend/internal/domain"
)

func (app *application) MarkTopicAsFinished(ctx context.Context, input io.MarkTopicInput) error {
	ctx, span := app.tracer.Start(ctx, "(*application.MarkTopicAsFinished)")
	defer span.End()

	err := app.repository.Topic.Update(ctx, input.Slug, func(topic *domain.Topic) (bool, error) {
		if topic.AccountID != input.AccountID {
			return false, cerrors.NotFound
		}

		if topic.Finished {
			return false, nil
		}

		topic.MarkAsFinished()
		return true, nil
	})
	if err != nil {
		if errors.Is(err, domain.ErrTopicNotFound) {
			return cerrors.Wrap(cerrors.NotFound, err)
		}
		return err
	}

	return nil
}
