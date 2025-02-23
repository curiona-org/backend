package worker

import (
	"context"

	"github.com/hibiken/asynq"
	"github.com/roadmap-thesis/backend/internal/googleapi/book"
	"github.com/roadmap-thesis/backend/internal/googleapi/youtube"
	"github.com/roadmap-thesis/backend/internal/repository"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type Worker interface {
	Start() error

	EnqueueSearchYoutubeExternalResources(
		ctx context.Context,
		payload SearchYoutubeExternalResourcesInput,
	) error
	EnqueueSearchGoogleBooksExternalResources(
		ctx context.Context,
		payload SearchGoogleBooksExternalResourcesInput,
	) error

	searchYoutubeExternalResources(ctx context.Context, task *asynq.Task) error
	searchGoogleBooksExternalResources(ctx context.Context, task *asynq.Task) error
}

type worker struct {
	srv         *asynq.Server
	queue       *asynq.Client
	repository  *repository.Repository
	googleBooks book.Client
	youtube     youtube.Client
	tracer      trace.Tracer
}

func New(
	queue *asynq.Client,
	srv *asynq.Server,
	repository *repository.Repository,
	googleBooks book.Client,
	youtube youtube.Client,
) Worker {
	tracer := otel.Tracer("worker")
	return &worker{
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

func (w *worker) Start() error {
	mux := asynq.NewServeMux()

	mux.HandleFunc(TaskSearchYoutubeExternalResources, w.searchYoutubeExternalResources)
	mux.HandleFunc(TaskSearchGoogleBooksExternalResources, w.searchGoogleBooksExternalResources)

	return w.srv.Start(mux)
}
