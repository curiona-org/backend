package app

import (
	"context"

	"github.com/roadmap-thesis/backend/internal/domain"
)

func (app *application) MarkTopicAsIncomplete(ctx context.Context, slug string) error {
	ctx, span := app.tracer.Start(ctx, "(*application.MarkTopicAsIncomplete)")
	defer span.End()

	err := app.repository.Topic.Update(ctx, slug, func(topic *domain.Topic) (bool, error) {
		if !topic.Finished {
			return false, nil
		}

		topic.MarkAsIncomplete()
		return true, nil
	})
	if err != nil {
		return err
	}

	return nil
}
