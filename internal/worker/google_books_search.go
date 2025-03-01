package worker

import (
	"context"

	"github.com/curiona-org/backend/internal/domain"
	"github.com/curiona-org/backend/internal/logger"
	"github.com/hibiken/asynq"
	"github.com/vmihailenco/msgpack/v5"
	"go.opentelemetry.io/otel/codes"
)

type SearchGoogleBooksExternalResourcesInput struct {
	TopicID     int
	SearchQuery string
}

func (w *worker) EnqueueSearchGoogleBooksExternalResources(ctx context.Context, payload SearchGoogleBooksExternalResourcesInput) error {
	ctx, span := w.tracer.Start(ctx, "(*worker.EnqueueSearchGoogleBooksExternalResources)")
	defer span.End()

	data, err := msgpack.Marshal(payload)
	if err != nil {
		return err
	}

	task := asynq.NewTask(TaskSearchGoogleBooksExternalResources, data, asynq.MaxRetry(3))

	info, err := w.queue.EnqueueContext(ctx, task)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	log := logger.FromContext(ctx)
	log.Info().
		Str("task", TaskSearchGoogleBooksExternalResources).
		Str("task_id", info.ID).
		Int("topic_id", payload.TopicID).
		Str("search_query", payload.SearchQuery).
		Msg("enqueued task")

	return nil
}

func (w *worker) searchGoogleBooksExternalResources(ctx context.Context, task *asynq.Task) error {
	traceCtx, span := w.tracer.Start(ctx, "(*worker.searchGoogleBooksExternalResources)")
	defer span.End()

	var payload SearchGoogleBooksExternalResourcesInput
	if err := msgpack.Unmarshal(task.Payload(), &payload); err != nil {
		return err
	}

	searchResult, err := w.youtube.Search(traceCtx, payload.SearchQuery)
	if err != nil {
		return err
	}

	externalResources := make([]*domain.ExternalResource, 0)
	for _, result := range searchResult {
		externalResource := domain.NewExternalResource(
			payload.TopicID,
			result.Title,
			result.URL,
			domain.ExternalResourceTypeYoutube,
		)

		externalResources = append(externalResources, externalResource)
	}

	if len(externalResources) == 0 {
		return nil
	}

	return w.repository.ExternalResource.BulkSave(traceCtx, payload.TopicID, externalResources)
}
