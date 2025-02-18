package application

import (
	"context"
	"errors"

	"github.com/roadmap-thesis/backend/internal/apperrors"
	"github.com/roadmap-thesis/backend/internal/application/io"
	"github.com/roadmap-thesis/backend/internal/domain"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (app *application) GetTopicBySlug(ctx context.Context, slug string) (io.GetTopicOutput, error) {
	ctx, span := app.tracer.Start(ctx, "(*application.GetTopicBySlug)", trace.WithAttributes(attribute.String("slug", slug)))
	defer span.End()

	topic, err := app.repository.Topic().GetBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, domain.ErrTopicNotFound) {
			return io.GetTopicOutput{}, apperrors.ResourceNotFound("topic")
		}
		return io.GetTopicOutput{}, err
	}

	return io.GetTopicOutput{
		ID:          topic.ID,
		RoadmapID:   topic.RoadmapID,
		ParentID:    topic.ParentID,
		Title:       topic.Title,
		Slug:        topic.Slug,
		Description: topic.Description,
		Order:       topic.Order,
		Finished:    topic.Finished,
		CreatedAt:   topic.CreatedAt,
		UpdatedAt:   topic.UpdatedAt,
	}, nil
}
