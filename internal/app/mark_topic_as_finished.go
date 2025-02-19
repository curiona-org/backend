package app

import (
	"context"

	"github.com/roadmap-thesis/backend/internal/domain"
)

func (app *application) MarkTopicAsFinished(ctx context.Context, slug string) error {
	ctx, span := app.tracer.Start(ctx, "(*application.MarkTopicAsFinished)")
	defer span.End()

	err := app.repository.Topic.Update(ctx, slug, func(topic *domain.Topic) (bool, error) {
		if topic.Finished {
			return false, nil
		}

		topic.MarkAsFinished()
		return true, nil
	})
	if err != nil {
		return err
	}

	return nil
}
