package worker

import (
	"context"

	"github.com/curiona-org/backend/internal/repository"
	"github.com/curiona-org/backend/pkg/googleapi/book"
	"github.com/curiona-org/backend/pkg/googleapi/youtube"
	"github.com/hibiken/asynq"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type Worker interface {
	// Start starts the worker and listens for incoming tasks.
	// Should be called in the main function.
	Start(ctx context.Context) error

	EnqueueSearchYoutubeExternalResources(
		ctx context.Context,
		payload SearchYoutubeExternalResourcesInput,
	) error
	EnqueueSearchGoogleBooksExternalResources(
		ctx context.Context,
		payload SearchGoogleBooksExternalResourcesInput,
	) error
}

var _ Worker = (*worker)(nil)

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

func (w *worker) Start(ctx context.Context) error {
	mux := asynq.NewServeMux()

	mux.HandleFunc(TaskSearchYoutubeExternalResources, w.searchYoutubeExternalResources)
	mux.HandleFunc(TaskSearchGoogleBooksExternalResources, w.searchGoogleBooksExternalResources)

	return w.srv.Run(mux)
}
