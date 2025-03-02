package worker

import "context"

type noopWorker struct{}

func NewNoop() Worker {
	return &noopWorker{}
}

func (w *noopWorker) Start(ctx context.Context) error {
	_ = ctx
	return nil
}

func (w *noopWorker) EnqueueSearchYoutubeExternalResources(
	ctx context.Context,
	payload SearchYoutubeExternalResourcesInput,
) error {
	_ = ctx
	_ = payload
	return nil
}

func (w *noopWorker) EnqueueSearchGoogleBooksExternalResources(
	ctx context.Context,
	payload SearchGoogleBooksExternalResourcesInput,
) error {
	_ = ctx
	_ = payload
	return nil
}
