package worker

import (
	"context"
)

type SearchYoutubeExternalResourcesInput struct {
	TopicID     int
	SearchQuery string
}

type SearchGoogleBooksExternalResourcesInput struct {
	TopicID     int
	SearchQuery string
}

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

var _ Worker = (*asynqWorker)(nil)
var _ Worker = (*noopWorker)(nil)
