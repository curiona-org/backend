package worker

import (
	"context"

	"github.com/curiona-org/backend/internal/domain"
	"github.com/curiona-org/backend/internal/logger"
	"github.com/curiona-org/backend/internal/repository"
	"github.com/curiona-org/backend/pkg/googleapi/book"
	"github.com/curiona-org/backend/pkg/googleapi/youtube"
	"github.com/hibiken/asynq"
	"github.com/vmihailenco/msgpack/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type asynqWorker struct {
	srv         *asynq.Server
	queue       *asynq.Client
	repository  *repository.Repository
	googleBooks book.Client
	youtube     youtube.Client
	tracer      trace.Tracer
}

func NewAsynq(
	queue *asynq.Client,
	srv *asynq.Server,
	repository *repository.Repository,
	googleBooks book.Client,
	youtube youtube.Client,
) Worker {
	tracer := otel.Tracer("worker")
	return &asynqWorker{
		queue:       queue,
		srv:         srv,
		repository:  repository,
		googleBooks: googleBooks,
		youtube:     youtube,
		tracer:      tracer,
	}
}

var (
	TaskSearchYoutubeExternalResources     = "queue:external_resources:youtube_search"
	TaskSearchGoogleBooksExternalResources = "queue:external_resources:google_books"
)

func (w *asynqWorker) Start(ctx context.Context) error {
	mux := asynq.NewServeMux()

	mux.HandleFunc(TaskSearchYoutubeExternalResources, w.searchYoutubeExternalResources)
	mux.HandleFunc(TaskSearchGoogleBooksExternalResources, w.searchGoogleBooksExternalResources)

	return w.srv.Run(mux)
}

func (w *asynqWorker) EnqueueSearchYoutubeExternalResources(ctx context.Context, payload SearchYoutubeExternalResourcesInput) error {
	ctx, span := w.tracer.Start(ctx, "(*asynqWorker.EnqueueSearchYoutubeExternalResources)")
	defer span.End()

	data, err := msgpack.Marshal(payload)
	if err != nil {
		return err
	}

	task := asynq.NewTask(TaskSearchYoutubeExternalResources, data, asynq.MaxRetry(3))

	info, err := w.queue.EnqueueContext(ctx, task)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	log := logger.FromContext(ctx)
	log.Info().
		Str("task", TaskSearchYoutubeExternalResources).
		Str("task_id", info.ID).
		Int("topic_id", payload.TopicID).
		Str("search_query", payload.SearchQuery).
		Msg("enqueued task")

	return nil
}

func (w *asynqWorker) searchYoutubeExternalResources(ctx context.Context, task *asynq.Task) error {
	traceCtx, span := w.tracer.Start(ctx, "(*asynqWorker.searchYoutubeExternalResources)")
	defer span.End()

	var payload SearchYoutubeExternalResourcesInput
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
			result.Channel,
			result.URL,
			result.Thumbnail,
			domain.ExternalResourceTypeYoutube,
		)

		externalResources = append(externalResources, externalResource)
	}

	if len(externalResources) == 0 {
		return nil
	}

	return w.repository.ExternalResource.BulkSave(traceCtx, payload.TopicID, externalResources)
}

func (w *asynqWorker) EnqueueSearchGoogleBooksExternalResources(ctx context.Context, payload SearchGoogleBooksExternalResourcesInput) error {
	ctx, span := w.tracer.Start(ctx, "(*asynqWorker.EnqueueSearchGoogleBooksExternalResources)")
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

func (w *asynqWorker) searchGoogleBooksExternalResources(ctx context.Context, task *asynq.Task) error {
	traceCtx, span := w.tracer.Start(ctx, "(*asynqWorker.searchGoogleBooksExternalResources)")
	defer span.End()

	var payload SearchGoogleBooksExternalResourcesInput
	if err := msgpack.Unmarshal(task.Payload(), &payload); err != nil {
		return err
	}

	searchResult, err := w.googleBooks.Search(traceCtx, payload.SearchQuery)
	if err != nil {
		return err
	}

	externalResources := make([]*domain.ExternalResource, 0)
	for _, result := range searchResult {
		externalResource := domain.NewExternalResource(
			payload.TopicID,
			result.Title,
			result.Authors,
			result.URL,
			result.Cover,
			domain.ExternalResourceTypeBook,
		)

		externalResources = append(externalResources, externalResource)
	}

	if len(externalResources) == 0 {
		return nil
	}

	return w.repository.ExternalResource.BulkSave(traceCtx, payload.TopicID, externalResources)
}
