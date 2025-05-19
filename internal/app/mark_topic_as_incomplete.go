package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/curiona-org/backend/internal/app/io"
	"github.com/curiona-org/backend/internal/cerrors"
	"github.com/curiona-org/backend/internal/domain"
)

func (app *application) MarkTopicAsIncomplete(ctx context.Context, input io.MarkTopicInput) error {
	ctx, span := app.tracer.Start(ctx, "(*application.MarkTopicAsIncomplete)")
	defer span.End()

	err := app.repository.Topic.UpdateTopicStatus(ctx, input.AccountID, input.Slug, func(roadmap *domain.Roadmap, topic *domain.Topic) (bool, error) {
		if topic.AccountID != input.AccountID {
			return false, cerrors.ErrNotFound
		}

		if !topic.IsFinished {
			fmt.Println("Topic is already incomplete")
			return false, nil
		}

		topic.MarkAsIncomplete()
		return true, nil
	})
	if err != nil {
		if errors.Is(err, domain.ErrTopicNotFound) || errors.Is(err, domain.ErrRoadmapNotFound) {
			return cerrors.ErrNotFound.Msg("topic")
		}
		return err
	}

	return nil
}
